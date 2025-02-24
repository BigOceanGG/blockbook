package db

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/juju/errors"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/eth"
	"github.com/trezor/blockbook/bchain/coins/trx"
	"github.com/trezor/blockbook/common"
)

// TxCache is handle to TxCacheServer
type TxCache struct {
	db        *RocksDB
	chain     bchain.BlockChain
	metrics   *common.Metrics
	is        *common.InternalState
	enabled   bool
	chainType bchain.ChainType
}

// NewTxCache creates new TxCache interface and returns its handle
func NewTxCache(db *RocksDB, chain bchain.BlockChain, metrics *common.Metrics, is *common.InternalState, enabled bool) (*TxCache, error) {
	if !enabled {
		glog.Info("txcache: disabled")
	}
	return &TxCache{
		db:        db,
		chain:     chain,
		metrics:   metrics,
		is:        is,
		enabled:   enabled,
		chainType: chain.GetChainParser().GetChainType(),
	}, nil
}

// GetTransaction returns transaction either from RocksDB or if not present from blockchain
// it the transaction is confirmed, it is stored in the RocksDB
func (c *TxCache) GetTransaction(txid string) (*bchain.Tx, int, error) {
	var tx *bchain.Tx
	var h uint32
	var err error
	if c.enabled {
		tx, h, err = c.db.GetTx(txid)
		if err != nil {
			return nil, 0, err
		}
		if tx != nil {
			// number of confirmations is not stored in cache, they change all the time
			_, bestheight, _ := c.is.GetSyncState()
			tx.Confirmations = bestheight - h + 1
			c.metrics.TxCacheEfficiency.With(common.Labels{"status": "hit"}).Inc()
			glog.Info("GetTransaction by cache : ", txid)
			return tx, int(h), nil
		}
	}
	glog.Info("GetTransaction by chain : ", txid)
	tx, err = c.chain.GetTransaction(txid)
	if err != nil {
		return nil, 0, err
	}
	if c.chainType == bchain.ChainTronType {
		tx.Confirmations = c.is.BestHeight - h + 1
	}
	c.metrics.TxCacheEfficiency.With(common.Labels{"status": "miss"}).Inc()
	// cache only confirmed transactions
	if tx.Confirmations > 0 {
		if c.chainType == bchain.ChainBitcoinType {
			ta, err := c.db.GetTxAddresses(txid)
			if err != nil {
				return nil, 0, err
			}
			switch {
			case ta == nil:
				// the transaction may not yet be indexed, in that case:
				if tx.BlockHeight > 0 {
					// Check if the tx height value is set.
					h = tx.BlockHeight
				} else {
					// Get the height from the backend's bestblock.
					h, err = c.chain.GetBestBlockHeight()
					if err != nil {
						return nil, 0, err
					}
				}
			default:
				h = ta.Height
			}
		} else if c.chainType == bchain.ChainEthereumType {
			h, err = eth.GetHeightFromTx(tx)
			if err != nil {
				return nil, 0, err
			}
		} else if c.chainType == bchain.ChainTronType {
			h, err = trx.GetHeightFromTx(tx)
			if err != nil {
				return nil, 0, err
			}
			tx.BlockHeight = h
		} else {
			return nil, 0, errors.New("Unknown chain type")
		}
		if c.enabled {
			err = c.db.PutTx(tx, h, tx.Blocktime)
			// do not return caching error, only log it
			if err != nil {
				glog.Warning("PutTx ", tx.Txid, ",error ", err)
			}
		}
	} else {
		return tx, -1, nil
	}
	return tx, int(h), nil
}

func (c *TxCache) GetTransactionSpecific(txid string) (json.RawMessage, error) {
	var tx *bchain.Tx
	var h uint32
	var err error
	if c.enabled {
		tx, h, err = c.db.GetTx(txid)
		if err != nil {
			return nil, err
		}
		if tx != nil {
			_, bestheight, _ := c.is.GetSyncState()
			tx.Confirmations = bestheight - h + 1
			c.metrics.TxCacheEfficiency.With(common.Labels{"status": "hit"}).Inc()
			glog.Info("GetTransactionSpecific by cache : ", txid)
			return c.chain.GetTransactionSpecific(tx)
		}
	}
	glog.Info("GetTransactionSpecific by chain : ", txid)
	tx, err = c.chain.GetTransaction(txid)
	if err != nil {
		return nil, err
	}
	if c.chainType == bchain.ChainTronType {
		tx.Confirmations = c.is.BestHeight - h + 1
	}
	c.metrics.TxCacheEfficiency.With(common.Labels{"status": "miss"}).Inc()
	// cache only confirmed transactions
	if tx.Confirmations > 0 {
		if c.chainType == bchain.ChainBitcoinType {
			ta, err := c.db.GetTxAddresses(txid)
			if err != nil {
				return nil, err
			}
			switch {
			case ta == nil:
				// the transaction may not yet be indexed, in that case:
				if tx.BlockHeight > 0 {
					// Check if the tx height value is set.
					h = tx.BlockHeight
				} else {
					// Get the height from the backend's bestblock.
					h, err = c.chain.GetBestBlockHeight()
					if err != nil {
						return nil, err
					}
				}
			default:
				h = ta.Height
			}
		} else if c.chainType == bchain.ChainEthereumType {
			h, err = eth.GetHeightFromTx(tx)
			if err != nil {
				return nil, err
			}
		} else if c.chainType == bchain.ChainTronType {
			h, err = trx.GetHeightFromTx(tx)
			if err != nil {
				return nil, err
			}
			tx.BlockHeight = h
		} else {
			return nil, errors.New("Unknown chain type")
		}
		if c.enabled {
			err = c.db.PutTx(tx, h, tx.Blocktime)
			// do not return caching error, only log it
			if err != nil {
				glog.Warning("PutTx ", tx.Txid, ",error ", err)
			}
		}
	}
	return c.chain.GetTransactionSpecific(tx)
}
