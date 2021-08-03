package trx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/juju/errors"
	"github.com/trezor/blockbook/bchain"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"runtime/debug"
	"time"
)

type Configuration struct {
	CoinName            string `json:"coin_name"`
	CoinShortcut        string `json:"coin_shortcut"`
	RPCURL              string `json:"rpc_url"`
	RPCTimeout          int    `json:"rpc_timeout"`
	MessageQueueBinding string `json:"message_queue_binding"`
	MempoolWorkers      int    `json:"mempool_workers"`
	MempoolSubWorkers   int    `json:"mempool_sub_workers"`
}

type TrxRPC struct {
	*bchain.BaseChain
	client      http.Client
	rpcURL      string
	pushHandler func(bchain.NotificationType)
	mq          *bchain.MQ
	Mempool     *bchain.MempoolBitcoinType
	ChainConfig *Configuration
}

type TrxBlock struct {
	BlockID      string `json:"blockID"`
	Block_header struct {
		Raw_data struct {
			Number     uint32 `json:"number"`
			ParentHash string `json:"parentHash"`
			Timestamp  int64  `json:"timestamp"`
		} `json:"raw_data"`
	} `json:"block_header"`
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

func NewTrxRPC(config json.RawMessage, pushHandler func(bchain.NotificationType)) (bchain.BlockChain, error) {
	var err error
	var c Configuration
	err = json.Unmarshal(config, &c)
	if err != nil {
		return nil, errors.Annotatef(err, "Invalid configuration file")
	}

	transport := &http.Transport{
		Dial:                (&net.Dialer{KeepAlive: 600 * time.Second}).Dial,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100, // necessary to not to deplete ports
	}

	s := &TrxRPC{
		BaseChain:   &bchain.BaseChain{},
		client:      http.Client{Timeout: time.Duration(c.RPCTimeout) * time.Second, Transport: transport},
		rpcURL:      c.RPCURL,
		ChainConfig: &c,
		pushHandler: pushHandler,
	}

	return s, nil
}

// CreateMempool creates mempool if not already created, however does not initialize it
func (b *TrxRPC) CreateMempool(chain bchain.BlockChain) (bchain.Mempool, error) {
	if b.Mempool == nil {
		b.Mempool = bchain.NewMempoolBitcoinType(chain, b.ChainConfig.MempoolWorkers, b.ChainConfig.MempoolSubWorkers)
	}
	return b.Mempool, nil
}

// InitializeMempool creates ZeroMQ subscription and sets AddrDescForOutpointFunc to the Mempool
func (b *TrxRPC) InitializeMempool(addrDescForOutpoint bchain.AddrDescForOutpointFunc, onNewTxAddr bchain.OnNewTxAddrFunc, onNewTx bchain.OnNewTxFunc) error {
	if b.Mempool == nil {
		return errors.New("Mempool not created")
	}
	b.Mempool.AddrDescForOutpoint = addrDescForOutpoint
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
	req := make(map[string]string)
	var bestBlock TrxBlock
	err := b.GetCall("/wallet/getnowblock", req, &bestBlock)
	if err != nil {
		return "", err
	}
	return bestBlock.BlockID, nil
}

func (b *TrxRPC) GetBestBlockHeight() (uint32, error) {
	req := make(map[string]string)
	var bestBlock TrxBlock
	err := b.GetCall("/wallet/getnowblock", req, &bestBlock)
	if err != nil {
		return 0, err
	}
	return bestBlock.Block_header.Raw_data.Number, nil
}

func (b *TrxRPC) GetBlock(hash string, height uint32) (*bchain.Block, error) {
	return nil, nil
}

func (b *TrxRPC) GetBlockHash(height uint32) (string, error) {
	req := make(map[string]interface{})
	req["num"] = height

	var block TrxBlock
	err := b.PostCall("/wallet/getblockbynum", req, &block)
	if err != nil {
		return "", err
	}
	return block.BlockID, nil
}

func (b *TrxRPC) GetBlockHeader(hash string) (*bchain.BlockHeader, error) {
	req := make(map[string]interface{})
	req["value"] = hash

	var block TrxBlock
	err := b.PostCall("/wallet/getblockbyid", req, &block)
	if err != nil {
		return nil, err
	}
	confirmations, err := b.computeConfirmations(block.Block_header.Raw_data.Number)
	if err != nil {
		return nil, err
	}
	return &bchain.BlockHeader{
		Hash:          block.BlockID,
		Prev:          block.Block_header.Raw_data.ParentHash,
		Height:        block.Block_header.Raw_data.Number,
		Confirmations: confirmations,
		Time:          block.Block_header.Raw_data.Timestamp,
	}, nil
}

func (b *TrxRPC) computeConfirmations(n uint32) (int, error) {
	bh, err := b.GetBestBlockHeight()
	if err != nil {
		return 0, err
	}
	// transaction in the best block has 1 confirmation
	return int(bh - n + 1), nil
}

func (b *TrxRPC) GetBlockInfo(hash string) (*bchain.BlockInfo, error) {
	return nil, nil
}

func (b *TrxRPC) GetMempoolTransactions() ([]string, error) {
	return nil, nil
}

func (b *TrxRPC) GetTransaction(txid string) (*bchain.Tx, error) {
	return nil, nil
}

func (b *TrxRPC) GetChainInfo() (*bchain.ChainInfo, error) {
	return nil, nil
}

func (b *TrxRPC) GetSubversion() string {
	return ""
}

func (b *TrxRPC) GetCoinName() string {
	return ""
}

func (b *TrxRPC) GetTransactionForMempool(txid string) (*bchain.Tx, error) {
	return nil, nil
}

func (b *TrxRPC) GetTransactionSpecific(tx *bchain.Tx) (json.RawMessage, error) {
	return nil, nil
}

func (b *TrxRPC) Initialize() error {
	//fmt.Println(b.GetBestBlockHeight())
	//fmt.Println(b.GetBestBlockHash())
	//fmt.Println(b.GetBlockHash(32494927))
	fmt.Println(b.GetBlockHeader("0000000001efd54f1668d169d342410e3e5d2c5e8aec17f71cae48bbd0758ab1"))
	return nil
}

func (b *TrxRPC) Shutdown(ctx context.Context) error {
	return nil
}

func (b *TrxRPC) SendRawTransaction(tx string) (string, error) {
	return "", nil
}

func (b *TrxRPC) PostCall(url string, req interface{}, res interface{}) error {
	configData, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest("POST", b.rpcURL+url, bytes.NewBuffer([]byte(configData)))
	if err != nil {
		return err
	}
	httpRes, err := b.client.Do(httpReq)
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

func (b *TrxRPC) GetCall(url string, req map[string]string, res interface{}) error {
	httpReq, err := http.NewRequest("GET", b.rpcURL+url, nil)
	if err != nil {
		return err
	}
	q := httpReq.URL.Query()
	for k, v := range req {
		q.Add(k, v)
	}
	if len(q) > 0 {
		httpReq.URL.RawQuery = q.Encode()
	}
	httpRes, err := b.client.Do(httpReq)
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
