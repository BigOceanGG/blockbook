package trx

import (
	"encoding/hex"
	"github.com/trezor/blockbook/bchain"
)

type TrxParser struct {
	*bchain.BaseParser
}

func has0xPrefix(s string) bool {
	return len(s) >= 2 && s[0] == '0' && (s[1]|32) == 'x'
}

// NewEthereumParser returns new EthereumParser instance
func NewTrxParser(b int) *TrxParser {
	return &TrxParser{&bchain.BaseParser{
		BlockAddressesToKeep: b,
		AmountDecimalPoint:   18,
	}}
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
