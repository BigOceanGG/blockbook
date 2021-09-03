package trx

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/glog"
	"github.com/juju/errors"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/common"
	"google.golang.org/grpc"
	"math/big"
	"strconv"
	"time"
)

type Configuration struct {
	CoinName             string `json:"coin_name"`
	CoinShortcut         string `json:"coin_shortcut"`
	RPCURL               string `json:"rpc_url"`
	RPCTimeout           int    `json:"rpc_timeout"`
	MessageQueueBinding  string `json:"message_queue_binding"`
	MempoolWorkers       int    `json:"mempool_workers"`
	MempoolSubWorkers    int    `json:"mempool_sub_workers"`
	BlockAddressesToKeep int    `json:"block_addresses_to_keep"`

	MempoolTxTimeoutHours       int  `json:"mempoolTxTimeoutHours"`
	QueryBackendOnMempoolResync bool `json:"queryBackendOnMempoolResync"`
}

type TrxRPC struct {
	*bchain.BaseChain
	conn        *client.GrpcClient
	pushHandler func(bchain.NotificationType)
	mq          *bchain.MQ
	Mempool     *bchain.MempoolEthereumType
	ChainConfig *Configuration
	Parser      *TrxParser
}

func NewTrxRPC(config json.RawMessage, pushHandler func(bchain.NotificationType)) (bchain.BlockChain, error) {
	var err error
	var c Configuration
	err = json.Unmarshal(config, &c)
	if err != nil {
		return nil, errors.Annotatef(err, "Invalid configuration file")
	}

	conn := client.NewGrpcClientWithTimeout(c.RPCURL, 20*time.Second)
	if err := conn.Start([]grpc.DialOption{grpc.WithInsecure()}...); err != nil {
		return nil, err
	}

	s := &TrxRPC{
		BaseChain:   &bchain.BaseChain{},
		conn:        conn,
		ChainConfig: &c,
		pushHandler: pushHandler,
	}

	s.Parser = NewTrxParser(c.BlockAddressesToKeep, s)

	return s, nil
}

// CreateMempool creates mempool if not already created, however does not initialize it
func (b *TrxRPC) CreateMempool(chain bchain.BlockChain) (bchain.Mempool, error) {
	if b.Mempool == nil {
		b.Mempool = bchain.NewMempoolEthereumType(chain, b.ChainConfig.MempoolTxTimeoutHours, b.ChainConfig.QueryBackendOnMempoolResync)
	}
	return b.Mempool, nil
}

// InitializeMempool creates ZeroMQ subscription and sets AddrDescForOutpointFunc to the Mempool
func (b *TrxRPC) InitializeMempool(addrDescForOutpoint bchain.AddrDescForOutpointFunc, onNewTxAddr bchain.OnNewTxAddrFunc, onNewTx bchain.OnNewTxFunc) error {
	if b.Mempool == nil {
		return errors.New("Mempool not created")
	}
	b.Mempool.OnNewTxAddr = onNewTxAddr
	b.Mempool.OnNewTx = onNewTx
	if b.mq == nil {
		mq, err := bchain.NewMQ(b.ChainConfig.MessageQueueBinding, b.pushHandler)
		if err != nil {
			glog.Error("mq: ", err)
			return err
		}
		b.mq = mq
	}
	return nil
}

// EstimateFee returns fee estimation.
func (b *TrxRPC) EstimateFee(blocks int) (big.Int, error) {
	// use EstimateSmartFee if EstimateFee is not supported
	return big.Int{}, nil
}

func (b *TrxRPC) EstimateSmartFee(blocks int, conservative bool) (big.Int, error) {
	return big.Int{}, nil
}

func (b *TrxRPC) GetBestBlockHash() (string, error) {
	block, err := b.conn.GetNowBlock()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(block.Blockid), nil
}

func (b *TrxRPC) GetBestBlockHeight() (uint32, error) {
	block, err := b.conn.GetNowBlock()
	if err != nil {
		return 0, err
	}
	return uint32(block.BlockHeader.RawData.Number), nil
}

func (b *TrxRPC) GetBlock(hash string, height uint32) (*bchain.Block, error) {
	if hash == "" {
		h, err := b.GetBlockHash(height)
		if err != nil {
			return nil, err
		}
		hash = h
	}

	block, err := b.conn.GetBlockByID(hash)
	if err != nil {
		return nil, err
	}

	confirmations, err := b.computeConfirmations(block.BlockHeader.RawData.Number)
	if err != nil {
		return nil, err
	}
	var cblock bchain.Block
	cblock.BlockHeader = bchain.BlockHeader{
		Hash:          hash,
		Prev:          hex.EncodeToString(block.BlockHeader.RawData.ParentHash),
		Height:        uint32(block.BlockHeader.RawData.Number),
		Confirmations: confirmations,
		Time:          block.BlockHeader.RawData.Timestamp,
	}

	blockExtention, err := b.conn.GetBlockByNum(block.BlockHeader.RawData.Number)
	if err != nil {
		return nil, err
	}
	for _, tx := range blockExtention.Transactions {
		btx, err := b.Parser.trxtotx(tx, block.BlockHeader.RawData.Timestamp, uint32(confirmations))
		if err == nil {
			cblock.Txs = append(cblock.Txs, btx)
		}
	}

	return &cblock, nil
}

func (b *TrxRPC) GetBlockHash(height uint32) (string, error) {
	block, err := b.conn.GetBlockByNum(int64(height))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(block.Blockid), nil
}

func (b *TrxRPC) GetBlockHeader(hash string) (*bchain.BlockHeader, error) {
	block, err := b.conn.GetBlockByID(hash)
	if err != nil {
		return nil, err
	}

	confirmations, err := b.computeConfirmations(block.BlockHeader.RawData.Number)
	if err != nil {
		return nil, err
	}
	return &bchain.BlockHeader{
		Hash:          hash,
		Prev:          hex.EncodeToString(block.BlockHeader.RawData.ParentHash),
		Height:        uint32(block.BlockHeader.RawData.Number),
		Confirmations: confirmations,
		Time:          block.BlockHeader.RawData.Timestamp,
	}, nil
}

func (b *TrxRPC) computeConfirmations(n int64) (int, error) {
	//block, err := b.conn.GetNowBlock()
	//if err != nil {
	//	return 0, err
	//}
	//// transaction in the best block has 1 confirmation
	//return int(block.BlockHeader.RawData.Number - n + 1), nil
	return 0, nil
}

func (b *TrxRPC) GetBlockInfo(hash string) (*bchain.BlockInfo, error) {
	block, err := b.conn.GetBlockByID(hash)
	if err != nil {
		return nil, err
	}

	confirmations, err := b.computeConfirmations(block.BlockHeader.RawData.Number)
	if err != nil {
		return nil, err
	}

	var blockInfo bchain.BlockInfo
	blockInfo.BlockHeader = bchain.BlockHeader{
		Hash:          hash,
		Prev:          hex.EncodeToString(block.BlockHeader.RawData.ParentHash),
		Height:        uint32(block.BlockHeader.RawData.Number),
		Confirmations: confirmations,
		Time:          block.BlockHeader.RawData.Timestamp,
	}

	blockInfo.Version = common.JSONNumber(strconv.Itoa(int(block.BlockHeader.RawData.Version)))
	blockInfo.MerkleRoot = hex.EncodeToString(block.BlockHeader.RawData.TxTrieRoot)

	blockExtention, err := b.conn.GetBlockByNum(block.BlockHeader.RawData.Number)
	if err != nil {
		return nil, err
	}

	for _, tx := range blockExtention.Transactions {
		blockInfo.Txids = append(blockInfo.Txids, hex.EncodeToString(tx.Txid))
	}
	return &blockInfo, nil
}

func (b *TrxRPC) GetMempoolTransactions() ([]string, error) {
	return nil, nil
}

func (b *TrxRPC) GetTransaction(txid string) (*bchain.Tx, error) {
	tx, err := b.conn.GetTransactionInfoByID(txid)
	if err != nil {
		return nil, err
	}
	confirmations, err := b.computeConfirmations(tx.BlockNumber)
	if err != nil {
		return nil, err
	}
	return &bchain.Tx{
		Txid:          txid,
		BlockHeight:   uint32(tx.BlockNumber),
		Confirmations: uint32(confirmations),
		Time:          tx.BlockTimeStamp,
	}, nil
}

func (b *TrxRPC) GetChainInfo() (*bchain.ChainInfo, error) {
	block, err := b.conn.GetNowBlock()
	if err != nil {
		return nil, err
	}
	return &bchain.ChainInfo{
		Bestblockhash: hex.EncodeToString(block.Blockid),
		Blocks:        int(block.BlockHeader.RawData.Number),
		Chain:         b.ChainConfig.CoinName,
	}, nil
}

func (b *TrxRPC) GetSubversion() string {
	return ""
}

func (b *TrxRPC) GetCoinName() string {
	return b.ChainConfig.CoinName
}

func (b *TrxRPC) GetTransactionForMempool(txid string) (*bchain.Tx, error) {
	return b.GetTransaction(txid)
}

func (b *TrxRPC) GetTransactionSpecific(tx *bchain.Tx) (json.RawMessage, error) {
	csd, ok := tx.CoinSpecificData.(SpecificTransaction)
	if !ok {
		req := make(map[string]interface{})
		req["value"] = tx.Txid

		//var transaction TrxTx
		//err := b.PostCall("/wallet/gettransactionbyid", req, &transaction)
		//if err != nil {
		//	return nil, err
		//}
		//tx.CoinSpecificData, err = b.Parser.trxtospecifictx(transaction)
		//if err != nil {
		//	return nil, err
		//}
		csd, ok = tx.CoinSpecificData.(SpecificTransaction)
		if !ok {
			return nil, errors.New("Cannot get CoinSpecificData")
		}
	}
	m, err := json.Marshal(&csd)
	return json.RawMessage(m), err
}

func (b *TrxRPC) Initialize() error {
	return nil
}

func (b *TrxRPC) Shutdown(ctx context.Context) error {
	if b.mq != nil {
		if err := b.mq.Shutdown(ctx); err != nil {
			glog.Error("MQ.Shutdown error: ", err)
			return err
		}
	}
	return nil
}

func (b *TrxRPC) SendRawTransaction(tx string) (string, error) {
	var t core.Transaction
	ret, err := b.conn.Broadcast(&t)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(ret.Message), nil
}

func (b *TrxRPC) GetChainParser() bchain.BlockChainParser {
	return b.Parser
}
