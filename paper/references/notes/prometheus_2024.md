# Prometheus — Monitoring system & time series database

**Author:** Cloud Native Computing Foundation (CNCF)
**Year:** 2024
**Source:** Official project page
**URL:** https://prometheus.io/

## Summary

Prometheus is a CNCF Graduated monitoring system and time-series database. Uses a pull-based model to scrape metrics from instrumented targets at configurable intervals. Provides PromQL for querying, built-in alerting, and native integration with Grafana for visualization.

## Key Findings

- CNCF Graduated project — production-ready, widely adopted
- Pull-based scraping model: Prometheus scrapes `/metrics` endpoints
- Metric types: counter, gauge, histogram, summary
- PromQL for querying time-series data
- De-facto standard for cloud-native monitoring alongside Grafana

## Pertinent Information for TCC

- Cited alongside OpenTelemetry in the proposal section (technologies)
- Your service exports metrics in Prometheus format via the OTel Prometheus exporter
- Instrumented metrics: `event_processing_duration_seconds` (histogram), `event_processing_total` (counter)
- Enables monitoring of sync performance, error rates, and event processing latency

## BibTeX Key

`prometheus2024`
