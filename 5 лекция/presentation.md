---
marp: true
theme: default
size: 16:9
paginate: true
---

# Ошибки, defer, panic в Go
## Занятие 5

---

# План

- Ошибки как значения
- Оборачивание и проверка ошибок
- `defer`: порядок и типичные кейсы
- `panic` и `recover`
- Практические советы и задание

---

# Ошибки в Go

В Go ошибка это обычное значение типа `error`.

```go
type error interface {
    Error() string
}
```

> Ошибки не исключения. Их нужно явно возвращать и обрабатывать.

---

# Базовый паттерн обработки

```go
f, err := os.Open("data.txt")
if err != nil {
    return fmt.Errorf("open data.txt: %w", err)
}
defer f.Close()
```

Правило: после вызова, который может вернуть ошибку, сразу проверяем `err`.

---

# Почему `if err != nil` много раз?

- Явный поток управления
- Прозрачный контракт функции
- Предсказуемое поведение без скрытых переходов

```go
data, err := io.ReadAll(f)
if err != nil {
    return fmt.Errorf("read file: %w", err)
}
```

---

# Создание ошибок

```go
var ErrNotFound = errors.New("not found")

func Find(id string) (Task, error) {
    if id == "" {
        return Task{}, errors.New("empty id")
    }
    return Task{}, ErrNotFound
}
```

`errors.New` полезен для простых статических ошибок.

---

# Форматирование и оборачивание

```go
if err != nil {
    return fmt.Errorf("load config %q: %w", path, err)
}
```

- `%v` просто печатает ошибку
- `%w` оборачивает ошибку, сохраняя цепочку причин

---

# Проверка типа ошибки: `errors.Is`

```go
if errors.Is(err, os.ErrNotExist) {
    // файл не найден
}

if errors.Is(err, ErrNotFound) {
    // доменная ошибка
}
```

`errors.Is` проходит по цепочке обёрнутых ошибок.

---

# Извлечение конкретного типа: `errors.As`

```go
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println("операция:", pathErr.Op)
    fmt.Println("путь:", pathErr.Path)
}
```

Используйте, когда важны дополнительные поля ошибки.

---

# Кастомный тип ошибки

```go
type ValidationError struct {
    Field string
    Msg   string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation %s: %s", e.Field, e.Msg)
}
```

Кастомный тип удобен, когда вызывающему коду нужны детали.

---

# Антипаттерны работы с ошибками

- Игнорирование ошибки: `_ = doSomething()`
- Потеря контекста: `return err` вместо осмысленного wrapping
- Логирование и возврат одной и той же ошибки на каждом уровне

> Обычно ошибку логируют на границе приложения, а внутри только возвращают.

---

# `defer`: отложенный вызов

`defer` откладывает выполнение функции до выхода из текущей функции.

```go
func read(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    return io.ReadAll(f)
}
```

---

# Порядок `defer`

`defer` выполняются в обратном порядке (LIFO).

```go
defer fmt.Println("1")
defer fmt.Println("2")
defer fmt.Println("3")
// вывод: 3, 2, 1
```

Полезно, когда ресурс зависит от другого ресурса.

---

# Аргументы `defer` вычисляются сразу

```go
x := 10
defer fmt.Println(x) // запомнит 10
x = 20
```

Выведется `10`, потому что значение аргумента фиксируется в момент `defer`.

---

# Ошибки при `Close`

```go
func save(path string, data []byte) (err error) {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer func() {
        if cerr := f.Close(); err == nil && cerr != nil {
            err = cerr
        }
    }()

    _, err = f.Write(data)
    return err
}
```

Не теряйте ошибку `Close`, особенно при записи файлов.

---

# Что такое `panic`

`panic` немедленно прерывает обычное выполнение функции.

```go
func mustPositive(n int) {
    if n < 0 {
        panic("negative number")
    }
}
```

Во время паники всё равно выполняются `defer`.

---

# Когда `panic` уместна

- Критическая ошибка инициализации (в `main`/`init`)
- Внутренние helper-функции `Must...`

Не используйте `panic` для обычных бизнес-ошибок.

---

# `recover`: перехват паники

`recover` работает только внутри `defer`.

```go
func safeRun(fn func()) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()
    fn()
    return nil
}
```

---

# Граница восстановления

Рекомендуемый подход в HTTP/worker:
- Внутри бизнес-кода возвращаем `error`
- На верхнем уровне (middleware/goroutine wrapper) делаем `recover`
- Логируем stack trace и возвращаем контролируемый ответ

---

# Пример middleware recovery

```go
func RecoverMW(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                log.Printf("panic: %v", rec)
                http.Error(w, "internal error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

---

# Ошибки против panic

- `error`:
  - ожидаемые сбои (файл не найден, валидация, таймаут)
  - часть бизнес-контракта
- `panic`:
  - нештатная программная ситуация
  - обычно баг или нарушенный инвариант

---

# Практические правила

- Добавляйте контекст к ошибке (`fmt.Errorf(...: %w, err)`)
- Интерфейсы проверяйте через `errors.Is/As`, а не строки
- `defer` ставьте сразу после успешного открытия ресурса
- `panic` не должна покидать границы сервера/воркера

---

# Ваши вопросы?

---


# Юнит и интеграционные тесты

---

# Что разберём

- Unit-тесты в Go (`testing`, `go test`)
- Табличные тесты и подтесты
- `testify` и `httptest`
- Интеграционные тесты и `build tags`

---

# Зачем тесты в проекте

- Проверяют, что код работает как ожидается
- Дают уверенность при рефакторинге
- Упрощают поддержку растущего проекта
- Фиксируют контракт поведения функций и модулей

---

# Пакет `testing` и `go test`

```bash
go test ./...
```

Базовые соглашения:
- Файл теста: `*_test.go`
- Функция: `TestXxx(t *testing.T)`
- Основные методы: `t.Error`, `t.Fatal`, `t.Log`

---

# Простой unit-тест

```go
func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5

    if got != want {
        t.Errorf("Add(2, 3) = %d; want %d", got, want)
    }
}
```

> Unit-тест проверяет небольшой участок логики в изоляции.

---

# Табличные тесты

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 2, 3, 5},
        {"negative", -1, -1, -2},
        {"mixed", 5, -3, 2},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Add(tt.a, tt.b); got != tt.want {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

---

# `testify`: меньше boilerplate

```go
func TestCalculate(t *testing.T) {
    result, err := Calculate(2, 3)
    require.NoError(t, err)
    assert.Equal(t, 5, result)

    result, err = Calculate(5, 0)
    require.Error(t, err)
    assert.Equal(t, 0, result)
}
```

`testify` делает проверки читаемее и короче.

---

# Тестирование HTTP через `httptest`

```go
func TestGreetingHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/greet?name=Alice", nil)
    w := httptest.NewRecorder()

    srv.GreetingHandler(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}
```

Позволяет тестировать обработчики без запуска реального сервера.

---

# Моки и зависимости

Чтобы мокать зависимости, код должен зависеть от интерфейсов:

```go
type Database interface {
    GetUser(id int) (string, error)
}

type Service struct {
    DB Database
}
```

Это ключ к тестируемой архитектуре.

---

# Генерация моков (`mockgen`)

Для сложных интерфейсов моки лучше генерировать автоматически.

```bash
go install go.uber.org/mock/mockgen@latest
mockgen -source=user.go -destination=user_mock.go -package=user
```

- `-source` — файл с интерфейсом
- `-destination` — файл с моком
- `-package` — пакет сгенерированного файла

---

# Пример теста со сгенерированным моком

```go
func TestGetUserName_Success(t *testing.T) {
    ctrl := gomock.NewController(t)
    mockDB := NewMockDatabase(ctrl)
    mockDB.EXPECT().GetUser(1).Return("Alice", nil)

    service := &Service{DB: mockDB}
    name, err := service.GetUserName(1)
    require.NoError(t, err)
    assert.Equal(t, "Alice", name)
}
```

---

# Интеграционные тесты

Интеграционные тесты проверяют взаимодействие нескольких компонентов:

- Сервис + база данных
- Сервис + внешний API
- Несколько внутренних модулей вместе

Они медленнее unit-тестов, но проверяют реальную интеграцию.

---

# `build tags` для интеграционных тестов

```go
//go:build integration

package integration
```

```bash
go test -tags=integration ./...
```

Так можно отделять быстрые unit-тесты от тяжёлых интеграционных.

---

# Test Suite подход

```go
type UnitTestSuite struct {
    suite.Suite
    service *UserService
    mockRepo *MockUserRepository
}

func TestUnitSuite(t *testing.T) {
    suite.Run(t, new(UnitTestSuite))
}
```

Suites помогают переиспользовать setup/teardown и структурировать набор тестов.

---

# Параллельный запуск тестов

Go может запускать тесты параллельно, чтобы ускорить прогон.

```go
func TestSlowA(t *testing.T) {
    t.Parallel()
    // ...
}

func TestSlowB(t *testing.T) {
    t.Parallel()
    // ...
}
```

```bash
go test -parallel=4 ./...
```

---

# Flaky-тесты

Flaky-тесты — недетерминированные тесты: то проходят, то падают при одном и том же коде.

Типичные причины:
- зависимость от времени или случайности
- общие глобальные состояния
- гонки и неявная параллельность

```bash
go test -count=100 ./...
```

---

# Code coverage

Code coverage — метрика, показывающая, какая часть кода была выполнена при запуске тестов.

- Помогает найти непокрытые участки
- Не гарантирует отсутствие багов
- Нужен баланс: метрика + качество сценариев

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

# Практические команды

```bash
go test ./...
go test -v ./...
go test -race ./...
go test -count=100 ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

