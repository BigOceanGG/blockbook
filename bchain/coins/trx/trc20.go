package trx

import (
	common2 "github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/trezor/blockbook/bchain"
	"math/big"
	"sync"
)

var cachedContracts = make(map[string]*bchain.Trc20Contract)
var cachedContractsMux sync.Mutex

func (b *TrxRPC) TronTypeGetTrc20ContractInfo(contractDesc bchain.AddressDescriptor) (*bchain.Trc20Contract, error) {
	cds := string(contractDesc)
	cachedContractsMux.Lock()
	contract, found := cachedContracts[cds]
	cachedContractsMux.Unlock()
	if !found {
		address := common2.EncodeCheck(contractDesc)
		symbol, err := b.conn.TRC20GetSymbol(address)
		if err != nil {
			symbol = ""
		}

		tokenDecimals, err := b.conn.TRC20GetDecimals(address)
		if err != nil {
			tokenDecimals = big.NewInt(0)
		}

		name, err := b.conn.TRC20GetName(address)
		if err != nil {
			name = ""
		}
		contract = &bchain.Trc20Contract{
			Contract: address,
			Name:     name,
			Symbol:   symbol,
			Decimals: int(tokenDecimals.Int64()),
		}

		if symbol == "" {
			contract = nil
		}

		cachedContractsMux.Lock()
		cachedContracts[cds] = contract
		cachedContractsMux.Unlock()
	}
	return contract, nil
}

func (b *TrxRPC) TronTypeGetTrc20ContractBalance(addrDesc, contractDesc bchain.AddressDescriptor) (*big.Int, error) {
	value, err := b.conn.TRC20ContractBalance(common2.EncodeCheck(addrDesc), common2.EncodeCheck(contractDesc))
	if err != nil {
		return nil, err
	}
	return value, nil
}
