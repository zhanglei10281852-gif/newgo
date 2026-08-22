# Geothermal Microseismic Operations

Geothermal Microseismic Operations is a pure Go backend for coordinating enhanced-geothermal field work. Operators register wells and sensor stations, ingest monitoring batches, classify microseismic events, request hydraulic-stimulation permits, and complete independent risk reviews before a job can proceed. State transitions, capacity reservations, idempotency, audit evidence, durable retries, and restart recovery are persisted in SQLite.

The workflow is: register a well, calibrate its stations, open a monitoring batch, classify every event, submit a stimulation permit, and approve it only after risk review. Cross-entity operations run in transactions and every change emits an audit event and a durable notification job.

Run with `go run ./cmd/seed-user` and `go run ./cmd/server`. Health endpoints are `/healthz` and `/readyz`; login is `/api/v1/auth/login`. Build with `docker build --platform linux/amd64 -t geothermal-ops .`.

