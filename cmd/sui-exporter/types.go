package main

import (
	"encoding/json"
	"fmt"
)

type SuiAddress string

type SuiValidatorSummary struct {
	SuiAddress                   SuiAddress `json:"suiAddress"`
	P2PAddress                   string     `json:"p2pAddress"`
	Name                         string     `json:"name"`
	CommissionRate               float64    `json:"commissionRate,string"`
	NextEpochPrimaryAddress      string     `json:"nextEpochPrimaryAddress"`
	ImageUrl                     string     `json:"imageUrl"`
	Description                  string     `json:"description"`
	GasPrice                     uint64     `json:"gasPrice,string"`
	ExchangeRatesId              string     `json:"exchangeRatesId"`
	ExchangeRatesSize            uint64     `json:"exchangeRatesSize,string"`
	NetAddress                   string     `json:"netAddress"`
	NextEpochNetAddress          string     `json:"nextEpochNetAddress"`
	NextEpochWorkerAddress       string     `json:"nextEpochWorkerAddress"`
	NextEpochProtocolPubkeyBytes string     `json:"nextEpochProtocolPubkeyBytes"`
	NextEpochGasPrice            uint64     `json:"nextEpochGasPrice,string"`
	NextEpochP2PAddress          string     `json:"nextEpochP2pAddress"`
	NextEpochNetworkPubkeyBytes  string     `json:"nextEpochNetworkPubkeyBytes"`
	NextEpochProofOfPossession   string     `json:"nextEpochProofOfPossession"`
	NextEpochCommissionRate      uint64     `json:"nextEpochCommissionRate,string"`
	NextEpochStakeShare          float64
	NextEpochStake               uint64 `json:"nextEpochStake,string"`
	NextEpochWorkerPubkeyBytes   string `json:"nextEpochWorkerPubkeyBytes"`
	NetworkPubkeyBytes           string `json:"networkPubkeyBytes"`
	OperationCapId               string `json:"operationCapId"`
	PendingStake                 uint64 `json:"pendingStake,string"`
	ProtocolPubkeyBytes          string `json:"protocolPubkeyBytes"`
	ProjectUrl                   string `json:"projectUrl"`
	PendingTotalSuiWithdraw      uint64 `json:"pendingTotalSuiWithdraw,string"`
	PendingPoolTokenWithdraw     uint64 `json:"pendingPoolTokenWithdraw,string"`
	PoolTokenBalance             uint64 `json:"poolTokenBalance,string"`
	PrimaryAddress               string `json:"primaryAddress"`
	ProofOfPossessionBytes       string `json:"proofOfPossessionBytes"`
	RewardsPool                  uint64 `json:"rewardsPool,string"`
	StakingPoolId                string `json:"stakingPoolId"`
	StakingPoolSuiBalance        uint64 `json:"stakingPoolSuiBalance,string"`
	StakingPoolActivationEpoch   uint64 `json:"stakingPoolActivationEpoch,string"`
	StakingPoolDeactivationEpoch uint64 `json:"stakingPoolDeactivationEpoch,string"`
	VotingPower                  uint64 `json:"votingPower,string"`
	WorkerAddress                string `json:"workerAddress"`
	WorkerPubkeyBytes            string `json:"workerPubkeyBytes"`
}

type SuiSystemStateSummary struct {
	ActiveValidators                      []SuiValidatorSummary `json:"activeValidators"`
	AtRiskValidators                      []SuiAddress          `json:"atRiskValidators"`
	Epoch                                 uint64                `json:"epoch,string"`
	EpochDurationMs                       uint64                `json:"epochDurationMs,string"`
	EpochStartTimestampMs                 uint64                `json:"epochStartTimestampMs,string"`
	InactivePoolsId                       string                `json:"inactivePoolsId"`
	InactivePoolsSize                     uint64                `json:"inactivePoolsSize,string"`
	MMxValidatorCount                     uint64                `json:"maxValidatorCount,string"`
	MinValidatorJoiningStake              uint64                `json:"minValidatorJoiningStake,string"`
	PendingActiveValidatorsId             string                `json:"pendingActiveValidatorsId"`
	PendingActiveValidatorsSize           uint64                `json:"pendingActiveValidatorsSize,string"`
	PendingRemovals                       []interface{}         `json:"pendingRemovals"`
	ProtocolVersion                       uint64                `json:"protocolVersion,string"`
	ReferenceGasPrice                     uint64                `json:"referenceGasPrice,string"`
	SafeMode                              bool                  `json:"safeMode"`
	SafeModeComputationRewards            uint64                `json:"safeModeComputationRewards.string"`
	SafeModeNonRefundableStorageFee       uint64                `json:"safeModeNonRefundableStorageFee,string"`
	SafeModeStorageRebates                uint64                `json:"safeModeStorageRebates,string"`
	SafeModeStorageRewards                uint64                `json:"safeModeStorageRewards,string"`
	StakeSubsidyBalance                   uint64                `json:"stakeSubsidyBalance,string"`
	StakeSubsidyCurrentDistributionAmount uint64                `json:"stakeSubsidyCurrentDistributionAmount,string"`
	StakeSubsidyDecreaseRate              uint64                `json:"stakeSubsidyDecreaseRate"`
	StakeSubsidyDistributionCounter       uint64                `json:"stakeSubsidyDistributionCounter,string"`
	StakeSubsidyPeriodLength              uint64                `json:"stakeSubsidyPeriodLength,string"`
	StakeSubsidyStartEpoch                uint64                `json:"stakeSubsidyStartEpoch,string"`
	StakingPoolMappingsId                 string                `json:"stakingPoolMappingsId"`
	StakingPoolMappingsSize               uint64                `json:"stakingPoolMappingsSize,string"`
	StorageFundNonRefundableBalance       uint64                `json:"storageFundNonRefundableBalance,string"`
	StorageFundTotalObjectStorageRebates  uint64                `json:"storageFundTotalObjectStorageRebates,string"`
	SystemStateVersion                    uint64                `json:"systemStateVersion,string"`
	TotalStake                            uint64                `json:"totalStake,string"`
	ValidatorCandidatesId                 string                `json:"validatorCandidatesId"`
	ValidatorCandidatesSize               uint64                `json:"validatorCandidatesSize,string"`
	ValidatorLowStakeGracePeriod          uint64                `json:"validatorLowStakeGracePeriod,string"`
	ValidatorLowStakeThreshold            uint64                `json:"validatorLowStakeThreshold,string"`
	ValidatorReportRecords                []interface{}         `json:"validatorReportRecords"`
	ValidatorVeryLowStakeThreshold        uint64                `json:"validatorVeryLowStakeThreshold,string"`
}

type CheckpointDigest string
type EndOfEpochData interface{}
type TransactionDigest string
type CheckpointCommitment string

type GasCostSummary struct {
	ComputationCost uint64 `json:"computationCost"`
	StorageCost     uint64 `json:"storageCost"`
	StorageRebate   uint64 `json:"storageRebate"`
}

type Checkpoint struct {
	CheckpointCommitments      []CheckpointCommitment `json:"checkpointCommitments"`
	Digest                     CheckpointDigest       `json:"digest"`
	EndOfEpochData             EndOfEpochData         `json:"endOfEpochData"`
	Epoch                      uint64                 `json:" epoch"`
	EpochRollingGasCostSummary GasCostSummary         `json:" epochRollingGasCostSummary"`
	NetworkTotalTransactions   uint64                 `json:" networkTotalTransactions"`
	PreviousDigest             CheckpointDigest       `json:" previousDigest"`
	SequenceNumber             uint64                 `json:" sequenceNumber"`
	TimestampMs                uint64                 `json:" timestampMs"`
	Transactions               []TransactionDigest    `json:" transactions"`
}

type ValidatorReports struct {
	Records []ValidatorReport
}

type ValidatorReport struct {
	Key     string
	Reports []string
}

// Custom unmarshaller to make the validator reports look nicer
// We chuck away the "contents" dicts that exists on both levels
func (vr *ValidatorReports) UnmarshalJSON(b []byte) error {
	if string(b) == "null" || string(b) == `""` {
		return nil
	}

	var f interface{}

	err := json.Unmarshal(b, &f)
	if err != nil {
		return err
	}

	m := f.(map[string]interface{})

	// Has a single key, "contents", skip this key
	reports := m["contents"].([]interface{})

	// Parses all the validator reports
	// @TODO this involves copying a bunch of memory, possibly unnecessary

	var parsed_reports []ValidatorReport

	for _, report := range reports {
		var single_vr *ValidatorReport = new(ValidatorReport)
		r := report.(map[string]interface{})

		single_vr.Key = r["key"].(string)
		c := (r["value"].(map[string]interface{}))["contents"].([]interface{})
		// Append all the reports
		for i := range c {
			single_vr.Reports = append(single_vr.Reports, fmt.Sprintf("%v", c[i]))
		}

		parsed_reports = append(parsed_reports, *single_vr)
	}

	vr.Records = parsed_reports

	return nil
}

type SystemParameters struct {
	MinValidatorStake          uint64 `json:"min_validator_stake"`
	MaxValidatorCandidateCount uint64 `json:"max_validator_candidate_count"`
	StorageGasPrice            uint64 `json:"storage_gas_price"`
}

type StakeSubsidy struct {
	EpochCounter       uint64 `json:"epoch_counter"`
	Balance            Balance
	CurrentEpochAmount uint64 `json:"current_epoch_amount"`
}

type Balance struct {
	Value uint64
}

type Supply struct {
	Value uint64
}

type CheckpointContentsDigest string

type CheckpointSummary struct {
	ContentDigest              CheckpointContentsDigest `json:"content_digest"`
	Epoch                      uint64                   `json:"epoch"`
	EpochRollingGasCostSummary GasCostSummary           `json:"epoch_rolling_gas_cost_summary"`
	NetworkTotalTransactions   uint64                   `json:"network_total_transactions"`
	NextEpochCommittee         interface{}              `json:"next_epoch_committee"`
	PreviousDigest             CheckpointDigest         `json:"previous_digest"`
	SequenceNumber             uint64                   `json:"sequence_number"`
	TimestampMs                uint64                   `json:"timestamp_ms"`
}
