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


# License

MIT

# Attribution

This software was developed by Triton One (https://triton.one). Patches, suggestions and improvements are always welcome.
