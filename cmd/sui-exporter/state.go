package main

import (
	"sort"

	"github.com/rpcpool/sui-exporter/internal/suirpc"
)

// snapshot is the chain state from one collection cycle.
type snapshot struct {
	epoch                          uint64
	referenceGasPrice              uint64
	nextEpochReferenceGasPrice     uint64
	epochStartTimestampMs          uint64
	storageFundBalance             uint64
	stakeSubsidyBalance            uint64
	stakeSubsidyCurrentEpochAmount uint64
	totalStake                     uint64

	validators []validator
	reports    []validatorReport
	checkpoint *checkpoint
}

type validator struct {
	address string
	name    string

	votingPower              uint64
	commissionRate           uint64
	gasPrice                 uint64
	stake                    uint64
	pendingStake             uint64
	pendingWithdrawal        uint64
	poolTokenBalance         uint64
	rewardsPool              uint64
	pendingPoolTokenWithdraw uint64
	nextEpochStake           uint64
	nextEpochGasPrice        uint64
	nextEpochCommissionRate  uint64
	nextEpochStakeShare      float64
}

type validatorReport struct {
	address      string
	name         string
	reporter     string
	reporterName string
}

type checkpoint struct {
	sequenceNumber           uint64
	timestampMs              uint64
	totalNetworkTransactions uint64
	computationCost          uint64
	storageCost              uint64
	storageRebate            uint64
}

func snapshotFromEpoch(ep *suirpc.Epoch) *snapshot {
	ss := ep.GetSystemState()
	s := &snapshot{
		epoch:                          ep.GetEpoch(),
		referenceGasPrice:              ep.GetReferenceGasPrice(),
		epochStartTimestampMs:          ss.GetEpochStartTimestampMs(),
		storageFundBalance:             ss.GetStorageFund().GetNonRefundableBalance(),
		stakeSubsidyBalance:            ss.GetStakeSubsidy().GetBalance(),
		stakeSubsidyCurrentEpochAmount: ss.GetStakeSubsidy().GetCurrentDistributionAmount(),
		totalStake:                     ss.GetValidators().GetTotalStake(),
	}

	names := make(map[string]string)
	var totalNextEpochStake uint64
	for _, v := range ss.GetValidators().GetActiveValidators() {
		pool := v.GetStakingPool()
		s.validators = append(s.validators, validator{
			address:                  v.GetAddress(),
			name:                     v.GetName(),
			votingPower:              v.GetVotingPower(),
			commissionRate:           v.GetCommissionRate(),
			gasPrice:                 v.GetGasPrice(),
			stake:                    pool.GetSuiBalance(),
			pendingStake:             pool.GetPendingStake(),
			pendingWithdrawal:        pool.GetPendingTotalSuiWithdraw(),
			poolTokenBalance:         pool.GetPoolTokenBalance(),
			rewardsPool:              pool.GetRewardsPool(),
			pendingPoolTokenWithdraw: pool.GetPendingPoolTokenWithdraw(),
			nextEpochStake:           v.GetNextEpochStake(),
			nextEpochGasPrice:        v.GetNextEpochGasPrice(),
			nextEpochCommissionRate:  v.GetNextEpochCommissionRate(),
		})
		names[v.GetAddress()] = v.GetName()
		totalNextEpochStake += v.GetNextEpochStake()
	}

	if totalNextEpochStake > 0 {
		for i := range s.validators {
			s.validators[i].nextEpochStakeShare = float64(s.validators[i].nextEpochStake) / float64(totalNextEpochStake)
		}
	}
	s.nextEpochReferenceGasPrice = nextEpochReferenceGasPrice(s.validators)

	for _, r := range ss.GetValidatorReportRecords() {
		for _, reporter := range r.GetReporters() {
			s.reports = append(s.reports, validatorReport{
				address:      r.GetReported(),
				name:         names[r.GetReported()],
				reporter:     reporter,
				reporterName: names[reporter],
			})
		}
	}

	return s
}

// nextEpochReferenceGasPrice is the lowest next-epoch gas quote backed by 2/3 of the voting power.
func nextEpochReferenceGasPrice(validators []validator) uint64 {
	quotes := make([]validator, len(validators))
	copy(quotes, validators)
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].nextEpochGasPrice < quotes[j].nextEpochGasPrice })

	var total uint64
	for _, v := range quotes {
		total += v.votingPower
	}

	threshold := 2.0 / 3.0 * float64(total)
	var cumulative uint64
	for _, v := range quotes {
		cumulative += v.votingPower
		if float64(cumulative) >= threshold {
			return v.nextEpochGasPrice
		}
	}
	return 0
}

func checkpointFromProto(cp *suirpc.Checkpoint) *checkpoint {
	sum := cp.GetSummary()
	gas := sum.GetEpochRollingGasCostSummary()
	return &checkpoint{
		sequenceNumber:           cp.GetSequenceNumber(),
		timestampMs:              uint64(sum.GetTimestamp().AsTime().UnixMilli()),
		totalNetworkTransactions: sum.GetTotalNetworkTransactions(),
		computationCost:          gas.GetComputationCost(),
		storageCost:              gas.GetStorageCost(),
		storageRebate:            gas.GetStorageRebate(),
	}
}
