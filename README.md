## LoadBalancer & RateLimiter for http-requests
#### Тестовое задание для Cloud.ru Camp.

В проекте реализовано следующее:
- HTTP-сервер на 8080 порту.
- LoadBalancer с 2-мя методами: RR-RoundRobin, LC-LeastConnection
- RateLimiter на основе TokenBucket
- HealthCheck для серверов переадресации
- GraceFull ShutDown
- Базовое логирование
- Файл конфигураций "config.yaml"
- Упаковка решения в Dockerfile & docker-compose
- Интеграционное тестирование
- Ошибки в формате JSON

### Запуск проекта в докере (в корне проекта):
###### флаг -d для запуска докер-контейнера на фоне 
```bash
docker-compose up -d
```
### Запуск проекта без докера (в корне проекта):
###### go mod download - только при первом запуске
```bash
go mod download
go run loadbalancer/cmd
```
### Запуск 3-х фейковых серверов для теста балансировщика (в корне проекта):
Сервера будут иметь следующие адреса:
- http://localhost:8001
- http://localhost:8002
- http://localhost:8003
###### Windows OS:
```bash
.\demo\start_servers_windows.cmd
```
###### Linux OS:
```bash
.\demo\start_servers.sh
```
### Интеграционные тесты (в корне проекта)
Тест включает в себя проверку работы балансировщика, ограничителя и 'gracefull' завершения.
```bash
go test -v ./test/integration/... -tags=integration -timeout=30s
```
### Нагрузочное тестирование Apache Bench (из ../Apache24/bin)
Чтобы выжать из сервера все соки и проверить пропускную способность, отключи 'rate_limit' в config.yaml.
```bash
ab -n 5000 -c 1000 http://localhost:8080/
```

### Ниже пример настроек балансировщика и ограничителя в config.yaml
```yaml
port: 8080 # порт для внешнего доступа к серверу балансировщика
server_shutdown_timeout_sec: 5 # время серверу на выключение в секундах
lb_method: "RR" # LB-Метод работы балансировщика. "RR"-roundRobin, "LC"-leastConnections
backends: # сервера для переадресации (замените на свои, или запустите эти, /demo/start_servers..)
  - http://localhost:8001
  - http://localhost:8002
  - http://localhost:8003
# ниже настройки для ограничителя запросов
rate_limit:
  enabled: true # true|false - включить|выключить ограничитель
  cleanup_interval: 1m # интервал отчистки информации о токенах старых запросов  
  default: # настройки по умолчанию 
    requests_per_sec: 100 # кол-во запросов в секунду
    burst: 200 # кол-во запросов в секунду для резкого скачка
  special_limits: # пример индивидуальных конфигураций для IP
    - ips: ["192.168.1.100", "192.168.1.101"]
      limit:
        requests_per_sec: 50
        burst: 100
    - ips: ["10.0.0.5"]
      limit:
        requests_per_sec: 2
        burst: 5
    #...
```
