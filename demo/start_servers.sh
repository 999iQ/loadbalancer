#!/bin/bash
go run demo/mock_server.go --port=8001 &
go run demo/mock_server.go --port=8002 &
go run demo/mock_server.go --port=8003 &
