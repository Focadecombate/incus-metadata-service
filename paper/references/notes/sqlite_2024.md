# SQLite

**Author:** Hipp, D. Richard
**Year:** 2024
**Source:** Official project page
**URL:** https://www.sqlite.org/

## Summary

SQLite is a self-contained, serverless, zero-configuration, transactional SQL database engine. Most widely deployed database in the world (estimated trillions of active installations). Stored in a single disk file, requires no separate server process, and supports most of SQL-92.

## Key Findings

- Embedded database — no separate server process, single file storage
- ACID compliant with WAL (Write-Ahead Logging) mode for concurrent reads
- Write serialization: only one writer at a time (key limitation for high-concurrency writes)
- Ideal for edge deployments and applications with simplified operational requirements
- Used by major software: Android, iOS, Chrome, Firefox, macOS, Windows 10

## Pertinent Information for TCC

- Cited in the proposal section as the storage layer choice
- Paired with SQLC for type-safe generated queries — eliminates ORM overhead
- Write serialization is acknowledged as a limitation in the conclusion section
- The serverless nature aligns with your service's goal of simplified deployment (no external DB process)

## BibTeX Key

`sqlite2024`
