#!/bin/bash

go test ./... -race -p 1 -coverprofile=coverage.txt -covermode=atomic -exec sudo ./...
go tool cover -html=coverage.txt -o cover.html
open cover.html
