#!/bin/bash

GOOS=darwin GOARCH=amd64 go build -o build/krakentrader krakentrader.go
GOOS=linux GOARCH=amd64 go build -o build/krakentrader krakentrader.go
GOOS=windows GOARCH=amd64 go build -o build/krakentrader.exe krakentrader.go
