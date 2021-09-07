package trx

import (
	"encoding/hex"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/juju/errors"
	"github.com/trezor/blockbook/bchain"
	"math/big"
)

const TronTypeAddressHexLen = 42
const TronTypeAddressDescriptorLen = 21

type TrxParser struct {
	*bchain.BaseParser
	rpc *TrxRPC
}

type trxCompleteTransaction struct {
	Tx     *core.Transaction `json:"tx,omitempty"`
	Height int64             `json:"height"`
	Type   core.Transaction_Contract_ContractType
	Data   map[string]interface{} `json:"data"`
	Value  *bchain.Trc20Transfer  `json:"value,omitempty"`
}

type TriggerContract struct {
	Constant_result *[]string `json:"constant_result"`
	Result          struct {
		Result bool `json:"result"`
	} `json:"result"`
}

type TrxSpecificContract struct {
	Owner_address    string `json:"owner_address"`
	Contract_address string `json:"contract_address"`
	Type             string `json:"type"`
	Name             string `json:"name"`
	Symbol           string `json:"symbol"`
	Decimals         int    `json:"decimals"`
}

type SpecificTransaction struct {
	TxID             string                `json:"txID"`
	SpecificContract []TrxSpecificContract `json:"specificcontract"`
}

func has0xPrefix(s string) bool {
	return len(s) >= 2 && s[0] == '0' && (s[1]|32) == 'x'
}

// NewEthereumParser returns new EthereumParser instance
func NewTrxParser(b int, rpc *TrxRPC) *TrxParser {
	return &TrxParser{
		&bchain.BaseParser{
			BlockAddressesToKeep: b,
			AmountDecimalPoint:   18,
		},
		rpc}
}

func (p *TrxParser) GetAddrDescFromAddress(address string) (bchain.AddressDescriptor, error) {
	return hex.DecodeString(address)
}

func (p *TrxParser) GetAddrDescFromVout(output *bchain.Vout) (bchain.AddressDescriptor, error) {
	if len(output.ScriptPubKey.Addresses) != 1 {
		return nil, bchain.ErrAddressMissing
	}
	return p.GetAddrDescFromAddress(output.ScriptPubKey.Addresses[0])
}

func (p *TrxParser) GetAddressesFromAddrDesc(addrDesc bchain.AddressDescriptor) ([]string, bool, error) {
	return []string{hex.EncodeToString(addrDesc)}, true, nil
}

func (p *TrxParser) GetScriptFromAddrDesc(addrDesc bchain.AddressDescriptor) ([]byte, error) {
	return addrDesc, nil
}

func (p *TrxParser) GetChainType() bchain.ChainType {
	return bchain.ChainTronType
}

func (p *TrxParser) PackTxid(txid string) ([]byte, error) {
	if has0xPrefix(txid) {
		txid = txid[2:]
	}
	return hex.DecodeString(txid)
}

func (p *TrxParser) EthereumTypeGetErc20FromTx(tx *bchain.Tx) ([]bchain.Erc20Transfer, error) {
	var r []bchain.Erc20Transfer
	return r, nil
}

func (p *TrxParser) PackedTxidLen() int {
	return 32
}

func GetHeightFromTx(tx *bchain.Tx) (uint32, error) {
	csd, ok := tx.CoinSpecificData.(trxCompleteTransaction)
	if !ok {
		return 0, errors.New("Missing CoinSpecificData")
	}
	return uint32(csd.Height), nil
}

func (p *TrxParser) TronTypeGetTrc20FromTx(tx *bchain.Tx) ([]bchain.Trc20Transfer, error) {
	var trcs []bchain.Trc20Transfer
	trx, ok := tx.CoinSpecificData.(trxCompleteTransaction)
	if ok {
		if trx.Value != nil {
			trcs = append(trcs, *trx.Value)
		}
	} else {
		txx, err := p.rpc.conn.GetTransactionByID(tx.Txid)
		if err != nil {
			return nil, err
		}
		complete, err := p.rpc.GetComplete(txx, tx.Txid)
		if err != nil {
			return nil, err
		}
		if complete.Value != nil && complete.Value.Contract != "" {
			trcs = append(trcs, *complete.Value)
		}
	}
	return trcs, nil
}

func (p *TrxParser) trxtotx(txid string, tx *core.Transaction, blocktime int64, confirmations uint32, height int64) (*bchain.Tx, error) {
	complete, err := p.rpc.GetComplete(tx, txid)
	if err != nil {
		return nil, err
	}

	var from, to string
	var amount big.Int
	if complete.Value != nil {
		from = complete.Value.From
		to = complete.Value.To
		amount = complete.Value.Amount
	}
	return &bchain.Tx{
		Blocktime:     blocktime,
		Confirmations: confirmations,
		// Hex
		// LockTime
		Time:        blocktime,
		BlockHeight: uint32(height),
		Txid:        txid,
		Vin: []bchain.Vin{
			{
				Addresses: []string{from},
				// Coinbase
				// ScriptSig
				// Sequence
				// Txid
				// Vout
			},
		},
		Vout: []bchain.Vout{
			{
				N:        0, // there is always up to one To address
				ValueSat: amount,
				ScriptPubKey: bchain.ScriptPubKey{
					// Hex
					Addresses: []string{to},
				},
			},
		},
		CoinSpecificData: *complete,
	}, nil
}
