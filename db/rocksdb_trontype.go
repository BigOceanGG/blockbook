package db

import (
	"bytes"
	"encoding/hex"
	"github.com/trezor/blockbook/bchain/coins/trx"

	vlq "github.com/bsm/go-vlq"
	"github.com/golang/glog"
	"github.com/juju/errors"
	"github.com/tecbot/gorocksdb"
	"github.com/trezor/blockbook/bchain"
)

type tronBlockTx struct {
	btxID     []byte
	from, to  bchain.AddressDescriptor
	contracts []ethBlockTxContract
}

func (d *RocksDB) GetTronAddrDescContracts(addrDesc bchain.AddressDescriptor) (*AddrContracts, error) {
	val, err := d.db.GetCF(d.ro, d.cfh[cfAddressContracts], addrDesc)
	if err != nil {
		return nil, err
	}
	defer val.Free()
	buf := val.Data()
	if len(buf) == 0 {
		return nil, nil
	}
	tt, l := unpackVaruint(buf)
	buf = buf[l:]
	nct, l := unpackVaruint(buf)
	buf = buf[l:]
	c := make([]AddrContract, 0, 4)
	for len(buf) > 0 {
		if len(buf) < trx.TronTypeAddressDescriptorLen {
			return nil, errors.New("Invalid data stored in cfAddressContracts for AddrDesc " + addrDesc.String())
		}
		txs, l := unpackVaruint(buf[trx.TronTypeAddressDescriptorLen:])
		contract := append(bchain.AddressDescriptor(nil), buf[:trx.TronTypeAddressDescriptorLen]...)
		c = append(c, AddrContract{
			Contract: contract,
			Txs:      txs,
		})
		buf = buf[trx.TronTypeAddressDescriptorLen+l:]
	}
	return &AddrContracts{
		TotalTxs:       tt,
		NonContractTxs: nct,
		Contracts:      c,
	}, nil
}

func (d *RocksDB) addToContractsTronType(addrDesc bchain.AddressDescriptor, btxID []byte, contract bchain.AddressDescriptor, addresses addressesMap, addressContracts map[string]*AddrContracts) error {
	var err error
	strAddrDesc := string(addrDesc)
	ac, e := addressContracts[strAddrDesc]
	if !e {
		ac, err = d.GetTronAddrDescContracts(addrDesc)
		if err != nil {
			return err
		}
		if ac == nil {
			ac = &AddrContracts{}
		}
		addressContracts[strAddrDesc] = ac
		d.cbs.balancesMiss++
	} else {
		d.cbs.balancesHit++
	}
	if contract != nil {
		// do not store contracts for 0x0000000000000000000000000000000000000000 address
		if !isZeroAddress(addrDesc) {
			// locate the contract and set i to the index in the array of contracts
			_, found := findContractInAddressContracts(contract, ac.Contracts)
			if !found {
				ac.Contracts = append(ac.Contracts, AddrContract{Contract: contract})
			}
		}
	}
	at, found := addresses[strAddrDesc]
	if !found {
		addresses[strAddrDesc] = append(at, txIndexes{
			btxID: btxID,
		})
	}
	return nil
}

func (d *RocksDB) processAddressesAndContractsTronType(block *bchain.Block, addresses addressesMap, addressContracts map[string]*AddrContracts) ([]tronBlockTx, error) {
	var blockTxs []tronBlockTx
	for _, tx := range block.Txs {
		btxID, err := d.chainParser.PackTxid(tx.Txid)
		if err != nil {
			return nil, err
		}
		var blockTx tronBlockTx
		blockTx.btxID = btxID
		var from, to bchain.AddressDescriptor
		// there is only one output address in EthereumType transaction, store it in format txid 0
		if len(tx.Vout) == 1 && len(tx.Vout[0].ScriptPubKey.Addresses) == 1 {
			to, err = d.chainParser.GetAddrDescFromAddress(tx.Vout[0].ScriptPubKey.Addresses[0])
			if err != nil {
				// do not log ErrAddressMissing, transactions can be without to address (for example eth contracts)
				if err != bchain.ErrAddressMissing {
					glog.Warningf("rocksdb: addrDesc: %v - height %d, tx %v, output", err, block.Height, tx.Txid)
				}
				continue
			}
			blockTx.to = to
		}
		// there is only one input address in EthereumType transaction, store it in format txid ^0
		if len(tx.Vin) == 1 && len(tx.Vin[0].Addresses) == 1 {
			from, err = d.chainParser.GetAddrDescFromAddress(tx.Vin[0].Addresses[0])
			if err != nil {
				if err != bchain.ErrAddressMissing {
					glog.Warningf("rocksdb: addrDesc: %v - height %d, tx %v, input", err, block.Height, tx.Txid)
				}
				continue
			}
			blockTx.from = from
		}
		// store erc20 transfers
		trc20, err := d.chainParser.TronTypeGetTrc20FromTx(&tx)
		if err != nil {
			glog.Warningf("rocksdb: GetErc20FromTx %v - height %d, tx %v", err, block.Height, tx.Txid)
		}
		for _, t := range trc20 {
			var contract, from, to bchain.AddressDescriptor
			contract, err = d.chainParser.GetAddrDescFromAddress(t.Contract)
			if err == nil {
				from, err = d.chainParser.GetAddrDescFromAddress(t.From)
				if err == nil {
					to, err = d.chainParser.GetAddrDescFromAddress(t.To)
				}
			}
			if err != nil {
				glog.Warningf("rocksdb: GetTrc20FromTx %v - height %d, tx %v, transfer %v", err, block.Height, tx.Txid, t)
				continue
			}
			if err = d.addToContractsTronType(to, btxID, contract, addresses, addressContracts); err != nil {
				return nil, err
			}
			blockTx.contracts = append(blockTx.contracts, ethBlockTxContract{
				from,
				contract,
			})
		}
		blockTxs = append(blockTxs, blockTx)
	}

	return blockTxs, nil
}

func (d *RocksDB) storeAndCleanupBlockTxsTronType(wb *gorocksdb.WriteBatch, block *bchain.Block, blockTxs []tronBlockTx) error {
	pl := d.chainParser.PackedTxidLen()
	buf := make([]byte, 0, (pl+2*trx.TronTypeAddressDescriptorLen)*len(blockTxs))
	varBuf := make([]byte, vlq.MaxLen64)
	zeroAddress := make([]byte, trx.TronTypeAddressDescriptorLen)
	appendAddress := func(a bchain.AddressDescriptor) {
		if len(a) != trx.TronTypeAddressDescriptorLen {
			buf = append(buf, zeroAddress...)
		} else {
			buf = append(buf, a...)
		}
	}
	for i := range blockTxs {
		blockTx := &blockTxs[i]
		buf = append(buf, blockTx.btxID...)
		appendAddress(blockTx.from)
		appendAddress(blockTx.to)
		l := packVaruint(uint(len(blockTx.contracts)), varBuf)
		buf = append(buf, varBuf[:l]...)
		for j := range blockTx.contracts {
			c := &blockTx.contracts[j]
			appendAddress(c.addr)
			appendAddress(c.contract)
		}
	}
	key := packUint(block.Height)
	wb.PutCF(d.cfh[cfBlockTxs], key, buf)
	return d.cleanupBlockTxs(wb, block)
}

func (d *RocksDB) getBlockTxsTronType(height uint32) ([]tronBlockTx, error) {
	pl := d.chainParser.PackedTxidLen()
	val, err := d.db.GetCF(d.ro, d.cfh[cfBlockTxs], packUint(height))
	if err != nil {
		return nil, err
	}
	defer val.Free()
	buf := val.Data()
	// nil data means the key was not found in DB
	if buf == nil {
		return nil, nil
	}
	// buf can be empty slice, this means the block did not contain any transactions
	bt := make([]tronBlockTx, 0, 8)
	getAddress := func(i int) (bchain.AddressDescriptor, int, error) {
		if len(buf)-i < trx.TronTypeAddressDescriptorLen {
			glog.Error("rocksdb: Inconsistent data in blockTxs ", hex.EncodeToString(buf))
			return nil, 0, errors.New("Inconsistent data in blockTxs")
		}
		a := append(bchain.AddressDescriptor(nil), buf[i:i+trx.TronTypeAddressDescriptorLen]...)
		// return null addresses as nil
		for _, b := range a {
			if b != 0 {
				return a, i + trx.TronTypeAddressDescriptorLen, nil
			}
		}
		return nil, i + trx.TronTypeAddressDescriptorLen, nil
	}
	var from, to bchain.AddressDescriptor
	for i := 0; i < len(buf); {
		if len(buf)-i < pl {
			glog.Error("rocksdb: Inconsistent data in blockTxs ", hex.EncodeToString(buf))
			return nil, errors.New("Inconsistent data in blockTxs")
		}
		txid := append([]byte(nil), buf[i:i+pl]...)
		i += pl
		from, i, err = getAddress(i)
		if err != nil {
			return nil, err
		}
		to, i, err = getAddress(i)
		if err != nil {
			return nil, err
		}
		cc, l := unpackVaruint(buf[i:])
		i += l
		contracts := make([]ethBlockTxContract, cc)
		for j := range contracts {
			contracts[j].addr, i, err = getAddress(i)
			if err != nil {
				return nil, err
			}
			contracts[j].contract, i, err = getAddress(i)
			if err != nil {
				return nil, err
			}
		}
		bt = append(bt, tronBlockTx{
			btxID:     txid,
			from:      from,
			to:        to,
			contracts: contracts,
		})
	}
	return bt, nil
}

func (d *RocksDB) disconnectBlockTxsTronType(wb *gorocksdb.WriteBatch, height uint32, blockTxs []tronBlockTx, contracts map[string]*AddrContracts) error {
	glog.Info("Disconnecting block ", height, " containing ", len(blockTxs), " transactions")
	addresses := make(map[string]map[string]struct{})
	disconnectAddress := func(btxID []byte, addrDesc, contract bchain.AddressDescriptor) error {
		var err error
		// do not process empty address
		if len(addrDesc) == 0 {
			return nil
		}
		s := string(addrDesc)
		txid := string(btxID)
		// find if tx for this address was already encountered
		mtx, ftx := addresses[s]
		if !ftx {
			mtx = make(map[string]struct{})
			mtx[txid] = struct{}{}
			addresses[s] = mtx
		} else {
			_, ftx = mtx[txid]
			if !ftx {
				mtx[txid] = struct{}{}
			}
		}
		c, fc := contracts[s]
		if !fc {
			c, err = d.GetTronAddrDescContracts(addrDesc)
			if err != nil {
				return err
			}
			contracts[s] = c
		}
		if c != nil {
			if !ftx {
				c.TotalTxs--
			}
			if contract == nil {
				if c.NonContractTxs > 0 {
					c.NonContractTxs--
				} else {
					glog.Warning("AddressContracts ", addrDesc, ", EthTxs would be negative, tx ", hex.EncodeToString(btxID))
				}
			} else {
				i, found := findContractInAddressContracts(contract, c.Contracts)
				if found {
					if c.Contracts[i].Txs > 0 {
						c.Contracts[i].Txs--
						if c.Contracts[i].Txs == 0 {
							c.Contracts = append(c.Contracts[:i], c.Contracts[i+1:]...)
						}
					} else {
						glog.Warning("AddressContracts ", addrDesc, ", contract ", i, " Txs would be negative, tx ", hex.EncodeToString(btxID))
					}
				} else {
					glog.Warning("AddressContracts ", addrDesc, ", contract ", contract, " not found, tx ", hex.EncodeToString(btxID))
				}
			}
		} else {
			glog.Warning("AddressContracts ", addrDesc, " not found, tx ", hex.EncodeToString(btxID))
		}
		return nil
	}
	for i := range blockTxs {
		blockTx := &blockTxs[i]
		if err := disconnectAddress(blockTx.btxID, blockTx.from, nil); err != nil {
			return err
		}
		// if from==to, tx is counted only once and does not have to be disconnected again
		if !bytes.Equal(blockTx.from, blockTx.to) {
			if err := disconnectAddress(blockTx.btxID, blockTx.to, nil); err != nil {
				return err
			}
		}
		for _, c := range blockTx.contracts {
			if err := disconnectAddress(blockTx.btxID, c.addr, c.contract); err != nil {
				return err
			}
		}
		wb.DeleteCF(d.cfh[cfTransactions], blockTx.btxID)
	}
	for a := range addresses {
		key := packAddressKey([]byte(a), height)
		wb.DeleteCF(d.cfh[cfAddresses], key)
	}
	return nil
}

// DisconnectBlockRangeEthereumType removes all data belonging to blocks in range lower-higher
// it is able to disconnect only blocks for which there are data in the blockTxs column
func (d *RocksDB) DisconnectBlockRangeTronType(lower uint32, higher uint32) error {
	blocks := make([][]tronBlockTx, higher-lower+1)
	for height := lower; height <= higher; height++ {
		blockTxs, err := d.getBlockTxsTronType(height)
		if err != nil {
			return err
		}
		// nil blockTxs means blockTxs were not found in db
		if blockTxs == nil {
			return errors.Errorf("Cannot disconnect blocks with height %v and lower. It is necessary to rebuild index.", height)
		}
		blocks[height-lower] = blockTxs
	}
	wb := gorocksdb.NewWriteBatch()
	defer wb.Destroy()
	contracts := make(map[string]*AddrContracts)
	for height := higher; height >= lower; height-- {
		if err := d.disconnectBlockTxsTronType(wb, height, blocks[height-lower], contracts); err != nil {
			return err
		}
		key := packUint(height)
		wb.DeleteCF(d.cfh[cfBlockTxs], key)
		wb.DeleteCF(d.cfh[cfHeight], key)
	}
	d.storeAddressContracts(wb, contracts)
	err := d.db.Write(d.wo, wb)
	if err == nil {
		d.is.RemoveLastBlockTimes(int(higher-lower) + 1)
		glog.Infof("rocksdb: blocks %d-%d disconnected", lower, higher)
	}
	return err
}
