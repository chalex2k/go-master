---
marp: true
theme: default
size: 16:9
paginate: true
---

# gRPC в Go
## Занятие 7

---

# Цель занятия

Понять, что такое `gRPC`, как он работает внутри,
и как поднять `gRPC`-сервис на Go.

---

# Сегодня разберём

- Что такое `gRPC` и где он применяется
- Как устроен протокол: `HTTP/2` + `Protocol Buffers`
- Контракты через `.proto`
- Типы RPC: unary и streaming
- Как писать `gRPC` server/client на Go
- Ошибки, deadline, metadata, interceptors

---

# Что такое gRPC

`gRPC` — это RPC-фреймворк от Google.

Идея:
- Описываем API в `.proto`
- Генерируем код для клиента и сервера
- Вызываем удалённый метод почти как локальную функцию

---

# gRPC vs REST

`gRPC`:
- Бинарный формат (`protobuf`)
- Жёсткий контракт (schema-first)
- Быстрее и компактнее по сети
- Нативный streaming

`REST`:
- Проще дебажить вручную (`JSON`)
- Удобнее для публичных HTTP API

---

# Как работает gRPC

1. Клиент вызывает метод `Service.Method`
2. Запрос сериализуется в `protobuf`
3. Данные идут по `HTTP/2`
4. Сервер десериализует и вызывает handler
5. Ответ возвращается тем же путём

---

# Почему HTTP/2 важен

`HTTP/2` даёт:
- Мультиплексирование (много запросов в одном соединении)
- Бинарный формат
- Header compression
- Долгоживущие соединения и меньшие задержки

---

# HTTP/1.1 vs HTTP/2

`HTTP/1.1`:
- Обычно 1 запрос на соединение одновременно
- Для параллелизма открывают несколько TCP-соединений
- Head-of-line проблема на уровне очереди запросов
- Заголовки отправляются почти как есть (много повторов)

`HTTP/2`:
- Много stream в одном TCP-соединении
- Фреймы разных запросов перемешиваются (multiplexing)
- Сжатие заголовков (`HPACK`)
- Меньше latency на множестве мелких RPC

---

# Что такое Protocol Buffers

`protobuf` — формат сериализации + язык описания схемы.

Плюсы:
- Компактный бинарный формат
- Быстрый encode/decode
- Кодогенерация для многих языков
- Эволюция схемы при правильных правилах

---

# Структура .proto

```proto
syntax = "proto3";

package task.v1;

option go_package = "example.com/lesson7/gen/task/v1;taskv1";

service TaskService {
  rpc CreateTask(CreateTaskRequest) returns (CreateTaskResponse);
}
```

---

# Сообщения и поля

```proto
message CreateTaskRequest {
  string title = 1;
  string description = 2;
}

message CreateTaskResponse {
  int64 id = 1;
}
```

**Важно:** номера полей (`= 1`, `= 2`) — часть wire-формата.

---

# Правила эволюции схемы

- Не переиспользуйте номера удалённых полей
- Удалённые номера помечайте как `reserved`
- Новые поля добавляйте с новыми номерами
- Не меняйте тип поля без миграционного плана

---

# Типы RPC

- `Unary`: 1 запрос -> 1 ответ
- `Server streaming`: 1 запрос -> поток ответов
- `Client streaming`: поток запросов -> 1 ответ
- `Bidirectional streaming`: поток <-> поток

---

# Unary RPC

```proto
rpc GetTask(GetTaskRequest) returns (GetTaskResponse);
```

Обычный request/response сценарий,
аналог HTTP handler по модели взаимодействия.

---

# Server Streaming RPC

```proto
rpc ListTasks(ListTasksRequest) returns (stream Task);
```

Сервер постепенно отправляет элементы,
клиент читает до `io.EOF`.

---

# Client Streaming RPC

```proto
rpc UploadTasks(stream CreateTaskRequest) returns (UploadTasksResponse);
```

Клиент отправляет много сообщений,
сервер отвечает один раз после `CloseAndRecv`.

---

# Bidirectional Streaming RPC

```proto
rpc TaskEvents(stream TaskEvent) returns (stream TaskEvent);
```

Обе стороны читают и пишут независимо:
чат, live-события, real-time синхронизация.

---

# Go: server streaming handler

```go
func (s *TaskServer) ListTasks(
    req *taskv1.ListTasksRequest,
    stream taskv1.TaskService_ListTasksServer,
) error {
    for _, t := range s.tasks {
        if err := stream.Send(t); err != nil { return err }
    }
    return nil
}
```

---

# Go: client streaming handler

```go
func (s *TaskServer) UploadTasks(
    stream taskv1.TaskService_UploadTasksServer,
) error {
    var count int32
    for {
        req, err := stream.Recv()
        if err == io.EOF { return stream.SendAndClose(&taskv1.UploadTasksResponse{Count: count}) }
        if err != nil { return err }
        _ = req
        count++
    }
}
```

---

# Go: bidirectional streaming handler

```go
func (s *TaskServer) TaskEvents(
    stream taskv1.TaskService_TaskEventsServer,
) error {
    for {
        ev, err := stream.Recv()
        if err == io.EOF { return nil }
        if err != nil { return err }
        if err := stream.Send(&taskv1.TaskEvent{Message: "ack: " + ev.GetMessage()}); err != nil { return err }
    }
}
```

---

# Go: вызовы всех типов на клиенте

```go
// unary
_, _ = client.GetTask(ctx, &taskv1.GetTaskRequest{Id: 1})

// server streaming
ss, _ := client.ListTasks(ctx, &taskv1.ListTasksRequest{})
for { t, err := ss.Recv(); if err == io.EOF { break }; if err != nil { break }; _ = t }

// client streaming
cs, _ := client.UploadTasks(ctx)
_ = cs.Send(&taskv1.CreateTaskRequest{Title: "A"})
_ = cs.Send(&taskv1.CreateTaskRequest{Title: "B"})
_, _ = cs.CloseAndRecv()

// bidirectional streaming
bs, _ := client.TaskEvents(ctx)
_ = bs.Send(&taskv1.TaskEvent{Message: "ping"})
_, _ = bs.Recv()
```

---

# Ваши вопросы?

---

# Генерация кода в Go

Обычно нужны плагины:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

И сам `protoc`.

---

# Команда protoc

```bash
protoc \
  -I api \
  --go_out=. --go_opt=module=example.com/lesson7 \
  --go-grpc_out=. --go-grpc_opt=module=example.com/lesson7 \
  api/task/v1/task.proto
```

---

# Что сгенерируется

- Типы сообщений (`*.pb.go`)
- gRPC-интерфейсы и client stubs (`*_grpc.pb.go`)
- Интерфейс сервера вида `TaskServiceServer`
- Клиент `NewTaskServiceClient(conn)`

---

# Минимальный gRPC server (Go)

```go
lis, err := net.Listen("tcp", ":50051")
if err != nil { log.Fatal(err) }

grpcServer := grpc.NewServer()
taskv1.RegisterTaskServiceServer(grpcServer, &TaskServer{})

if err := grpcServer.Serve(lis); err != nil {
    log.Fatal(err)
}
```

---

# Реализация unary метода

```go
func (s *TaskServer) CreateTask(
    ctx context.Context,
    req *taskv1.CreateTaskRequest,
) (*taskv1.CreateTaskResponse, error) {
    if req.GetTitle() == "" {
        return nil, status.Error(codes.InvalidArgument, "title is required")
    }

    id := int64(42)
    return &taskv1.CreateTaskResponse{Id: id}, nil
}
```

---

# Минимальный gRPC client (Go)

```go
conn, err := grpc.NewClient(
    "localhost:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
)
if err != nil { log.Fatal(err) }
defer conn.Close()

client := taskv1.NewTaskServiceClient(conn)
resp, err := client.CreateTask(ctx, &taskv1.CreateTaskRequest{Title: "Learn gRPC"})
```

---

# Статусы и ошибки

В `gRPC` используем статус-коды:
- `codes.InvalidArgument`
- `codes.NotFound`
- `codes.AlreadyExists`
- `codes.Internal`
- `codes.Unavailable`

Возврат ошибки:

```go
return nil, status.Error(codes.NotFound, "task not found")
```

---

# Deadline и cancellation

Клиент должен ограничивать время вызова:

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

resp, err := client.GetTask(ctx, req)
```

Если deadline истёк -> `codes.DeadlineExceeded`.

---

# Metadata

`Metadata` — пары ключ/значение в запросе/ответе.

Используется для:
- trace-id и correlation-id
- auth token
- служебных флагов

---

# Interceptors

Interceptor — middleware для `gRPC`.

Типовые задачи:
- Логирование
- Метрики
- Трейсинг
- Аутентификация/авторизация

---

# Unary interceptor пример

```go
func LoggingInterceptor(
    ctx context.Context,
    req any,
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (any, error) {
    start := time.Now()
    resp, err := handler(ctx, req)
    log.Printf("method=%s dur=%s err=%v", info.FullMethod, time.Since(start), err)
    return resp, err
}
```

---

# Безопасность

Для production обычно включают `TLS`:
- Сервер с сертификатом
- Клиент с проверкой сертификата сервера
- При необходимости `mTLS` для взаимной аутентификации

Локально для демо часто используют `insecure`.

---

# Версионирование API

Практика:
- Версия в пакете: `task.v1`, `task.v2`
- Не ломать контракты внутри версии
- Депрекейтить методы, а не удалять резко
- Планировать migration window

---

# Ваши вопросы?

---
