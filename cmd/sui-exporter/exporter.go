package main

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/rpcpool/sui-exporter/internal/suirpc"
)

// ledgerClient is the part of suirpc.LedgerServiceClient the exporter uses.
type ledgerClient interface {
	GetEpoch(ctx context.Context, in *suirpc.GetEpochRequest, opts ...grpc.CallOption) (*suirpc.GetEpochResponse, error)
	GetCheckpoint(ctx context.Context, in *suirpc.GetCheckpointRequest, opts ...grpc.CallOption) (*suirpc.GetCheckpointResponse, error)
}

var (
	epochReadMask      = &fieldmaskpb.FieldMask{Paths: []string{"epoch", "reference_gas_price", "system_state"}}
	checkpointReadMask = &fieldmaskpb.FieldMask{Paths: []string{
		"sequence_number", "summary.timestamp", "summary.total_network_transactions", "summary.epoch_rolling_gas_cost_summary",
	}}
)

type Exporter struct {
	exportValidatorMetrics  bool
	exportCheckpointMetrics bool
	exportValidatorReports  bool

	client    ledgerClient
	frequency time.Duration
	timeout   time.Duration

	mu    sync.RWMutex
	state *snapshot

	lastSuccess prometheus.Gauge
	errors      *prometheus.CounterVec

	referenceGasPrice              *prometheus.Desc
	nextEpochReferenceGasPrice     *prometheus.Desc
	epochStartTimestampMs          *prometheus.Desc
	storageFundBalance             *prometheus.Desc
	stakeSubsidyBalance            *prometheus.Desc
	stakeSubsidyCurrentEpochAmount *prometheus.Desc
	totalStake                     *prometheus.Desc
	totalTransactionsNumber        *prometheus.Desc

	validatorVotingPower              *prometheus.Desc
	validatorCommission               *prometheus.Desc
	validatorGasPrice                 *prometheus.Desc
	validatorPendingStake             *prometheus.Desc
	validatorPendingWithdrawal        *prometheus.Desc
	validatorStake                    *prometheus.Desc
	validatorNextEpochStake           *prometheus.Desc
	validatorNextEpochGasPrice        *prometheus.Desc
	validatorNextEpochCommission      *prometheus.Desc
	validatorNextEpochStakeShare      *prometheus.Desc
	validatorRewardPool               *prometheus.Desc
	validatorPoolTokenBalance         *prometheus.Desc
	validatorPendingPoolTokenWithdraw *prometheus.Desc
	validatorReport                   *prometheus.Desc

	checkpointSequenceNumber           *prometheus.Desc
	checkpointTimestampMs              *prometheus.Desc
	checkpointNetworkTotalTransactions *prometheus.Desc
	checkpointEpochComputationCost     *prometheus.Desc
	checkpointEpochStorageCost         *prometheus.Desc
	checkpointEpochStorageRebate       *prometheus.Desc
}

func NewExporter(client ledgerClient, frequency, timeout time.Duration, exportValidatorMetrics, exportCheckpointMetrics, exportValidatorReports bool) *Exporter {
	epoch := []string{"epoch"}
	val := []string{"epoch", "address", "name"}

	return &Exporter{
		exportValidatorMetrics:  exportValidatorMetrics,
		exportCheckpointMetrics: exportCheckpointMetrics,
		exportValidatorReports:  exportValidatorReports,
		client:                  client,
		frequency:               frequency,
		timeout:                 timeout,

		lastSuccess: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sui_exporter_last_success_timestamp_seconds",
			Help: "Unix time of the last successful GetEpoch call",
		}),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "sui_exporter_errors_total",
			Help: "Failed gRPC calls by method",
		}, []string{"call"}),

		referenceGasPrice:              prometheus.NewDesc("sui_reference_gas_price", "Information about gas price", epoch, nil),
		nextEpochReferenceGasPrice:     prometheus.NewDesc("sui_next_epoch_reference_gas_price", "Running estimation of the next epoch's reference gas price", epoch, nil),
		epochStartTimestampMs:          prometheus.NewDesc("sui_epoch_start_timestamp_ms", "Information about epoch start timestamp", epoch, nil),
		storageFundBalance:             prometheus.NewDesc("sui_storage_fund_balance", "Storage fund balance", epoch, nil),
		stakeSubsidyBalance:            prometheus.NewDesc("sui_stake_subsidy_balance", "Stake subsidy balance", epoch, nil),
		stakeSubsidyCurrentEpochAmount: prometheus.NewDesc("sui_stake_subsidy_current_epoch_amount", "Stake subsidy current epoch amount", epoch, nil),
		totalStake:                     prometheus.NewDesc("sui_total_stake", "Total stake", epoch, nil),
		totalTransactionsNumber:        prometheus.NewDesc("sui_total_transactions_number", "Total transactions number", epoch, nil),

		validatorVotingPower:              prometheus.NewDesc("sui_validator_voting_power", "Validator voting power", val, nil),
		validatorCommission:               prometheus.NewDesc("sui_validator_commission", "Validator commission", val, nil),
		validatorGasPrice:                 prometheus.NewDesc("sui_validator_gas_price", "Validator gas price", val, nil),
		validatorPendingStake:             prometheus.NewDesc("sui_validator_pending_stake", "Validator pending stake", val, nil),
		validatorPendingWithdrawal:        prometheus.NewDesc("sui_validator_pending_withdrawal", "Validator pending withdrawal", val, nil),
		validatorStake:                    prometheus.NewDesc("sui_validator_stake", "Validator stake", val, nil),
		validatorNextEpochStake:           prometheus.NewDesc("sui_validator_next_epoch_stake", "Validator next epoch stake", val, nil),
		validatorNextEpochGasPrice:        prometheus.NewDesc("sui_validator_next_epoch_gas_price", "Validator next epoch gas price", val, nil),
		validatorNextEpochCommission:      prometheus.NewDesc("sui_validator_next_epoch_commission", "Validator next epoch commission", val, nil),
		validatorNextEpochStakeShare:      prometheus.NewDesc("sui_validator_next_epoch_stake_share", "Validator next epoch stake share", val, nil),
		validatorRewardPool:               prometheus.NewDesc("sui_validator_reward_pool", "Validator reward pool", val, nil),
		validatorPoolTokenBalance:         prometheus.NewDesc("sui_validator_pool_token_balance", "Validator pool token balance", val, nil),
		validatorPendingPoolTokenWithdraw: prometheus.NewDesc("sui_validator_pending_pool_token_withdraw", "Validator pending pool token withdraw", val, nil),
		validatorReport: prometheus.NewDesc("sui_validator_report", "Validator slashing report",
			[]string{"epoch", "address", "name", "reporter", "reporter_name"}, nil),

		checkpointSequenceNumber:           prometheus.NewDesc("sui_checkpoint_sequence_number", "Checkpoint sequence number", epoch, nil),
		checkpointTimestampMs:              prometheus.NewDesc("sui_checkpoint_timestamp_ms", "Checkpoint timestamp in milliseconds", epoch, nil),
		checkpointNetworkTotalTransactions: prometheus.NewDesc("sui_checkpoint_network_total_transactions", "Checkpoint network total transactions", epoch, nil),
		checkpointEpochComputationCost:     prometheus.NewDesc("sui_checkpoint_epoch_computation_cost", "Checkpoint epoch computation cost", epoch, nil),
		checkpointEpochStorageCost:         prometheus.NewDesc("sui_checkpoint_epoch_storage_cost", "Checkpoint epoch storage cost", epoch, nil),
		checkpointEpochStorageRebate:       prometheus.NewDesc("sui_checkpoint_epoch_storage_rebate", "Checkpoint epoch storage rebate", epoch, nil),
	}
}

// Describe sends nothing, so the registry treats the exporter as unchecked.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.lastSuccess.Collect(ch)
	e.errors.Collect(ch)

	e.mu.RLock()
	s := e.state
	e.mu.RUnlock()
	if s == nil {
		return
	}

	epoch := strconv.FormatUint(s.epoch, 10)
	gauge := func(d *prometheus.Desc, v uint64, labels ...string) {
		ch <- prometheus.MustNewConstMetric(d, prometheus.GaugeValue, float64(v), labels...)
	}

	gauge(e.referenceGasPrice, s.referenceGasPrice, epoch)
	gauge(e.nextEpochReferenceGasPrice, s.nextEpochReferenceGasPrice, epoch)
	gauge(e.epochStartTimestampMs, s.epochStartTimestampMs, epoch)
	gauge(e.storageFundBalance, s.storageFundBalance, epoch)
	gauge(e.stakeSubsidyBalance, s.stakeSubsidyBalance, epoch)
	gauge(e.stakeSubsidyCurrentEpochAmount, s.stakeSubsidyCurrentEpochAmount, epoch)
	gauge(e.totalStake, s.totalStake, epoch)

	if cp := s.checkpoint; cp != nil {
		gauge(e.totalTransactionsNumber, cp.totalNetworkTransactions, epoch)
		if e.exportCheckpointMetrics {
			gauge(e.checkpointSequenceNumber, cp.sequenceNumber, epoch)
			gauge(e.checkpointTimestampMs, cp.timestampMs, epoch)
			gauge(e.checkpointNetworkTotalTransactions, cp.totalNetworkTransactions, epoch)
			gauge(e.checkpointEpochComputationCost, cp.computationCost, epoch)
			gauge(e.checkpointEpochStorageCost, cp.storageCost, epoch)
			gauge(e.checkpointEpochStorageRebate, cp.storageRebate, epoch)
		}
	}

	if e.exportValidatorMetrics {
		for _, v := range s.validators {
			gauge(e.validatorVotingPower, v.votingPower, epoch, v.address, v.name)
			gauge(e.validatorCommission, v.commissionRate, epoch, v.address, v.name)
			gauge(e.validatorGasPrice, v.gasPrice, epoch, v.address, v.name)
			gauge(e.validatorPendingStake, v.pendingStake, epoch, v.address, v.name)
			gauge(e.validatorPendingWithdrawal, v.pendingWithdrawal, epoch, v.address, v.name)
			gauge(e.validatorStake, v.stake, epoch, v.address, v.name)
			gauge(e.validatorNextEpochStake, v.nextEpochStake, epoch, v.address, v.name)
			gauge(e.validatorNextEpochGasPrice, v.nextEpochGasPrice, epoch, v.address, v.name)
			gauge(e.validatorNextEpochCommission, v.nextEpochCommissionRate, epoch, v.address, v.name)
			gauge(e.validatorPoolTokenBalance, v.poolTokenBalance, epoch, v.address, v.name)
			gauge(e.validatorRewardPool, v.rewardsPool, epoch, v.address, v.name)
			gauge(e.validatorPendingPoolTokenWithdraw, v.pendingPoolTokenWithdraw, epoch, v.address, v.name)
			ch <- prometheus.MustNewConstMetric(e.validatorNextEpochStakeShare, prometheus.GaugeValue, v.nextEpochStakeShare, epoch, v.address, v.name)
		}
	}

	if e.exportValidatorReports {
		for _, r := range s.reports {
			ch <- prometheus.MustNewConstMetric(e.validatorReport, prometheus.CounterValue, 1, epoch, r.address, r.name, r.reporter, r.reporterName)
		}
	}
}

// update fetches the chain state once. A failed GetEpoch clears the state so stale values are not exported.
func (e *Exporter) update() {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	epochResp, err := e.client.GetEpoch(ctx, &suirpc.GetEpochRequest{ReadMask: epochReadMask})
	if err != nil {
		e.errors.WithLabelValues("GetEpoch").Inc()
		log.Printf("Error getting epoch: %v", err)
		e.mu.Lock()
		e.state = nil
		e.mu.Unlock()
		return
	}
	s := snapshotFromEpoch(epochResp.GetEpoch())

	cpResp, err := e.client.GetCheckpoint(ctx, &suirpc.GetCheckpointRequest{ReadMask: checkpointReadMask})
	if err != nil {
		e.errors.WithLabelValues("GetCheckpoint").Inc()
		log.Printf("Error getting latest checkpoint: %v", err)
	} else {
		s.checkpoint = checkpointFromProto(cpResp.GetCheckpoint())
	}

	e.mu.Lock()
	e.state = s
	e.mu.Unlock()
	e.lastSuccess.SetToCurrentTime()
	log.Println("Updated node state: epoch", s.epoch)
}

func (e *Exporter) WatchState() {
	ticker := time.NewTicker(e.frequency)
	defer ticker.Stop()
	for {
		e.update()
		<-ticker.C
	}
}
