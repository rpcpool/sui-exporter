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
	state          *SuiSystemState
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
	validatorVotingPower           *prometheus.Desc
	validatorCommission            *prometheus.Desc
	validatorGasPrice              *prometheus.Desc
	validatorPendingStake          *prometheus.Desc
	validatorPendingWithdrawal     *prometheus.Desc
	validatorStake                 *prometheus.Desc
	validatorDelegationBalance     *prometheus.Desc
	validatorNextEpochStake        *prometheus.Desc
	validatorNextEpochDelegation   *prometheus.Desc
	validatorNextEpochGasPrice     *prometheus.Desc
	validatorNextEpochCommission   *prometheus.Desc
}

func NewExporter(uri string, frequency int) *Exporter {
	return &Exporter{
		state:     nil,
		rpcClient: jsonrpc.NewClient(uri),
		frequency: time.Duration(frequency),
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
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.referenceGasPriceDesc
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.nodeStateMutex.RLock()
	ch <- prometheus.MustNewConstMetric(e.referenceGasPriceDesc, prometheus.GaugeValue, float64(e.state.ReferenceGasPrice), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.nextEpochReferenceGasPriceDesc, prometheus.GaugeValue, float64(e.nextEpochReferenceGasPrice), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.epochStartTimestampMs, prometheus.GaugeValue, float64(e.state.EpochStartTimestampMs), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.storageFundBalance, prometheus.GaugeValue, float64(e.state.StorageFundBalance.Value), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.stakeSubsidyBalance, prometheus.GaugeValue, float64(e.state.StakeSubsidy.Balance.Value), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.stakeSubsidyCurrentEpochAmount, prometheus.GaugeValue, float64(e.state.StakeSubsidy.CurrentEpochAmount), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.totalValidatorStake, prometheus.GaugeValue, float64(e.state.Validators.ValidatorStake), e.epoch)
	ch <- prometheus.MustNewConstMetric(e.totalDelegationStake, prometheus.GaugeValue, float64(e.state.Validators.DelegationStake), e.epoch)

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
	}

	e.nodeStateMutex.RUnlock()
}

type GasQuote struct {
	quote uint64
	stake uint64
}

func (e *Exporter) WatchState() {
	ticker := time.NewTicker(e.frequency * time.Second)

	for {
		e.nodeStateMutex.Lock()

		err := e.rpcClient.CallFor(context.Background(), &e.state, "sui_getSuiSystemState")
		if err != nil {
			log.Printf("Error getting node state: %v", err)
		}

		var gasQuotes []GasQuote
		var totalStake uint64
		for _, validator := range e.state.Validators.ActiveValidators {
			nextEpochStake := validator.Metadata.NextEpochStake + validator.Metadata.NextEpochDelegation
			gasQuote := GasQuote{
				quote: validator.Metadata.NextEpochGasPrice,
				stake: nextEpochStake,
			}

			gasQuotes = append(gasQuotes, gasQuote)
			totalStake += nextEpochStake
		}

		// We count 2/3 by stake weight
		var cumulativeStake uint64
		countedStake := 2.0 / 3.0 * float64(totalStake)
		sort.Slice(gasQuotes, func(i, j int) bool { return gasQuotes[i].quote < gasQuotes[j].quote })
		for _, quote := range gasQuotes {
			cumulativeStake += quote.stake
			if float64(cumulativeStake) >= countedStake {
				e.nextEpochReferenceGasPrice = quote.quote
				break
			}
		}

		// Store epoch as string for prometheus labels
		e.epoch = strconv.Itoa(int(e.state.Epoch))

		e.nodeStateMutex.Unlock()
		<-ticker.C
	}
}
