package bchain

import (
	"github.com/golang/glog"
	"time"
)

type MempoolTronType struct {
	BaseMempool
	mempoolTimeoutTime   time.Duration
	queryBackendOnResync bool
	nextTimeoutRun       time.Time
}

func NewMempoolTronType(chain BlockChain, mempoolTxTimeoutHours int, queryBackendOnResync bool) *MempoolTronType {
	mempoolTimeoutTime := time.Duration(mempoolTxTimeoutHours) * time.Hour
	return &MempoolTronType{
		BaseMempool: BaseMempool{
			chain:        chain,
			txEntries:    make(map[string]txEntry),
			addrDescToTx: make(map[string][]Outpoint),
		},
		mempoolTimeoutTime:   mempoolTimeoutTime,
		queryBackendOnResync: queryBackendOnResync,
		nextTimeoutRun:       time.Now().Add(mempoolTimeoutTime),
	}
}

func (m *MempoolTronType) Notify(tx *Tx, txid string, height uint32) {
	mtx := m.txToMempoolTx(tx)
	mtx.Blockheight = height
	parser := m.chain.GetChainParser()
	addrIndexes := make([]addrIndex, 0, len(mtx.Vout)+len(mtx.Vin))
	for _, output := range mtx.Vout {
		addrDesc, err := parser.GetAddrDescFromVout(&output)
		if err != nil {
			if err != ErrAddressMissing {
				glog.Error("error in output addrDesc in ", txid, " ", output.N, ": ", err)
			}
			continue
		}
		if len(addrDesc) > 0 {
			addrIndexes = append(addrIndexes, addrIndex{string(addrDesc), int32(output.N)})
		}
	}
	for j := range mtx.Vin {
		input := &mtx.Vin[j]
		for i, a := range input.Addresses {
			addrIndexes, input.AddrDesc = appendAddress(addrIndexes, ^int32(i), a, parser)
		}
	}
	t, err := parser.TronTypeGetTrc20FromTx(tx)
	if err != nil {
		glog.Error("GetTrc20FromTx for tx ", txid, ", ", err)
	} else {
		mtx.Trc20 = t
		for i := range t {
			addrIndexes, _ = appendAddress(addrIndexes, ^int32(i+1), t[i].From, parser)
			addrIndexes, _ = appendAddress(addrIndexes, int32(i+1), t[i].To, parser)
		}
	}
	if m.OnNewTxAddr != nil {
		if m.chain.TronTypeGetTransactionNotify(tx) {
			sent := make(map[string]struct{})
			for _, si := range addrIndexes {
				if _, found := sent[si.addrDesc]; !found {
					m.OnNewTxAddr(tx, AddressDescriptor(si.addrDesc))
					sent[si.addrDesc] = struct{}{}
				}
			}
		}
	}
	if m.OnNewTx != nil {
		if m.chain.TronTypeGetTransactionNotify(tx) {
			m.OnNewTx(mtx)
		}
	}
}

func (m *MempoolTronType) Resync() (int, error) {
	return 0, nil
}
