# SUI Prometheus Exporter

Exports various details that are interesting for SUI validators. It reads chain state over Sui gRPC (`sui.rpc.v2.LedgerService`): `GetEpoch` with the system state, and the latest `GetCheckpoint`.

To build:

```
go build -o sui-exporter ./cmd/sui-exporter
```

To run:

```
sui-exporter -grpc mainnet.sui.rpcpool.com:443 -addr :9030
```

| Flag | Default | Meaning |
|---|---|---|
| `-grpc` | (required) | Sui gRPC address as `host:port` |
| `-grpc-tls` | `true` | Use TLS. Set `false` for a plaintext node port such as `127.0.0.1:9000` |
| `-grpc-token` | | Optional `x-token` header for token-protected endpoints |
| `-addr` | `:8080` | Listen address for `/metrics` |
| `-frequency` | `120` | Seconds between updates |
| `-timeout` | `10` | gRPC timeout in seconds |

To run it as a service, put it in a systemd unit:

```
[Unit]
Description=Simple SUI prometheus exporter

[Service]
Type=simple
ExecStart=/usr/local/bin/sui-exporter -grpc mainnet.sui.rpcpool.com:443
KillMode=process
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

# Metrics exported

## Base metrics

Details about current SUI state like gas price, epoch starts, stake and transaction number.

## Checkpoint metrics

Details about the latest checkpoint, including computation cost and storage cost.

Disable with `-export-checkpoint-metrics=false`

## Validator metrics

Details about the validators on their network such as their gas price survey items and compute costs as well as stake balances.

Disable with `-export-validator-metrics=false`

## Validator reports

`sui_validator_report` has one series per report record: the reported validator (`address`, `name`) and the reporter (`reporter`, `reporter_name`).

Disable with `-export-validator-reports=false`

## Exporter health

`sui_exporter_last_success_timestamp_seconds` is the time of the last successful `GetEpoch`. `sui_exporter_errors_total{call}` counts failed calls. When `GetEpoch` fails, the chain metrics disappear until the next success, so they never show stale values.

# Protos

`proto/` holds the `sui/rpc/v2` protos from [MystenLabs/sui-apis](https://github.com/MystenLabs/sui-apis) (Apache-2.0, see `proto/LICENSE`) at the commit in `Makefile`. `internal/suirpc` is generated from them. To update, change `SUI_APIS_REF`, then run `make proto-fetch proto`. `make proto` needs `protoc`, `protoc-gen-go` and `protoc-gen-go-grpc`.

# License

MIT

# Attribution

This software was developed by Triton One (https://triton.one). Patches, suggestions and improvements are always welcome.
