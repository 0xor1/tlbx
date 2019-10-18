#!/bin/bash
go fmt ./... && \
go test -coverprofile=coverage.out ./pkg/... && \
go tool cover -html=coverage.out && \
rm coverage.out