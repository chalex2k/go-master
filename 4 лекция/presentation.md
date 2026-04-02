---
marp: true
theme: default
size: 16:9
paginate: true
---

# ООП в Go
## Занятие 4

---

![width:900px](Screenshot%20from%202026-03-29%2023-50-43.png)

> Джон Боднер, "Go идиомы и паттерны проектирования"

---

# Встраивание (Embedding)

---

# Композиция вместо наследования

> "Отдавайте предпочтение объектной композиции, а не наследованию классов"
> — «Банда четырёх», «Паттерны проектирования», 1994

Go не имеет наследования, но поддерживает композицию через встраивание.

---

# Встраивание структур

```go
type Employee struct {
    Name string
    ID   string
}

func (e Employee) Description() string {
    return fmt.Sprintf("%s (%s)", e.Name, e.ID)
}

type Manager struct {
    Employee        // встроенное поле (без имени)
    Reports []Employee
}
```

---

# Использование встроенных полей

```go
m := Manager{
    Employee: Employee{
        Name: "Bob Bobson",
        ID:   "12345",
    },
    Reports: []Employee{},
}

fmt.Println(m.ID)          // 12345 — поле Employee доступно напрямую
fmt.Println(m.Description()) // "Bob Bobson (12345)" — метод тоже
```

---

# Что даёт встраивание

- Поля встроенного типа становятся полями внешней структуры
- Методы встроенного типа становятся методами внешней структуры
- Можно обращаться как напрямую, так и через имя типа

```go
fmt.Println(m.ID)         // напрямую
fmt.Println(m.Employee.ID) // через тип
```

---

# Затенение полей

Если внешняя структура имеет поле с тем же именем:

```go
type Inner struct {
    X int
}

type Outer struct {
    Inner
    X int  // затеняет Inner.X
}

o := Outer{Inner: Inner{X: 10}, X: 20}
fmt.Println(o.X)        // 20
fmt.Println(o.Inner.X)  // 10
```

---

# Встраивание — это не наследование

**Нельзя:**
- Присвоить Manager переменной типа Employee
- Использовать Manager как Employee

```go
var eFail Employee = m        // ОШИБКА компиляции!
var eOK Employee = m.Employee // OK — явно извлекаем поле
```

---

# Нет динамической диспетчеризации

Методы встроенного поля не знают о вмещающей структуре:

```go
type Inner struct{ A int }

func (i Inner) IntPrinter(val int) string {
    return fmt.Sprintf("Inner: %d", val)
}

func (i Inner) Double() string {
    return i.IntPrinter(i.A * 2)  // всегда вызывает Inner.IntPrinter
}

type Outer struct {
    Inner
    S string
}

func (o Outer) IntPrinter(val int) string {
    return fmt.Sprintf("Outer: %d", val)
}
```

---

# Пример динамической диспетчеризации

```go
o := Outer{
    Inner: Inner{A: 10},
    S:     "Hello",
}
fmt.Println(o.Double())  // "Inner: 20"
// Double() вызывает Inner.IntPrinter, не Outer.IntPrinter
```

> Встраивание — это делегирование, а не полиморфизм.

---

# Встраивание и интерфейсы

Встраивание позволяет вмещающей структуре реализовать интерфейс:

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type MyReader struct {
    bytes.Buffer  // встраиваем Buffer, у которого есть Read
}

// MyReader автоматически реализует Reader!
```

---

# Ваши вопросы?

---

# Интерфейсы

---

# Что такое интерфейс?

**Интерфейс** — набор методов, которые должен реализовать тип.

```go
type Stringer interface {
    String() string
}
```

**Имена интерфейсов** обычно оканчиваются на `-er`: `Reader`, `Writer`, `Closer`, `Handler`.

---

# Неявная реализация

В Go интерфейсы реализуются неявно. Тип не объявляет `implements`.

```go
type User struct {
    Name string
    Age  int
}

// Просто реализуем метод String
func (u User) String() string {
    return fmt.Sprintf("%s (%d)", u.Name, u.Age)
}

// User автоматически реализует Stringer!
var s fmt.Stringer = User{Name: "Alice", Age: 30}
fmt.Println(s)  // "Alice (30)"
```

---

# Типобезопасная утиная типизация

**В динамических языках (Python, Ruby):**
- Утиная типизация: "если ходит как утка..."
- Проблема: нет явных объявлений, сложно понять зависимости

**В Java:**
- Явное `implements` — жесткая привязка

**В Go:**
- Интерфейс определяет вызывающая сторона (клиент)
- Конкретный тип ничего не знает об интерфейсе
- Гибкость + типобезопасность

---

# Пример: определение интерфейса клиентом

```go
// Пакет client определяет, что ему нужно
type Logic interface {
    Process(data string) string
}

type Client struct {
    L Logic
}

// Пакет provider реализует функциональность
type LogicProvider struct{}

func (lp LogicProvider) Process(data string) string {
    return "processed: " + data
}

// В main:
c := Client{L: LogicProvider{}}  // работает!
```

---

# Маленькие интерфейсы — лучшие

Принцип: 1-3 метода, не больше.

```go
// Хорошо
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}
```

---

# Композиция интерфейсов

```go
type ReadWriter interface {
    Reader
    Writer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

> Комбинируйте маленькие интерфейсы вместо создания больших.

---

# Антипаттерн: большой интерфейс

```go
// Плохо
type Database interface {
    Create(task Task) error
    Get(id string) (Task, error)
    Update(task Task) error
    Delete(id string) error
    List() ([]Task, error)
    Close() error
    Ping() error
    Migrate() error
    // ... ещё 20 методов
}
```

---

# Интерфейсы в стандартной библиотеке

```go
func Fprintf(w io.Writer, format string, a ...any) (n int, err error)
```

`io.Writer` — интерфейс с одним методом. Работает с:
- `*os.File` (файл)
- `bytes.Buffer` (буфер в памяти)
- `net.Conn` (сетевое соединение)
- `strings.Builder` (строитель строк)

---

# Пустой интерфейс

`interface{}` или `any` (Go 1.18+) — может содержать любое значение.

```go
var anything any

anything = 42
anything = "hello"
anything = User{Name: "Alice"}
anything = func(x int) int { return x * 2 }
```

> Используйте `any` редко. Go — язык со строгой типизацией.

---

# Проверка реализации интерфейса

```go
type MyWriter struct{}

func (m MyWriter) Write(p []byte) (int, error) {
    return len(p), nil
}

// Проверка на этапе компиляции
var _ io.Writer = (*MyWriter)(nil)  // OK
var _ io.Writer = MyWriter{}        // OK
```

Если тип не реализует интерфейс — ошибка компиляции.

---

# Утверждение типа (Type Assertion)

```go
var w io.Writer = os.Stdout

// С проверкой (безопасно)
if closer, ok := w.(io.Closer); ok {
    closer.Close()
}

// Без проверки (паника, если не тот тип)
closer := w.(io.Closer)  // ОПАСНО!
```

---

# Переключатель типа (Type Switch)

```go
func describe(i any) {
    switch v := i.(type) {
    case int:
        fmt.Printf("Целое: %d\n", v)
    case string:
        fmt.Printf("Строка: %s\n", v)
    case User:
        fmt.Printf("Пользователь: %s\n", v.Name)
    default:
        fmt.Printf("Неизвестный тип: %T\n", v)
    }
}
```

---

# Интерфейсы и nil

Интерфейс = два указателя (тип + значение).

```go
var w io.Writer           // nil (тип=nil, значение=nil)

var b *bytes.Buffer = nil // конкретный тип, nil-значение
var w2 io.Writer = b      // тип=*bytes.Buffer, значение=nil

fmt.Println(w == nil)   // true
fmt.Println(w2 == nil)  // false! (тип не nil)
```

> Интерфейс равен `nil` только когда оба указателя `nil`.

---

# Сравнение интерфейсов

Два интерфейса равны, если равны их типы и значения.

```go
var w1, w2 io.Writer
w1 = os.Stdout
w2 = os.Stdout
fmt.Println(w1 == w2)  // true

// ОСТОРОЖНО: если динамический тип несравним — паника!
var m1, m2 any
m1 = []int{1, 2, 3}
m2 = []int{1, 2, 3}
// fmt.Println(m1 == m2)  // ПАНИКА: срезы несравнимы
```

---

# Ваши вопросы?

---

# Задание
## Время: 10 минут

Откройте https://go.dev/tour/methods/18 и исправьте тип `IPAddr` так, чтобы он реализовывал интерфейс `Stringer`.

---

# Практические советы

---

# Принимайте интерфейсы, возвращайте структуры

```go
// Хорошо
func NewService(storage Storage) *Service {  // принимаем интерфейс
    return &Service{storage: storage}
}

func (s *Service) GetTask(id string) Task {  // возвращаем структуру
    return Task{ID: id}
}
```

**Почему:**
- Приём интерфейса делает код гибким
- Возврат структуры позволяет добавлять методы без ломания API
- Добавление метода в интерфейс ломает все реализации

---

# Где определять интерфейс?

**Правило:** Интерфейс определяет потребитель, а не поставщик.

```go
// Плохо: интерфейс рядом с реализацией
package postgres

type Storage interface { ... }  // зачем? мы же реализация
type PostgresDB struct {}

// Хорошо: интерфейс там, где используется
package task

type TaskStorage interface {
    Save(task Task) error
    Find(id string) (Task, error)
}

type Service struct {
    storage TaskStorage  // потребитель определяет, что нужно
}
```

---

# Dependency Inversion Principle

**Принцип:** Модули верхнего уровня не должны зависеть от модулей нижнего уровня. Оба должны зависеть от абстракций.

```
Без DIP:
[Бизнес-логика] → [PostgreSQL]  (жёсткая связь)

С DIP:
[Бизнес-логика] → [Интерфейс Storage] ← [PostgreSQL]
                                 ↑
                            [MongoDB]
                            [Memory]
```

---

# Пример: проблема жёсткой связи

```go
// Плохо: бизнес-логика зависит от конкретной БД
type TaskService struct {
    db *PostgresDB  // конкретный тип
}

func (s *TaskService) CreateTask(title string) error {
    task := Task{ID: uuid.New(), Title: title}
    return s.db.Insert(task)  // привязаны к Postgres
}
```

**Проблемы:**
- Нельзя заменить PostgreSQL на MongoDB
- Нельзя написать unit-тест без реальной БД

---

# Пример: решение через интерфейс

```go
// Определяем абстракцию
type TaskStorage interface {
    Save(task Task) error
    FindByID(id string) (Task, error)
    Delete(id string) error
    List() ([]Task, error)
}

// Сервис зависит от интерфейса
type TaskService struct {
    storage TaskStorage  // интерфейс!
}

func (s *TaskService) CreateTask(title string) error {
    task := Task{ID: uuid.New(), Title: title}
    return s.storage.Save(task)  // не знает, какая БД
}
```

---

# Реализации интерфейса

```go
// PostgreSQL
type PostgresStorage struct {
    conn *sql.DB
}

func (p *PostgresStorage) Save(task Task) error {
    _, err := p.conn.Exec("INSERT INTO tasks...", task.ID, task.Title)
    return err
}

// In-memory (для тестов)
type MemoryStorage struct {
    tasks map[string]Task
    mu    sync.RWMutex
}

func (m *MemoryStorage) Save(task Task) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.tasks[task.ID] = task
    return nil
}
```

---

# Сборка зависимостей (Dependency Injection)

```go
func main() {
    var storage TaskStorage
    
    switch os.Getenv("DB_DRIVER") {
    case "postgres":
        storage = NewPostgresStorage("postgres://...")
    default:
        storage = NewMemoryStorage()  // для разработки
    }
    
    service := TaskService{storage: storage}
    service.CreateTask("Learn DIP")
}
```

> `main` — единственное место, знающее конкретные типы.

---

# Внедрение зависимостей без фреймворков

```go
// Пакет main собирает все зависимости

func main() {
    // 1. Создаём реализации
    logger := NewLogger()
    db := NewDatabase()
    cache := NewCache()
    
    // 2. Внедряем зависимости
    userService := NewUserService(db, logger)
    authService := NewAuthService(db, cache, logger)
    
    // 3. Создаём сервер
    server := NewServer(userService, authService)
    server.Run(":8080")
}
```

---

# Сегодня мы узнали

- **Встраивание** — композиция вместо наследования, делегирование методов
- **Интерфейсы** — неявная реализация, типобезопасная утиная типизация
- **Маленькие интерфейсы** — 1-3 метода, композиция интерфейсов
- **DIP** — зависимости от абстракций, не от конкретных типов
- **DI** — сборка в main, внедрение через конструкторы


