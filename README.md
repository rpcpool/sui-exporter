# SUI Prometheus Exporter

Exports various details that are interesting for SUI validators.

To build:

```
go build -o sui-exporter ./cmd/sui-exporter
```

To run, put it in a systemd unit:

```
[Unit]
Description=Simple SUI prometheus exporter

[Service]
Type=simple
ExecStart=/usr/local/bin/sui_exporter
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

Details about checkpoints, including computation cost and storage cost.

Disable with ```-enable-checkpoint-metrics=false```

## Validator metrics

Details about the validators on their network such as their gas price survey items and compute costs as well as stake balances.

Disable with ```-enable-validator-metrics=false```

# License

MIT

# Attribution

This software was developed by Triton One (https://triton.one). Patches, suggestions and improvements are always welcome.
