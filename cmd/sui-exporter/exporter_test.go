package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/rpcpool/sui-exporter/internal/suirpc"
)

type fakeLedger struct {
	epoch    *suirpc.Epoch
	epochErr error
}

func (f *fakeLedger) GetEpoch(context.Context, *suirpc.GetEpochRequest, ...grpc.CallOption) (*suirpc.GetEpochResponse, error) {
	if f.epochErr != nil {
		return nil, f.epochErr
	}
	return &suirpc.GetEpochResponse{Epoch: f.epoch}, nil
}

func (f *fakeLedger) GetCheckpoint(context.Context, *suirpc.GetCheckpointRequest, ...grpc.CallOption) (*suirpc.GetCheckpointResponse, error) {
	return &suirpc.GetCheckpointResponse{Checkpoint: &suirpc.Checkpoint{
		SequenceNumber: proto.Uint64(500),
		Summary: &suirpc.CheckpointSummary{
			TotalNetworkTransactions: proto.Uint64(9000),
			Timestamp:                timestamppb.New(time.UnixMilli(1700000000123)),
			EpochRollingGasCostSummary: &suirpc.GasCostSummary{
				ComputationCost: proto.Uint64(1), StorageCost: proto.Uint64(2), StorageRebate: proto.Uint64(3),
			},
		},
	}}, nil
}

func testValidator(addr, name string, power, nextGas, nextStake uint64) *suirpc.Validator {
	return &suirpc.Validator{
		Address:           proto.String(addr),
		Name:              proto.String(name),
		VotingPower:       proto.Uint64(power),
		CommissionRate:    proto.Uint64(400),
		NextEpochGasPrice: proto.Uint64(nextGas),
		NextEpochStake:    proto.Uint64(nextStake),
		StakingPool:       &suirpc.StakingPool{SuiBalance: proto.Uint64(nextStake)},
	}
}

func testEpoch() *suirpc.Epoch {
	return &suirpc.Epoch{
		Epoch:             proto.Uint64(42),
		ReferenceGasPrice: proto.Uint64(100),
		SystemState: &suirpc.SystemState{
			EpochStartTimestampMs: proto.Uint64(1700000000000),
			StorageFund:           &suirpc.StorageFund{NonRefundableBalance: proto.Uint64(7)},
			StakeSubsidy:          &suirpc.StakeSubsidy{Balance: proto.Uint64(1000), CurrentDistributionAmount: proto.Uint64(10)},
			Validators: &suirpc.ValidatorSet{
				TotalStake: proto.Uint64(400),
				ActiveValidators: []*suirpc.Validator{
					testValidator("0xa", "Triton One", 5000, 120, 100),
					testValidator("0xb", "Other", 3000, 90, 100),
					testValidator("0xc", "Third", 2000, 150, 200),
				},
			},
			ValidatorReportRecords: []*suirpc.ValidatorReportRecord{
				{Reported: proto.String("0xa"), Reporters: []string{"0xb", "0xc"}},
			},
		},
	}
}

func TestSnapshotFromEpoch(t *testing.T) {
	s := snapshotFromEpoch(testEpoch())

	if s.epoch != 42 || s.referenceGasPrice != 100 || s.totalStake != 400 || s.stakeSubsidyCurrentEpochAmount != 10 {
		t.Fatalf("unexpected base fields: %+v", s)
	}
	if got := s.validators[2].nextEpochStakeShare; got != 0.5 {
		t.Errorf("next epoch stake share = %v, want 0.5", got)
	}
	// Quotes sorted by price: 90 (3000), 120 (5000) reaches 8000 >= 2/3 of 10000.
	if s.nextEpochReferenceGasPrice != 120 {
		t.Errorf("next epoch reference gas price = %d, want 120", s.nextEpochReferenceGasPrice)
	}
	want := []validatorReport{
		{address: "0xa", name: "Triton One", reporter: "0xb", reporterName: "Other"},
		{address: "0xa", name: "Triton One", reporter: "0xc", reporterName: "Third"},
	}
	if len(s.reports) != len(want) || s.reports[0] != want[0] || s.reports[1] != want[1] {
		t.Errorf("reports = %+v, want %+v", s.reports, want)
	}
}

func TestCollect(t *testing.T) {
	e := NewExporter(&fakeLedger{epoch: testEpoch()}, time.Minute, time.Second, true, true, true)
	e.update()

	expected := `
# HELP sui_validator_report Validator slashing report
# TYPE sui_validator_report counter
sui_validator_report{address="0xa",epoch="42",name="Triton One",reporter="0xb",reporter_name="Other"} 1
sui_validator_report{address="0xa",epoch="42",name="Triton One",reporter="0xc",reporter_name="Third"} 1
# HELP sui_total_transactions_number Total transactions number
# TYPE sui_total_transactions_number gauge
sui_total_transactions_number{epoch="42"} 9000
# HELP sui_checkpoint_timestamp_ms Checkpoint timestamp in milliseconds
# TYPE sui_checkpoint_timestamp_ms gauge
sui_checkpoint_timestamp_ms{epoch="42"} 1.700000000123e+12
`
	if err := testutil.CollectAndCompare(e, strings.NewReader(expected),
		"sui_validator_report", "sui_total_transactions_number", "sui_checkpoint_timestamp_ms"); err != nil {
		t.Error(err)
	}
}

func TestFailedEpochClearsState(t *testing.T) {
	ledger := &fakeLedger{epoch: testEpoch()}
	e := NewExporter(ledger, time.Minute, time.Second, true, true, true)
	e.update()
	if testutil.CollectAndCount(e, "sui_reference_gas_price") != 1 {
		t.Fatal("expected chain metrics after a successful update")
	}

	ledger.epochErr = errors.New("unavailable")
	e.update()
	if n := testutil.CollectAndCount(e, "sui_reference_gas_price"); n != 0 {
		t.Errorf("chain metrics after a failed update = %d, want 0", n)
	}
	if v := testutil.ToFloat64(e.errors.WithLabelValues("GetEpoch")); v != 1 {
		t.Errorf("GetEpoch errors = %v, want 1", v)
	}
}
