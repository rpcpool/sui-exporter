package main

import (
	"context"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/ybbus/jsonrpc/v3"
)

type Exporter struct {
	exportValidatorMetrics  bool
	exportCheckpointMetrics bool
	exportValidatorReports  bool

	state             *SuiSystemState
	checkpointSummary *CheckpointSummary

	latestCheckpointSequenceNumber uint64
	totalTransactions              uint64

	epoch          string
	rpcClient      jsonrpc.RPCClient
	frequency      time.Duration
	nodeStateMutex sync.RWMutex

	nextEpochReferenceGasPrice uint64

	referenceGasPriceDesc          *prometheus.Desc
	nextEpochReferenceGasPriceDesc *prometheus.Desc
	epochStartTimestampMs          *prometheus.Desc
	storageFundBalance             *prometheus.Desc
	stakeSubsidyBalance            *prometheus.Desc
	stakeSubsidyCurrentEpochAmount *prometheus.Desc
	totalValidatorStake            *prometheus.Desc
	totalDelegationStake           *prometheus.Desc
	totalTransactionsNumber        *prometheus.Desc

	validatorReport *prometheus.Desc

	// Validator specific metrics
	validatorVotingPower             *prometheus.Desc
	validatorCommission              *prometheus.Desc
	validatorGasPrice                *prometheus.Desc
	validatorPendingStake            *prometheus.Desc
	validatorPendingWithdrawal       *prometheus.Desc
	validatorStake                   *prometheus.Desc
	validatorDelegationBalance       *prometheus.Desc
	validatorNextEpochStake          *prometheus.Desc
	validatorNextEpochDelegation     *prometheus.Desc
	validatorNextEpochGasPrice       *prometheus.Desc
	validatorNextEpochCommission     *prometheus.Desc
	validatorNextEpochStakeShare     *prometheus.Desc
	validatorNextEpochSelfStakeShare *prometheus.Desc

	// Checkpoint metrics
	checkpointSummaryDesc              *prometheus.Desc
	checkpointSequenceNumber           *prometheus.Desc
	checkpointTimestampMs              *prometheus.Desc
	checkpointNetworkTotalTransactions *prometheus.Desc
	checkpointEpochComputationCost     *prometheus.Desc
	checkpointEpochStorageCost         *prometheus.Desc
	checkpointEpochStorageRebate       *prometheus.Desc
}

func NewExporter(uri string, frequency int, exportValidatorMetrics bool, exportCheckpointMetrics bool, exportValidatorReports bool) *Exporter {
	return &Exporter{
		state:                   nil,
		exportValidatorMetrics:  exportValidatorMetrics,
		exportCheckpointMetrics: exportCheckpointMetrics,
		exportValidatorReports:  exportValidatorReports,
		rpcClient:               jsonrpc.NewClient(uri),
		frequency:               time.Duration(frequency),
		referenceGasPriceDesc: prometheus.NewDesc(
			"sui_reference_gas_price",
			"Information about gas price",
			[]string{"epoch"}, nil,
		),
		nextEpochReferenceGasPriceDesc: prometheus.NewDesc(
			"sui_next_epoch_reference_gas_price",
			"Running estimation of the next epoch's reference gas price",
			[]string{"epoch"}, nil,
		),
		epochStartTimestampMs: prometheus.NewDesc(
			"sui_epoch_start_timestamp_ms",
			"Information about epoch start timestamp",
			[]string{"epoch"}, nil,
		),
		storageFundBalance: prometheus.NewDesc(
			"sui_storage_fund_balance",
			"Storage fund balance",
			[]string{"epoch"}, nil,
		),
		stakeSubsidyBalance: prometheus.NewDesc(
			"sui_stake_subsidy_balance",
			"Stake subsidy balance",
			[]string{"epoch"}, nil,
		),
		stakeSubsidyCurrentEpochAmount: prometheus.NewDesc(
			"sui_stake_subsidy_current_epoch_amount",
			"Stake subsidy current epoch amount",
			[]string{"epoch"}, nil,
		),
		totalValidatorStake: prometheus.NewDesc(
			"sui_total_validator_stake",
			"Total validator stake",
			[]string{"epoch"}, nil,
		),
		totalDelegationStake: prometheus.NewDesc(
			"sui_total_delegation_stake",
			"Total delegation stake",
			[]string{"epoch"}, nil,
		),
		validatorVotingPower: prometheus.NewDesc(
			"sui_validator_voting_power",
			"Validator voting power",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorCommission: prometheus.NewDesc(
			"sui_validator_commission",
			"Validator commission",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorGasPrice: prometheus.NewDesc(
			"sui_validator_gas_price",
			"Validator gas price",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorPendingStake: prometheus.NewDesc(
			"sui_validator_pending_stake",
			"Validator pending stake",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorPendingWithdrawal: prometheus.NewDesc(
			"sui_validator_pending_withdrawal",
			"Validator pending withdrawal",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorStake: prometheus.NewDesc(
			"sui_validator_stake",
			"Validator stake",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorDelegationBalance: prometheus.NewDesc(
			"sui_validator_delegation_balance",
			"Validator delegation balance",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorNextEpochStake: prometheus.NewDesc(
			"sui_validator_next_epoch_stake",
			"Validator next epoch stake",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorNextEpochDelegation: prometheus.NewDesc(
			"sui_validator_next_epoch_delegation",
			"Validator next epoch delegation",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorNextEpochGasPrice: prometheus.NewDesc(
			"sui_validator_next_epoch_gas_price",
			"Validator next epoch gas price",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorNextEpochCommission: prometheus.NewDesc(
			"sui_validator_next_epoch_commission",
			"Validator next epoch commission",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorNextEpochStakeShare: prometheus.NewDesc(
			"sui_validator_next_epoch_stake_share",
			"Validator next epoch stake share",
			[]string{"epoch", "address", "name"}, nil,
		),
		validatorNextEpochSelfStakeShare: prometheus.NewDesc(
			"sui_validator_next_epoch_self_stake_share",
			"Validator next epoch self stake share",
			[]string{"epoch", "address", "name"}, nil,
		),
		totalTransactionsNumber: prometheus.NewDesc(
			"sui_total_transactions_number",
			"Total transactions number",
			[]string{"epoch"}, nil,
		),
		checkpointSequenceNumber: prometheus.NewDesc(
			"sui_checkpoint_sequence_number",
			"Checkpoint sequence number",
			[]string{"epoch"}, nil,
		),
		checkpointSummaryDesc: prometheus.NewDesc(
			"sui_checkpoint_summary",
			"Checkpoint summary",
			[]string{"epoch", "checkpoint", "digest", "previous_digest", "final_checkpoint"}, nil,
		),
		checkpointTimestampMs: prometheus.NewDesc(
			"sui_checkpoint_timestamp_ms",
			"Checkpoint timestamp in milliseconds",

			[]string{"epoch", "checkpoint"}, nil,
		),
		checkpointNetworkTotalTransactions: prometheus.NewDesc(
			"sui_checkpoint_network_total_transactions",
			"Checkpoint network total transactions",
			[]string{"epoch", "checkpoint"}, nil,
		),
		checkpointEpochComputationCost: prometheus.NewDesc(
			"sui_checkpoint_epoch_computation_cost",
			"Checkpoint epoch computation cost",
			[]string{"epoch", "checkpoint"}, nil,
		),
		checkpointEpochStorageCost: prometheus.NewDesc(
			"sui_checkpoint_epoch_storage_cost",
			"Checkpoint epoch storage cost",
			[]string{"epoch", "checkpoint"}, nil,
		),
		checkpointEpochStorageRebate: prometheus.NewDesc(
			"sui_checkpoint_epoch_storage_rebate",
			"Checkpoint epoch storage rebate",
			[]string{"epoch", "checkpoint"}, nil,
		),
		validatorReport: prometheus.NewDesc(
			"sui_validator_report",
			"Validator slashing report",
			[]string{"epoch", "address", "reporter"}, nil,
		),
	}

}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.referenceGasPriceDesc
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	if e.state == nil {
		return
	}
	e.nodeStateMutex.RLock()

	ch <- prometheus.MustNewConstMetric(e.referenceGasPriceDesc, prometheus.GaugeValue, float64(e.state.ReferenceGasPrice), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.nextEpochReferenceGasPriceDesc, prometheus.GaugeValue, float64(e.nextEpochReferenceGasPrice), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.epochStartTimestampMs, prometheus.GaugeValue, float64(e.state.EpochStartTimestampMs), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.storageFundBalance, prometheus.GaugeValue, float64(e.state.StorageFundBalance.Value), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.stakeSubsidyBalance, prometheus.GaugeValue, float64(e.state.StakeSubsidy.Balance.Value), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.stakeSubsidyCurrentEpochAmount, prometheus.GaugeValue, float64(e.state.StakeSubsidy.CurrentEpochAmount), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.totalValidatorStake, prometheus.GaugeValue, float64(e.state.Validators.ValidatorStake), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.totalDelegationStake, prometheus.GaugeValue, float64(e.state.Validators.DelegationStake), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.totalTransactionsNumber, prometheus.GaugeValue, float64(e.totalTransactions), e.epoch)

	if e.exportCheckpointMetrics {
		ch <- prometheus.MustNewConstMetric(e.checkpointSequenceNumber, prometheus.GaugeValue, float64(e.latestCheckpointSequenceNumber), e.epoch)

		lastCheckpoint := (e.checkpointSummary.NextEpochCommittee != nil)
		ch <- prometheus.MustNewConstMetric(e.checkpointSummaryDesc, prometheus.GaugeValue, 1, e.epoch, strconv.Itoa(int(e.checkpointSummary.SequenceNumber)), string(e.checkpointSummary.ContentDigest), string(e.checkpointSummary.PreviousDigest), strconv.FormatBool(lastCheckpoint))

		ch <- prometheus.MustNewConstMetric(e.checkpointTimestampMs, prometheus.GaugeValue, float64(e.checkpointSummary.TimestampMs), e.epoch, strconv.Itoa(int(e.checkpointSummary.SequenceNumber)))
		ch <- prometheus.MustNewConstMetric(e.checkpointNetworkTotalTransactions, prometheus.GaugeValue, float64(e.checkpointSummary.NetworkTotalTransactions), e.epoch, strconv.Itoa(int(e.checkpointSummary.SequenceNumber)))
		ch <- prometheus.MustNewConstMetric(e.checkpointEpochComputationCost, prometheus.GaugeValue, float64(e.checkpointSummary.EpochRollingGasCostSummary.ComputationCost), e.epoch, strconv.Itoa(int(e.checkpointSummary.SequenceNumber)))
		ch <- prometheus.MustNewConstMetric(e.checkpointEpochStorageCost, prometheus.GaugeValue, float64(e.checkpointSummary.EpochRollingGasCostSummary.StorageCost), e.epoch, strconv.Itoa(int(e.checkpointSummary.SequenceNumber)))
		ch <- prometheus.MustNewConstMetric(e.checkpointEpochStorageRebate, prometheus.GaugeValue, float64(e.checkpointSummary.EpochRollingGasCostSummary.StorageRebate), e.epoch, strconv.Itoa(int(e.checkpointSummary.SequenceNumber)))
	}
	if e.exportValidatorMetrics {
		for _, validator := range e.state.Validators.ActiveValidators {
			ch <- prometheus.MustNewConstMetric(e.validatorVotingPower, prometheus.GaugeValue, float64(validator.VotingPower), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorCommission, prometheus.GaugeValue, float64(validator.CommissionRate), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorGasPrice, prometheus.GaugeValue, float64(validator.GasPrice), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorPendingStake, prometheus.GaugeValue, float64(validator.PendingStake), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorPendingWithdrawal, prometheus.GaugeValue, float64(validator.PendingWithdraw), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorStake, prometheus.GaugeValue, float64(validator.StakeAmount), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorDelegationBalance, prometheus.GaugeValue, float64(validator.DelegationStakingPool.SuiBalance), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorNextEpochStake, prometheus.GaugeValue, float64(validator.Metadata.NextEpochStake), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorNextEpochDelegation, prometheus.GaugeValue, float64(validator.Metadata.NextEpochDelegation), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorNextEpochGasPrice, prometheus.GaugeValue, float64(validator.Metadata.NextEpochGasPrice), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorNextEpochCommission, prometheus.GaugeValue, float64(validator.Metadata.NextEpochCommissionRate), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorNextEpochStakeShare, prometheus.GaugeValue, float64(validator.Metadata.NextEpochStakeShare), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
			ch <- prometheus.MustNewConstMetric(e.validatorNextEpochSelfStakeShare, prometheus.GaugeValue, float64(validator.Metadata.NextEpochSelfStakeShare), e.epoch, validator.Metadata.SuiAddress, string(validator.Metadata.Name[:]))
		}
	}

	if e.exportValidatorReports {
		for _, reportedValidator := range e.state.ValidatorReportRecords.Records {

			for _, report := range reportedValidator.Reports {
				ch <- prometheus.MustNewConstMetric(e.validatorReport, prometheus.GaugeValue, float64(1), e.epoch, reportedValidator.Key, report)
			}
		}
	}

	e.nodeStateMutex.RUnlock()
}

type GasQuote struct {
	quote uint64
	vote  uint64
}

func (e *Exporter) WatchState() {
	ticker := time.NewTicker(e.frequency * time.Second)

	for {
		e.nodeStateMutex.Lock()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		err := e.rpcClient.CallFor(ctx, &e.totalTransactions, "sui_getTotalTransactionNumber")
		if err != nil {
			log.Printf("Error getting transaction number: %v", err)
			continue
		}

		if e.exportCheckpointMetrics {
			err = e.rpcClient.CallFor(ctx, &e.latestCheckpointSequenceNumber, "sui_getLatestCheckpointSequenceNumber")
			if err != nil {
				log.Printf("Error getting latest checkpoint sequence number: %v", err)
				continue
			}

			err = e.rpcClient.CallFor(ctx, &e.checkpointSummary, "sui_getCheckpointSummary", e.latestCheckpointSequenceNumber)
			if err != nil {
				log.Printf("Error getting checkpoint summary: %v", err)
				continue
			}
		}

		err = e.rpcClient.CallFor(ctx, &e.state, "sui_getSuiSystemState")
		if err != nil {
			log.Printf("Error getting node state: %v", err)
			continue
		}

		if e.exportValidatorMetrics {
			var totalStake uint64
			var totalVotingPower uint64

			var gasQuotes []GasQuote

			for i, validator := range e.state.Validators.ActiveValidators {
				// Calculate the next epoch expected share
				totalValidatorStake := validator.Metadata.NextEpochStake + validator.Metadata.NextEpochDelegation
				gasQuote := GasQuote{
					quote: validator.Metadata.NextEpochGasPrice,
					vote:  validator.VotingPower,
				}

				gasQuotes = append(gasQuotes, gasQuote)
				e.state.Validators.ActiveValidators[i].Metadata.NextEpochTotalStake = totalValidatorStake
				totalStake += totalValidatorStake
				totalVotingPower += validator.VotingPower
			}

			// We count 2/3 by stake weight
			var cumulativeVotePower uint64
			countedVotePower := 2.0 / 3.0 * float64(totalVotingPower)
			sort.Slice(gasQuotes, func(i, j int) bool { return gasQuotes[i].quote < gasQuotes[j].quote })
			for _, quote := range gasQuotes {
				cumulativeVotePower += quote.vote
				if float64(cumulativeVotePower) >= countedVotePower {
					e.nextEpochReferenceGasPrice = quote.quote
					break
				}
			}

			// Get the validator fee calculation params
			for i, validator := range e.state.Validators.ActiveValidators {
				e.state.Validators.ActiveValidators[i].Metadata.NextEpochSelfStakeShare =
					float64(validator.Metadata.NextEpochStake) /
						float64(validator.Metadata.NextEpochTotalStake)
				e.state.Validators.ActiveValidators[i].Metadata.NextEpochStakeShare =
					float64(validator.Metadata.NextEpochTotalStake) / float64(totalStake)
			}

		}

		// Store epoch as string for prometheus labels
		e.epoch = strconv.Itoa(int(e.state.Epoch))

		log.Println("Updated node state: checkpoint", e.state.Epoch)
		cancel()

		e.nodeStateMutex.Unlock()
		<-ticker.C
	}
}
