package trx

import (
	"encoding/hex"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/trezor/blockbook/bchain"
	"math/big"
	"strconv"
)

const TronTypeAddressHexLen = 42
const TronTypeAddressDescriptorLen = 21

type TrxParser struct {
	*bchain.BaseParser
	rpc *TrxRPC
}

type trxCompleteTransaction struct {
	Tx    *api.TransactionExtention `json:"tx"`
	Value map[string]interface{}    `json:"value"`
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
	return &TrxParser{&bchain.BaseParser{
		BlockAddressesToKeep: b,
		AmountDecimalPoint:   18,
	},
		rpc}
}

func Hextob(str string) []byte {
	slen := len(str)
	bHex := make([]byte, len(str)/2)
	ii := 0
	for i := 0; i < len(str); i = i + 2 {
		if slen != 1 {
			ss := string(str[i]) + string(str[i+1])
			bt, _ := strconv.ParseInt(ss, 16, 32)
			bHex[ii] = byte(bt)
			ii = ii + 1
			slen = slen - 2
		}
	}
	return bHex
}

func getResult(result string) string {
	if len(result) != 192 {
		return ""
	}
	start, _ := strconv.ParseUint(result[:64], 16, 32)
	length, _ := strconv.ParseUint(result[64:128], 16, 32)
	return string(Hextob(result[64+start*2 : 64+start*2+length*2]))
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

//func (p *TrxParser) trxtospecifictx(tx TrxTx) (SpecificTransaction, error) {
//	var specific SpecificTransaction
//	specific.TxID = tx.TxID
//	for _, raw := range tx.Raw_data.Contract {
//		var specificContract TrxSpecificContract
//		specificContract.Owner_address = raw.Parameter.Value.Owner_address
//		specificContract.Contract_address = raw.Parameter.Value.Contract_address
//		specificContract.Type = raw.Type
//		//if raw.Parameter.Value.Owner_address != "" && raw.Parameter.Value.Contract_address != "" {
//		//	t1, err := p.GetContract(raw.Parameter.Value.Owner_address, raw.Parameter.Value.Contract_address, "name()")
//		//	if err != nil {
//		//		return SpecificTransaction{}, err
//		//	}
//		//	if t1.(TriggerContract).Constant_result != nil {
//		//		for _, res := range *t1.(TriggerContract).Constant_result {
//		//			specificContract.Name = getResult(res)
//		//		}
//		//	}
//		//	t2, err := p.GetContract(raw.Parameter.Value.Owner_address, raw.Parameter.Value.Contract_address, "symbol()")
//		//	if err != nil {
//		//		return SpecificTransaction{}, err
//		//	}
//		//	if t2.(TriggerContract).Constant_result != nil {
//		//		for _, res := range *t2.(TriggerContract).Constant_result {
//		//			specificContract.Symbol = getResult(res)
//		//		}
//		//	}
//		//	t3, err := p.GetContract(raw.Parameter.Value.Owner_address, raw.Parameter.Value.Contract_address, "decimals()")
//		//	if err != nil {
//		//		return SpecificTransaction{}, err
//		//	}
//		//	if t3.(TriggerContract).Constant_result != nil {
//		//		for _, res := range *t3.(TriggerContract).Constant_result {
//		//			decimals, _ := strconv.Atoi(getResult(res))
//		//			specificContract.Decimals = decimals
//		//		}
//		//	}
//		//}
//
//		specific.SpecificContract = append(specific.SpecificContract, specificContract)
//	}
//
//	return specific, nil
//}

// PackedTxidLen returns length in bytes of packed txid
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
		if v, ok := trx.Value["contract_address"]; ok && len(v.([]uint8)) > 0 {
			trc.Contract = string(v.([]byte))
		}
		if v, ok := trx.Value["OwnerAddress"]; ok && len(v.([]uint8)) > 0 {
			trc.From = string(v.([]byte))
		}
		if v, ok := trx.Value["ToAddress"]; ok && len(v.([]uint8)) > 0 {
			trc.To = string(v.([]byte))
		}
		if v, ok := trx.Value["Amount"]; ok {
			trc.Tokens = *big.NewInt(v.(int64))
		}
		trcs = append(trcs, trc)
	}
	return trcs, nil
}

func (p *TrxParser) trxtotx(tx *api.TransactionExtention, blocktime int64, confirmations uint32) bchain.Tx {
	contract, err := getContract(tx.Transaction.RawData.Contract[0].Type, tx.Transaction.RawData.Contract[0].Parameter)
	if err != nil {
		return bchain.Tx{}
	}
	var from, to []string
	var amount int64
	if v, ok := contract["OwnerAddress"]; ok && len(v.([]uint8)) > 0 {
		from = []string{string(v.([]byte))}
	}
	if v, ok := contract["ToAddress"]; ok && len(v.([]uint8)) > 0 {
		to = []string{string(v.([]byte))}
	}
	if v, ok := contract["Amount"]; ok {
		amount = v.(int64)
	}

	ct := trxCompleteTransaction{
		Tx:    tx,
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
	}
}
