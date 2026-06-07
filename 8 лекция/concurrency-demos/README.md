# Concurrency Demos (Lesson 8)

Демонстрации по горутинам и синхронизации с отдельными точками входа.

## Запуск

```bash
cd lesson8/concurrency-demos
```

```bash
go run ./cmd/waitgroup
go run ./cmd/channels
go run ./cmd/select
go run ./cmd/mutex
go run ./cmd/rwmutex
go run ./cmd/atomic
go run ./cmd/errgroup
go run ./cmd/fanout
go run ./cmd/fanin
go run ./cmd/pipeline
```

## Что показывает

- `cmd/waitgroup`: базовое ожидание завершения горутин
- `cmd/channels`: producer-consumer с `close` и `range`
- `cmd/select`: timeout + cancel через `context`
- `cmd/mutex`: защита map через `sync.Mutex`
- `cmd/rwmutex`: множественные читатели + писатель
- `cmd/atomic`: атомарный счетчик
- `cmd/errgroup`: группа задач с общей отменой и ошибкой
- `cmd/fanout`: распределение задач по workers
- `cmd/fanin`: merge нескольких каналов
- `cmd/pipeline`: композиция `fan-out + fan-in` с `context`
