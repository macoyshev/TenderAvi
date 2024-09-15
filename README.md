# Запуск сервиса
Необходимые переменные окружения
- `SERVER_ADDRESS`
- `POSTGRES_CONN`

### Запуск веб-сервера в контейнере
Для запуска сервиса в докер контейнере передайте необходимые переменные через флаг -e или создайте .env файл с необходимыми переменными.
Через флаг:
```
$ docker build -t avi-image .
$ docker run --name avi-container -e SERVER_ADDRESS=<...> POSTGRES_CONN=<...> -p <host-port>:<server-port> avi-image
```
`server-port` должен совпадать с портом в `SERVER_ADDRESS`

Через .evn файл:
```
$ docker build -t avi-image .
$ docker run --name avi-container --env-file .env -p <host-port>:<server-port> avi-image
```

## Другие способы запуска
### Запуск веб-сервера
Проверьте наличие необходимых переменных окружения и выполните:
```
$ go mod download
$ go run ./cmd
```

### Запуск веб-сервера через Makefile
Для работы корректного запуска через make нужно указать имя .env файла в переменной `ENVFILE` в makefile по умолчанию используется .env.template. Запуск производится не в контейнере:
```
$ make run
```
