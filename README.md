<p align="center"><img src="docs/tempo/website/logo_and_name.png" alt="Tempo Logo"></p>

Grafana Tempo is a high volume, minimal dependency distributed tracing backend.  It is supports key/value lookup only and is designed to work in concert with logs and metrics (exemplars) for discovery.

Tempo is Jaeger, Zipkin, OpenCensus and OpenTelemetry compatible.  It ingests batches in any of the mentioned formats, buffers them and then writes them to GCS, S3 or local disk.  As such it is robust, cheap and easy to operate!

## Getting Started

- [Documentation](https://grafana.com/docs/tempo/latest/)
- [Deployment Examples](./example)
  - Deployment and log discovery Examples
- [What is Distributed Tracing?](https://opentracing.io/docs/overview/what-is-tracing/)

## Getting Help

If you have any questions or feedback regarding Tempo:

- Ask a question on the Tempo Slack channel. To invite yourself to the Grafana Slack, visit [https://slack.grafana.com/](https://slack.grafana.com/) and join the #tempo channel.
- [File an issue](https://github.com/grafana/tempo/issues/new/choose) for bugs, issues and feature suggestions.
- UI issues should be filed with [Grafana](https://github.com/grafana/grafana/issues/new/choose).

## OpenTelemetry

Tempo's receiver layer, wire format and storage format are all based directly on [standards](https://github.com/open-telemetry/opentelemetry-proto) and [code](https://github.com/open-telemetry/opentelemetry-collector) established by [OpenTelemetry](https://opentelemetry.io/).  We support open standards at Grafana!

## Other Components

### tempo-query
tempo-query is jaeger-query with a [hashicorp go-plugin](https://github.com/jaegertracing/jaeger/tree/master/plugin/storage/grpc) to support querying Tempo.

### tempo-vulture
tempo-vulture is tempo's bird themed consistency checking tool.  It queries Loki, extracts trace ids and then queries tempo.  It metrics 404s and traces with missing spans.

### tempo-cli
tempo-cli is the place to put any utility functionality related to tempo.

Currently, it supports dumping header information for all blocks from gcs/s3.
```
go run ./cmd/tempo-cli -backend=gcs -bucket ops-tools-tracing-ops -tenant-id single-tenant
```

It also supports connecting to tempo directly to get a trace result in JSON.
```console
$ go run cmd/tempo-cli/main.go -query-endpoint http://localhost:3100 -traceID 2a61c34ff39a1518 -orgID 1
{"batches":[{"resource":{"attributes":[{"key":"service.name","value":{"Value":{"string_value":"cortex-ingester"}}}.....}
```


## TempoDB

[TempoDB](https://github.com/grafana/tempo/tree/master/tempodb) is included in the this repository but is meant to be a stand alone key value database built on top of cloud object storage (gcs/s3).  It is a natively multitenant, supports a WAL and is the storage engine for Tempo.
