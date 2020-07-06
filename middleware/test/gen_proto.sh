#!/usr/bin/env bash
protoc -I. --gofast_out=plugins=grpc:./test_proto test.proto
protoc -I. --go_out=plugins=grpc:./common common.proto
