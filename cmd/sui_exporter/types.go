package main

import (
	"encoding/json"
	"fmt"
)

type SuiSystemState struct {
	Epoch                  uint64            `json:"epoch"`
	EpochStartTimestampMs  uint64            `json:"epoch_start_timestamp_ms"`
	Info                   string            `json:"uid"`
	Parameters             SystemParameters  `json:"parameters"`
	ReferenceGasPrice      uint64            `json:"reference_gas_price"`
	SafeMode               bool              `json:"safe_mode"`
	StakeSubsidy           StakeSubsidy      `json:"stake_subsidy"`
	StorageFundBalance     Balance           `json:"storage_fund"`
	TreasuryCap            Supply            `json:"treasury_cap"`
	ValidatorReportRecords *ValidatorReports `json:"validator_report_records"`
	Validators             ValidatorSet      `json:"validators"`
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

type ValidatorSet struct {
	ValidatorStake            uint64      `json:"validator_stake"`
	DelegationStake           uint64      `json:"delegation_stake"`
	ActiveValidators          []Validator `json:"active_validators"`
	PendingDelegationSwitches interface{} `json:"pending_delegation_switches"`
}

type Validator struct {
	Metadata              ValidatorMetadata `json:"metadata"`
	VotingPower           uint64            `json:"voting_power"`
	StakeAmount           uint64            `json:"stake_amount"`
	PendingStake          uint64            `json:"pending_stake"`
	PendingWithdraw       uint64            `json:"pending_withdraw"`
	GasPrice              uint64            `json:"gas_price"`
	DelegationStakingPool StakePool         `json:"delegation_staking_pool"`
	CommissionRate        uint64            `json:"commission_rate"`
}

type ValidatorMetadata struct {
	SuiAddress              string `json:"sui_address"`
	PubkeyBytes             []byte `json:"pubkey_bytes"`
	NetworkPubkeyBytes      []byte `json:"network_pubkey_bytes"`
	WorkerPubkeyBytes       []byte `json:"worker_pubkey_bytes"`
	ProofOfPossessionBytes  []byte `json:"proof_of_possession_bytes"`
	Name                    []byte
	Description             []byte
	ImageUrl                []byte `json:"image_url"`
	ProjectUrl              []byte `json:"project_url"`
	NetAddress              []byte `json:"net_address"`
	ConsensusAddress        []byte `json:"consensus_address"`
	WorkerAddress           []byte `json:"worker_address"`
	NextEpochStake          uint64 `json:"next_epoch_stake"`
	NextEpochDelegation     uint64 `json:"next_epoch_delegation"`
	NextEpochGasPrice       uint64 `json:"next_epoch_gas_price"`
	NextEpochCommissionRate uint64 `json:"next_epoch_comission_rate"`
	NextEpochTotalStake     uint64
	NextEpochStakeShare     float64
	NextEpochSelfStakeShare float64
}

type StakePool struct {
	ValidatorAddress      string      `json:"validator_address"`
	StartingEpoch         uint64      `json:"starting_epoch"`
	SuiBalance            uint64      `json:"sui_balance"`
	RewardsPool           Balance     `json:"rewards_pool"`
	DelegationTokenSupply Supply      `json:"delegation_token_supply"`
	PendingWithdraws      Withdrawals `json:"pending_withdraws"`
	PendingDelegations    Delegations `json:"pending_delegations"`
}

type Delegations struct {
	Id   string
	Size uint64
	Head interface{}
	Tail interface{}
}

type Withdrawals struct {
	Contents WithdrawContent
}

type WithdrawContent struct {
	Id   string
	Size uint64
}

type CheckpointContentsDigest string
type CheckpointDigest string

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

type GasCostSummary struct {
	ComputationCost uint64 `json:"computation_cost"`
	StorageCost     uint64 `json:"storage_cost"`
	StorageRebate   uint64 `json:"storage_rebate"`
}
