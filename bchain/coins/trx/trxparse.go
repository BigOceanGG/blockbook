package trx

import (
	"encoding/hex"
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
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
	Tx    *api.TransactionExtention              `json:"tx"`
	Type  core.Transaction_Contract_ContractType `json:"type"`
	Value interface{}                            `json:"value"`
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
	return nil, nil
}

func (p *TrxParser) GetAddressesFromAddrDesc(addrDesc bchain.AddressDescriptor) ([]string, bool, error) {
	return []string{}, true, nil
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

func (p *TrxParser) TronTypeGetTrc20FromTx(tx *bchain.Tx) ([]bchain.Trc20Transfer, error) {
	var trcs []bchain.Trc20Transfer
	var err error
	trx, ok := tx.CoinSpecificData.(trxCompleteTransaction)
	if ok {
		if err != nil {
			return trcs, err
		}
		var trc bchain.Trc20Transfer
		if trx.Type == core.Transaction_Contract_TransferContract {
			data := trx.Value.(core.TransferContract)
			trc.From = hex.EncodeToString(data.OwnerAddress)
			trc.To = hex.EncodeToString(data.ToAddress)
			trc.Contract = ""
			trc.Tokens = *big.NewInt(data.Amount)
		} else {
			data := trx.Value.(core.TriggerSmartContract)
			trc.From = hex.EncodeToString(data.OwnerAddress)
			trc.Contract = hex.EncodeToString(data.ContractAddress)
			trc.To = hex.EncodeToString(data.ContractAddress)
			trc.Tokens = *big.NewInt(0)
		}
		trcs = append(trcs, trc)
	}
	return trcs, nil
}

func (p *TrxParser) trxtotx(tx *api.TransactionExtention, blocktime int64, confirmations uint32) (bchain.Tx, error) {
	if len(tx.Transaction.RawData.Contract) == 0 {
		return bchain.Tx{}, fmt.Errorf("No contract")
	}

	contractType := tx.Transaction.RawData.Contract[0].Type
	contract, err := getContractInfo(contractType, tx.Transaction.RawData.Contract[0].Parameter)
	if err != nil {
		return bchain.Tx{}, err
	}

	var from, to []string
	var amount int64
	if contractType == core.Transaction_Contract_TransferContract {
		data := contract.(core.TransferContract)
		from = []string{hex.EncodeToString(data.OwnerAddress)}
		to = []string{hex.EncodeToString(data.ToAddress)}
		amount = data.Amount
	} else {
		data := contract.(core.TriggerSmartContract)
		from = []string{hex.EncodeToString(data.OwnerAddress)}
		to = []string{hex.EncodeToString(data.ContractAddress)}
		amount = 0
	}

	ct := trxCompleteTransaction{
		Tx:    tx,
		Type:  contractType,
		Value: contract,
	}
	return bchain.Tx{
		Blocktime:     blocktime,
		Confirmations: confirmations,
		// Hex
		// LockTime
		Time: blocktime,
		Txid: hex.EncodeToString(tx.Txid),
		Vin: []bchain.Vin{
			{
				Addresses: from,
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
				ValueSat: *big.NewInt(amount),
				ScriptPubKey: bchain.ScriptPubKey{
					// Hex
					Addresses: to,
				},
			},
		},
		CoinSpecificData: ct,
	}, nil
}
