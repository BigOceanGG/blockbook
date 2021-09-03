package trx

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

func getContractInfo(contractType core.Transaction_Contract_ContractType, parameter *any.Any) (interface{}, error) {
	switch contractType {
	case core.Transaction_Contract_TransferContract:
		var c core.TransferContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return c, nil
	case core.Transaction_Contract_TriggerSmartContract:
		var c core.TriggerSmartContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return c, nil
	default:
		return nil, fmt.Errorf("Tx inconsistent")
	}
}

func getContract(contractType core.Transaction_Contract_ContractType, parameter *any.Any) (map[string]interface{}, error) {
	switch contractType {
	case core.Transaction_Contract_AccountCreateContract:
		var c core.AccountCreateContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_TransferContract:
		var c core.TransferContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_TransferAssetContract:
		var c core.TransferAssetContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_VoteWitnessContract:
		var c core.VoteWitnessContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_WitnessCreateContract:
		var c core.WitnessCreateContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_AssetIssueContract:
		var c core.AssetIssueContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_ParticipateAssetIssueContract:
		var c core.ParticipateAssetIssueContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_AccountUpdateContract:
		var c core.AccountUpdateContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_FreezeBalanceContract:
		var c core.FreezeBalanceContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_UnfreezeBalanceContract:
		var c core.UnfreezeBalanceContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_WithdrawBalanceContract:
		var c core.WithdrawBalanceContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_UnfreezeAssetContract:
		var c core.UnfreezeAssetContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_UpdateAssetContract:
		var c core.UpdateAssetContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil

	case core.Transaction_Contract_ProposalCreateContract:
		var c core.ProposalCreateContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_ProposalApproveContract:
		var c core.ProposalApproveContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_ProposalDeleteContract:
		var c core.ProposalDeleteContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_SetAccountIdContract:
		var c core.SetAccountIdContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_CustomContract:
		return nil, fmt.Errorf("Tx inconsistent")
	case core.Transaction_Contract_CreateSmartContract:
		var c core.CreateSmartContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_TriggerSmartContract:
		var c core.TriggerSmartContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_UpdateSettingContract:
		var c core.UpdateSettingContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_ExchangeCreateContract:
		var c core.ExchangeCreateContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_ExchangeInjectContract:
		var c core.ExchangeInjectContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_ExchangeWithdrawContract:
		var c core.ExchangeWithdrawContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_ExchangeTransactionContract:
		var c core.ExchangeTransactionContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_UpdateEnergyLimitContract:
		var c core.UpdateEnergyLimitContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_AccountPermissionUpdateContract:
		var c core.AccountPermissionUpdateContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_ClearABIContract:
		var c core.ClearABIContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_UpdateBrokerageContract:
		var c core.UpdateBrokerageContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	case core.Transaction_Contract_ShieldedTransferContract:
		var c core.ShieldedTransferContract
		if err := ptypes.UnmarshalAny(parameter, &c); err != nil {
			return nil, fmt.Errorf("Tx inconsistent")
		}
		return structs.Map(c), nil
	default:
		return nil, fmt.Errorf("Tx inconsistent")
	}

}
