#!/usr/bin/env just --justfile

default: build

bin:
  @test -d ./bin || mkdir ./bin

build: bin
  CGO_ENABLED=0 go build -trimpath -o ./bin/run-boinc ./cmd/run-boinc

clean:
  go clean ./...
  rm -rf ./bin
