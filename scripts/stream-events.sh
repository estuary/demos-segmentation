#!/bin/bash

# Pipe CSV rows into Flow's CSV WebSocket ingestion API.
go run event-generator/generate.go \
    | pv --line-mode --quiet --rate-limit 1000 \
    | websocat --protocol json/v1 ws://localhost:8080/ingest/examples/segment/events \
    > /dev/null
