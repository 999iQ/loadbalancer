@echo off
echo Starting backend servers...
start cmd /k "go run demo/mock_server.go 8001"
start cmd /k "go run demo/mock_server.go 8002"
start cmd /k "go run demo/mock_server.go 8003"
echo Servers started in separate windows