---
marp: true
theme: default
size: 16:9
paginate: true
---

# Горутины и синхронизация в Go
## Занятие 8

---

# Цель занятия

Научиться писать безопасный конкурентный код в Go:
- запускать параллельные задачи через `goroutine`
- синхронизировать доступ к данным
- корректно обрабатывать ошибки и отмену через `errgroup`


---

# Что такое горутина

`goroutine` — лёгкий поток выполнения, управляемый runtime Go.

```go
go func() {
    // параллельная работа
}()
```

Свойства:
- дешёвый запуск
- стартовый стек: примерно `2 KB` (растёт динамически)
- multiplexing на ограниченном числе OS thread

---

# Память и конкурентность

Если две горутины работают с общим mutable state
без синхронизации, поведение не определено.

В Go это проявляется как `data race`.

Безопасность строится на:
- каналах (передача владения данными)
- блокировках (`Mutex`/`RWMutex`)
- атомиках (`atomic`)

---

# Типовые классы ошибок

- `data race`: одновременный небезопасный доступ к данным
- `deadlock`: все горутины ждут друг друга
- `goroutine leak`: фоновые горутины не завершаются

---

# `WaitGroup`: базовая координация

```go
var wg sync.WaitGroup

for i := 0; i < 3; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        fmt.Println("worker", id)
    }(i)
}

wg.Wait()
```

`WaitGroup` отвечает только за ожидание завершения.

---

# Ошибки при `WaitGroup`

Частые проблемы:
- `Add` внутри горутины
- `Add` после запуска части горутин (гонка между `Add` и `Wait`)
- забыли `Add(1)` перед `go` (получаем отрицательный счетчик на `Done`)
- забыли `Done`
- лишний `Done()` (panic: negative WaitGroup counter)
- копирование `WaitGroup` по значению

Паттерн:
- `Add` до `go`
- `defer Done()` сразу в начале горутины
- не вызывать новый `Add`, пока предыдущий `Wait` еще активен

---

# Каналы: полный пример producer-consumer

```go
jobs := make(chan int, 4)

var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    for j := range jobs {
        fmt.Println("process", j)
    }
}()

for i := 1; i <= 5; i++ {
    jobs <- i
}
close(jobs)
wg.Wait()
```

---

# Что важно в этом примере

- producer владеет `close(jobs)`
- consumer читает через `range jobs` до закрытия
- `WaitGroup` гарантирует завершение consumer
- буфер канала сглаживает неравномерность producer/consumer

---

# `select`: timeout и cancel

```go
select {
case msg := <-dataCh:
    _ = msg
case <-time.After(200 * time.Millisecond):
    return errors.New("timeout")
case <-ctx.Done():
    return ctx.Err()
}
```

`select` нужен, когда у операции несколько условий завершения.

---

# Анти-паттерн: вечная горутина

```go
go func() {
    for {
        item := <-in
        handle(item)
    }
}()
```

Проблема: нет пути завершения.

Решение:
- `range in` + `close`
- или `select` с `ctx.Done()`

---

# `Mutex`: защита составного состояния

```go
type Counter struct {
    mu sync.Mutex
    m  map[string]int
}

func (c *Counter) Inc(key string) {
    c.mu.Lock()
    c.m[key]++
    c.mu.Unlock()
}
```

Используй для map/slice/struct с инвариантами.

---

# `RWMutex`: когда нужен

`RLock` допускает много читателей,
`Lock` — эксклюзивно для записи.

Подходит, если:
- чтений значительно больше записей
- критические секции короткие

---

# `atomic`: для простых скаляров

```go
var requests int64
atomic.AddInt64(&requests, 1)
```

Хорошо для:
- счётчиков
- флагов состояния

Плохо для сложных инвариантов между несколькими полями.

---

# Как выбирать примитив

- Нужен обмен данными и сигналами -> `channel`
- Нужна защита общей структуры -> `Mutex`/`RWMutex`
- Нужна простая lock-free операция -> `atomic`
- Нужен parallel fan-out с ошибками/cancel -> `errgroup`

---

# `errgroup`: формальная модель

`errgroup.Group`:
- запускает набор функций `func() error`
- ждёт завершения всех
- возвращает первую ненулевую ошибку из `Wait()`

`errgroup.WithContext(ctx)` дополнительно:
- отдаёт дочерний context
- отменяет его при первой ошибке или после `Wait()`

---

# `errgroup.WithContext`: базовый пример

```go
func RunAll(ctx context.Context, tasks []Task) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, t := range tasks {
        g.Go(func() error {
            return t.Run(ctx)
        })
    }

    return g.Wait()
}
```

---

# Требование к задачам в `errgroup`

Чтобы отмена работала корректно, каждая задача обязана:
- принимать `ctx`
- завершаться при `<-ctx.Done()`
- использовать context-aware I/O API

Иначе `g.Wait()` может ждать дольше ожидаемого.

---

# Паттерн `fan-out`

`fan-out` = один входной поток задач -> несколько workers.

Цели:
- повысить throughput
- ограничить параллелизм контролируемым числом workers

Схема:
- producer пишет в `jobs`
- `N` workers читают из `jobs`
- каждый worker публикует результат в `results`

---

# `fan-out`: пример

```go
jobs := make(chan int)
results := make(chan int)

for w := 0; w < 3; w++ {
    go func() {
        for j := range jobs {
            results <- j * j
        }
    }()
}
```

Producer:
- отправляет задачи в `jobs`
- закрывает `jobs` когда задачи закончились

---

# Паттерн `fan-in`

`fan-in` = несколько источников -> один объединённый поток.

Цели:
- собрать результаты из нескольких каналов
- отдать единый output downstream-обработчику

Ключевая задача:
- корректно закрыть итоговый канал ровно один раз

---

# `fan-in`: пример merge

```go
func merge(cs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    wg.Add(len(cs))
    for _, c := range cs {
        go func() {
            defer wg.Done()
            for v := range c {
                out <- v
            }
        }()
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

---

# `fan-out + fan-in` вместе

Композиция pipeline:
- входные данные -> `fan-out` workers
- workers вычисляют независимо
- `fan-in` merge собирает единый поток

Плюсы:
- горизонтальное масштабирование обработки
- чистая декомпозиция этапов

---

# Граница ответственности в pipeline

- Кто создаёт канал, обычно тот и закрывает
- Владелец `out` отвечает за `close(out)`
- У каждого этапа должен быть shutdown path (`ctx.Done()`)
- Ошибка в одном этапе должна доходить до оркестратора (`errgroup`)

---

# Формальный шаблон pipeline с `errgroup`

- Оркестратор: `g, ctx := errgroup.WithContext(parent)`
- Stage A: producer, пишет в `jobs`, закрывает `jobs`
- Stage B: workers, читают `jobs`, пишут в `results`
- Stage C: merger, закрывает `results` после завершения workers
- В `g.Wait()` возвращаем ошибку пайплайна

