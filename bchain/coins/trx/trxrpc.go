package trx

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	common2 "github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/glog"
	"github.com/juju/errors"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/common"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

const trc20TransferEventSignature = "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

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
	Mempool     *bchain.MempoolTronType
	ChainConfig *Configuration
	Parser      *TrxParser
}

type TrxTxResult struct {
	Result  bool   `json:"result"`
	Code    string `json:"code"`
	Txid    string `json:"txid"`
	Message string `json:"message"`
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
		b.Mempool = bchain.NewMempoolTronType(chain, b.ChainConfig.MempoolTxTimeoutHours, b.ChainConfig.QueryBackendOnMempoolResync)
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

	trans, err := b.conn.GetBlockInfoByNum(block.BlockHeader.RawData.Number)
	if err != nil {
		return nil, err
	}

	if len(trans.TransactionInfo) != len(block.Transactions) {
		glog.Error("Inconsistent number of transactions")
		return nil, errors.New("Inconsistent number of transactions")
	}

	for i, tx := range trans.TransactionInfo {
		btx, err := b.Parser.trxtotx(block.Transactions[i], tx, uint32(confirmations))
		if err == nil {
			cblock.Txs = append(cblock.Txs, *btx)
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
	block, err := b.conn.GetNowBlock()
	if err != nil {
		return 0, err
	}
	// transaction in the best block has 1 confirmation
	return int(block.BlockHeader.RawData.Number - n + 1), nil
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
	tx, err := b.conn.GetTransactionByID(txid)
	if err != nil {
		return nil, err
	}

	txinfo, err := b.conn.GetTransactionInfoByID(txid)
	if err != nil {
		return nil, err
	}

	confirmations, err := b.computeConfirmations(txinfo.BlockNumber)
	if err != nil {
		return nil, err
	}
	return b.Parser.trxtotx(tx, txinfo, uint32(confirmations))
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

func (b *TrxRPC) GetComplete(tx *core.Transaction, txinfo *core.TransactionInfo) (*trxCompleteTransaction, error) {
	contractType := tx.RawData.Contract[0].Type
	data, err := getContract(contractType, tx.RawData.Contract[0].Parameter)
	if err != nil {
		return nil, err
	}

	res := trxCompleteTransaction{
		Tx:     tx,
		TxInfo: txinfo,
	}

	var value bchain.Trc20Transfer
	if contractType == core.Transaction_Contract_TransferContract {
		if v, ok := data["OwnerAddress"]; ok && len(v.([]uint8)) > 0 {
			value.From = hex.EncodeToString(v.([]byte))
		}
		if v, ok := data["ToAddress"]; ok && len(v.([]uint8)) > 0 {
			value.To = hex.EncodeToString(v.([]byte))
		}
		if v, ok := data["Amount"]; ok {
			value.Amount = *big.NewInt(v.(int64))
		}
		res.Value = &value
	} else if contractType == core.Transaction_Contract_TriggerSmartContract {
		if len(txinfo.Log) > 0 && len(txinfo.Log[0].Topics) > 0 && hex.EncodeToString(txinfo.Log[0].Topics[0]) == trc20TransferEventSignature {
			value.From = "41" + hex.EncodeToString(txinfo.Log[0].Topics[1][12:])
			value.To = "41" + hex.EncodeToString(txinfo.Log[0].Topics[2][12:])
			if amount, err := strconv.ParseInt(hex.EncodeToString(txinfo.Log[0].Data), 16, 64); err == nil {
				value.Amount = *big.NewInt(amount)
			}
			if v, ok := data["ContractAddress"]; ok && len(v.([]uint8)) > 0 {
				value.Contract = hex.EncodeToString(v.([]byte))
			}
			res.Value = &value
		}
	}

	return &res, nil
}

func (b *TrxRPC) TronTypeGetTransactionNotify(tx *bchain.Tx) bool {
	csd, ok := tx.CoinSpecificData.(trxCompleteTransaction)
	if ok {
		contractType := csd.Tx.RawData.Contract[0].Type
		if contractType == core.Transaction_Contract_TransferContract {
			return true
		}
		if contractType == core.Transaction_Contract_TriggerSmartContract {
			if len(csd.TxInfo.Log) > 0 &&
				len(csd.TxInfo.Log[0].Topics) > 0 &&
				hex.EncodeToString(csd.TxInfo.Log[0].Topics[0]) == trc20TransferEventSignature {
				return true
			}
		}
	}
	return false
}

func (b *TrxRPC) GetTransactionSpecific(tx *bchain.Tx) (json.RawMessage, error) {
	csd, ok := tx.CoinSpecificData.(trxCompleteTransaction)
	if !ok {
		txx, err := b.conn.GetTransactionByID(tx.Txid)
		if err != nil {
			return nil, err
		}

		txinfo, err := b.conn.GetTransactionInfoByID(tx.Txid)
		if err != nil {
			return nil, err
		}

		complete, err := b.GetComplete(txx, txinfo)
		if err != nil {
			return nil, err
		}
		csd = *complete
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

func (b *TrxRPC) SendRawTransaction(hex string) (string, error) {
	//var tx core.Transaction
	//tx.Signature = append(tx.Signature, []byte(hex))
	//ret, err := b.conn.Broadcast(&tx)
	//if err != nil {
	//	glog.Info(ret)
	//}

	client := http.Client{Timeout: 25 * time.Second}
	req := make(map[string]string)
	req["transaction"] = hex

	var res TrxTxResult
	err := b.PostCall(client, "https://api.trongrid.io/wallet/broadcasthex", req, &res)
	if err != nil {
		return res.Txid, err
	}

	if res.Code != api.ReturnResponseCode_name[0] {
		return res.Code, errors.Errorf("%v", res.Message)
	}

	return res.Txid, nil
}

func safeDecodeResponse(body io.ReadCloser, res interface{}) (err error) {
	var data []byte
	defer func() {
		if r := recover(); r != nil {
			glog.Error("unmarshal json recovered from panic: ", r, "; data: ", string(data))
			debug.PrintStack()
			if len(data) > 0 && len(data) < 2048 {
				err = errors.Errorf("Error: %v", string(data))
			} else {
				err = errors.New("Internal error")
			}
		}
	}()
	data, err = ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &res)
}

func (b *TrxRPC) PostCall(client http.Client, url string, req interface{}, res interface{}) error {
	configData, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(configData)))
	if err != nil {
		return err
	}
	httpRes, err := client.Do(httpReq)
	// in some cases the httpRes can contain data even if it returns error
	// see http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/
	if httpRes != nil {
		defer httpRes.Body.Close()
	}
	if err != nil {
		return err
	}
	// if server returns HTTP error code it might not return json with response
	// handle both cases
	if httpRes.StatusCode != 200 {
		err = safeDecodeResponse(httpRes.Body, &res)
		if err != nil {
			return errors.Errorf("%v %v", httpRes.Status, err)
		}
		return nil
	}
	return safeDecodeResponse(httpRes.Body, &res)
}

func (b *TrxRPC) TronTypeGetBalance(addrDesc bchain.AddressDescriptor) (*big.Int, error) {
	acc, err := b.conn.GetAccount(common2.EncodeCheck(addrDesc))
	if err != nil {
		return nil, err
	}
	return big.NewInt(acc.GetBalance()), nil
}

func (b *TrxRPC) GetChainParser() bchain.BlockChainParser {
	return b.Parser
}
