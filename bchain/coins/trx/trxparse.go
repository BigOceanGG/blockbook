package trx

import (
	"encoding/hex"
	"github.com/trezor/blockbook/bchain"
	"strconv"
)

type TrxParser struct {
	*bchain.BaseParser
	rpc TrxRPC
}

type TriggerContract struct {
	Constant_result []string `json:"constant_result"`
	Result          struct {
		Result bool `json:"result"`
	} `json:"result"`
}

func has0xPrefix(s string) bool {
	return len(s) >= 2 && s[0] == '0' && (s[1]|32) == 'x'
}

// NewEthereumParser returns new EthereumParser instance
func NewTrxParser(b int, rpc TrxRPC) *TrxParser {
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

func (p *TrxParser) GetContract(owner_address, contract_address, function_selector string) (interface{}, error) {
	req := make(map[string]interface{})
	req["owner_address"] = owner_address
	req["contract_address"] = contract_address
	req["function_selector"] = function_selector

	var trigger TriggerContract
	err := p.rpc.PostCall("/wallet/triggerconstantcontract", req, &trigger)
	if err != nil {
		return nil, err
	}
	return &trigger, nil
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
	return bchain.ChainEthereumType
}

func (p *TrxParser) PackTxid(txid string) ([]byte, error) {
	if has0xPrefix(txid) {
		txid = txid[2:]
	}
	return hex.DecodeString(txid)
}
