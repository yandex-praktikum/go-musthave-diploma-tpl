#!/bin/sh

printf "Download go...\n\n"

go mod download

printf "Build go...\n\n"

go run ./cmd/server
