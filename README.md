# Запуск сервиса
Необходимые переменные окружения
- `SERVER_ADDRESS`
- `POSTGRES_CONN`

### Запуск веб-сервера
```
make run
```
### Запуск веб-сервера в контейнере
```
make build
make run-c
```
### Запуск тестовой бд
```
make create-init-mock-db
```