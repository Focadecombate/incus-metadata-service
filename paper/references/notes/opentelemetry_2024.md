# OpenTelemetry — An observability framework for cloud-native software

**Author:** Cloud Native Computing Foundation (CNCF)
**Year:** 2024
**Source:** Official project page
**URL:** https://opentelemetry.io/

## Summary

OpenTelemetry is a CNCF observability framework that provides APIs, SDKs, and tools for instrumenting, generating, collecting, and exporting telemetry data (metrics, logs, traces). Formed from the merger of OpenTracing and OpenCensus. Vendor-neutral and supports multiple backends.

## Key Findings

- CNCF Incubating project — second most active CNCF project after Kubernetes
- Three pillars: metrics, logs, traces (your service uses metrics)
- Vendor-neutral: export to Prometheus, Jaeger, Grafana, Datadog, etc.
- Go SDK provides meter, histogram, counter, and gauge instruments
- Prometheus exporter allows serving metrics at `/metrics` endpoint

## Pertinent Information for TCC

- Cited in the proposal section (technologies) for the observability stack
- Your service uses OTel SDK to instrument event processing with histograms and counters
- Metrics exported in Prometheus format at `/metrics` endpoint
- Demonstrates production-readiness and cloud-native practices in your implementation

## BibTeX Key

`opentelemetry2024`
