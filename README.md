# Redfish Prometheus Exporter

Exports chassis power metrics from Redfish (validated with Dell PowerEdge XE9680) to Prometheus.

## Features
- Redfish Basic Auth with optional TLS verification skip
- Targets all chassis or a specific chassis
- Metrics:
  - `redfish_power_consumed_watts{chassis,control_name}`
  - `redfish_power_average_watts{chassis,control_name}`
  - `redfish_power_min_watts{chassis,control_name}`
  - `redfish_power_max_watts{chassis,control_name}`

## Install
```bash
# Go 1.21+
go build -o redfish-exporter .
```

## Configuration
You can configure via flags or a YAML file. Flags override config values.

Example `config.yaml`:
```yaml
web:
  listen_address: ":9102"
redfish:
  host: "https://192.168.1.100"
  username: "root"
  password: "password"
  insecure_tls: true
  chassis_id: ""      # optional; if empty, scrape all chassis
  timeout_sec: 10
```

## Run
Using config file:
```bash
./redfish-exporter --config.file=./config.yaml
```

Using flags only:
```bash
./redfish-exporter \
  --web.listen-address=":9102" \
  --redfish.host=https://<redfish-ip-or-host> \
  --redfish.username=<user> \
  --redfish.password=<pass> \
  --redfish.insecure=true \
  --redfish.chassis-id=<optional-chassis-id>
```

Metrics: `http://localhost:9102/metrics`
Health: `http://localhost:9102/healthz`

## Prometheus scrape config
```yaml
- job_name: redfish
  static_configs:
  - targets: ["localhost:9102"]
```

## Notes
- Tested against Dell XE9680 Redfish; other Redfish implementations may vary.
- If your BMC uses self-signed certs, set `--redfish.insecure` or `insecure_tls: true`.

## License
MIT


