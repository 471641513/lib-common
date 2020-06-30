#!/usr/bin/env bash
protoc -I. --gofast_out=plugins=grpc:./test_proto test.proto
