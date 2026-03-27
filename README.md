# Telegram Service
Сервис устанавливает и поддерживает несколько независимых соединений с Telegram через библиотеку [gotd](https://github.com/gotd/td)

[Старт](https://github.com/andreyxaxa/Telegram-Service/tree/main?tab=readme-ov-file#%D0%B7%D0%B0%D0%BF%D1%83%D1%81%D0%BA)

## Обзор / Описание архитектурных решений

Сервис позволяет:

- Динамически создавать и удалять соединения;
- Отправлять текстовые сообщения;
- Получать текстовые сообщения.

Взаимодействие с сервисом происходит через gRPC. Соединения  изолированы друг от друга. 
Проблемы с одним из соединений не влияют на работоспособность сервиса.

- Конфиг - [config/config.go](https://github.com/andreyxaxa/Telegram-Service/blob/main/config/config.go). Читается из `.env` файла.
- Логгер - [pkg/logger/logger.go](https://github.com/andreyxaxa/Telegram-Service/blob/main/pkg/logger/logger.go).
- Хранилище peer'ов - [pkg/peerstorage/peerstorage.go](https://github.com/andreyxaxa/Telegram-Service/blob/main/pkg/peerstorage/storage.go). Просто имплементирован интерфейс `peers.Storage`, только нужные методы, в остальных заглушки.
- Конфигурация grpc сервера - [pkg/grpcserver](https://github.com/andreyxaxa/Telegram-Service/tree/main/pkg/grpcserver)
  Позволяет конфигурировать сервер таким образом:
  ```go
  grpcServer := grpcserver.New(grpcserver.Port(cfg.GRPC.Port))
  ```
- Версионирование API - [internal/controller/grpc/v1](https://github.com/andreyxaxa/Telegram-Service/tree/main/internal/controller/grpc/v1)
  Для версии v2 нужно будет просто добавить папку `grpc/v2` с таким же содержимым, в файле [internal/controller/grpc/router.go](https://github.com/andreyxaxa/Telegram-Service/blob/main/internal/controller/grpc/router.go) добавить строку:
  ```go
  {
    v1.NewTelegramRoutes(app, t, l)
  }

  {
    v2.NewTelegramRoutes(app, t, l) // добавить
  }
  ```
- Структура сессии:
  ```go
  type Session struct {
	  ID          string
	  Client      *telegram.Client  
	  PeerManager *peers.Manager     // внутри хранилище - запоминаем, с кем уже общались, исключаем лишние сетевые запросы
	  Cancel      context.CancelFunc // функция для отмены контекста(он передается в client.Run())

	  IncomingMessages chan Message  // канал для стриминга сообщений 
	  Authorized       atomic.Bool
  }
  ```

## Структура проекта

```
.
├── cmd/                         # Entry point приложения, чтение конфига + запуск
├── config/                      # Конфиг
├── docs/proto/v1                # .proto + source
├── internal/                    # Внутренняя логика приложения
│   ├── app/telegram-service     # Инициализация и запуск всех компонентов в функции Run(она будет вызвана в cmd/)
│   ├── controller/grpc/         # Слой хендлеров сервера
│   │   ├── v1/                  # Структура контроллера, реализация методов TelegramService
│   │   └── router.go            # Создание роутера
│   ├── entity/                  # Сущности бизнес-логики
│   ├── repo/                    # Слой репозиториев(хранилища)
│   │   └── session/inmemory/    # Inmemory реализация (map + RWMutex)
│   └── usecase/                 # Бизнес-логика
│       └── telegram/            # CreateSession, DeleteSession, SendMessage, SubscribeMessages
├── .dockerignore
├── .env.example
├── .gitignore                
├── Dockerfile
├── Makefile
├── docker-compose.yml    
├── go.mod
├── go.sum
└── README.md
```

## Запуск

1. Клонируйте репозиторий
2. В корне создайте `.env` файл, скопируйте туда содержимое [env.example](https://github.com/andreyxaxa/Telegram-Service/blob/main/.env.example):
   ```
   cp .env.example .env
   ```
3. Измените значения переменных `TG_APP_ID` и `TG_APP_HASH`, полученные из https://my.telegram.org/
4. Запустите сервис:
   ```
   make compose-up
   ```

## Note

Использование [gotd](https://github.com/gotd/td) требует создания credentials на https://my.telegram.org/
В связи с ограничениями работы telegram в России, и особеностями создания этих секретов на https://my.telegram.org/, полное e2e тестирование у меня провести не удалось.
На практике сервис запустился, grpc ручки отвечают.

## Примеры вызова API с grpcurl

#### CreateSession
Request:
```bash
grpcurl -plaintext localhost:8081 pact.telegram.TelegramService/CreateSession
```
Response:
```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "qr_code": "tg://login?token=abc123..."
}
```
Откройте Telegram -> Settings -> Devices -> Scan QR и отсканируйте QR-код, сгенерированный из `qr_code` URL.

---

#### SubscribeMessages
Request:
```bash
grpcurl -plaintext \
  -d '{"session_id": "550e8400-e29b-41d4-a716-446655440000"}' \
  localhost:8081 pact.telegram.TelegramService/SubscribeMessages
```

Response(stream):
```json
{
  "message_id": 123457,
  "from": "someuser",
  "text": "Hey!",
  "timestamp": 1711555200
}
```

---

#### SendMessage
Request:
```bash
grpcurl -plaintext \
  -d '{"session_id": "550e8400-e29b-41d4-a716-446655440000", "peer": "@username", "text": "Hello!"}' \
  localhost:8081 pact.telegram.TelegramService/SendMessage
```

Response:
```json
{
  "message_id": 123456
}
```

---

#### DeleteSession
Request:
```bash
grpcurl -plaintext \
  -d '{"session_id": "550e8400-e29b-41d4-a716-446655440000"}' \
  localhost:8081 pact.telegram.TelegramService/DeleteSession
```

## Прочие `make` команды
Генерация исходников из .proto:
```
make proto-v1
```
docker compose down -v:
```
make compose-down
```
Зависимости:
```
make deps
```
