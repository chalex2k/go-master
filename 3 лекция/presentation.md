---
marp: true
theme: default
size: 16:9
paginate: true
---

# Структуры. Веб сервер.
## Занятие 3

---

# Структуры в Go

---

# Что такое структуры?

**Структура** — составной тип данных, группирующий несколько значений разных типов под одним именем.

```go
type Entry struct {
    Name    string
    Surname string
    Year    int
}
```

**Ключевые особенности:**
- Поля любых типов (включая структуры и срезы)
- Порядок полей важен
- Определяются на уровне пакета
- Заглавная буква = экспортируется

---

# Где определять структуры

```go
// На уровне пакета (доступна всему пакету)
type Person struct {
    FirstName string
    LastName  string
    Age       int
}

func main() {
    // Внутри функции (только здесь)
    type LocalPerson struct {
        Name string
        Age  int
    }
}
```

---

# Создание экземпляров

```go
var fred person          // Все поля = нулевые значения: "", 0, ""

bob := person{}          // Все поля = нулевые значения

beth := person{          // Литерал с именами полей (рекомендуется)
    age:  30,
    name: "Beth",
    // pet не указан → ""
}
```

**Преимущества именованного литерала:**
- Любой порядок полей
- Можно опускать поля
- Самодокументируемый код

---

# Функция-конструктор

```go
func newPerson(name, surname string, year int) *Entry {
    return &Entry{
        Name:    name,
        Surname: surname,
        Year:    year,
    }
}

p := newPerson("Alice", "Smith", 1990)
```

---

# Использование new()

```go
p := new(person)   // Возвращает *person с нулевыми значениями
p.name = "Alice"
p.age = 30
```

> `new(T)` выделяет память и возвращает указатель `*T`

---

# Доступ к полям и указатели

Точечная нотация (`.`) работает одинаково для значений и указателей:

```go
// Через значение
p := person{name: "Bob", age: 50}
fmt.Println(p.name)  // чтение
p.age = 51           // запись

// Через указатель — автоматическое разыменование
ptr := &person{name: "Alice", age: 30}
fmt.Println(ptr.name)  // Go автоматически разыменовывает
ptr.age = 31
```

---

# Передача в функции

```go
// По значению (копия)
func updateByValue(p person) {
    p.age = 100  // оригинал не изменится
}

// По указателю (изменение)
func updateByPtr(p *person) {
    p.age = 100  // оригинал изменится
}
```

---

# Сравнение структур

```go
type Comparable struct {
    Name string
    Age  int
}

c1 := Comparable{Name: "Bob", Age: 30}
c2 := Comparable{Name: "Bob", Age: 30}
fmt.Println(c1 == c2)  // true
```

```go
type NotComparable struct {
    Name    string
    Hobbies []string  // срез несравниваем
}

// nc1 == nc2  // ОШИБКА: нельзя сравнивать
```

---

# Преобразование структур

Можно преобразовать S1 → S2, если:
- Одинаковые имена и типы полей
- Одинаковый порядок полей

```go
type FirstPerson struct {
    name string
    age  int
}

type SecondPerson struct {
    name string
    age  int
}

f := FirstPerson{name: "Bob", age: 50}
s := SecondPerson(f)  // OK
```

---

# Анонимные структуры

Структуры без имени типа:

```go
pet := struct {
    name string
    kind string
}{
    name: "Fido",
    kind: "dog",
}
```

---

# Указатель или значение?

**Возвращайте указатель, если:**
- Нужно изменять структуру
- Состояние должно разделяться
- Структура содержит большие поля (срезы, карты, другие структуры)

**Возвращайте значение, если:**
- Структура небольшая и простая
- Нужна неизменяемость
- Потокобезопасность важна

---

# Методы структур

Метод — функция с получателем (receiver):

```go
type Point struct {
    X, Y float64
}

func (p Point) Distance() float64 {
    return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

func main() {
    p := Point{X: 3, Y: 4}
    fmt.Println(p.Distance())  // 5
}
```

---

# Ваши вопросы?

---

# Задание
## Время: 10 минут

Создайте структуру `Point` с полями `X` и `Y` (float64). Добавьте метод, который считает расстояние от точки до начала координат.

Подробнее о структурах: https://go.dev/tour/moretypes/4

---

# Пакеты и области видимости

---

# Что такое пакет?

**Пакет (package)** — директория с `.go` файлами, объединёнными общим назначением.

```
myapp/
├── main.go           (package main)
├── math/
│   └── math.go       (package math)
└── utils/
    ├── string.go     (package utils)
    └── number.go     (package utils)
```

---

# Ключевые правила пакетов

- Каждый файл начинается с `package <имя>`
- Все файлы в одной папке — один пакет
- Имя папки обычно совпадает с именем пакета
- **package main** — точка входа (исполняемый файл)
- Любой другой пакет — библиотека

---

# Импорт пакетов

```go
import "fmt"
import "math/rand"

// ИЛИ группировка
import (
    "fmt"
    "math/rand"
    "net/http"
    "myapp/utils"
    "github.com/gorilla/mux"
)
```

---

# Как Go ищет пакеты

1. **Стандартная библиотека** → `$GOROOT/src/`
2. **Пакеты из модуля** → `$GOPATH/pkg/mod/`
3. **Свои пакеты** → относительно `go.mod`

> Просто пишите полный путь от имени модуля — Go сам найдёт!

---

# Псевдонимы при импорте

```go
import (
    "fmt"
    myfmt "mylib/fmt"     // псевдоним при конфликте
    _ "image/png"         // только для side-effect (init)
    . "math"              // не рекомендуется
)
```

---

# Правило одной буквы

**Заглавная буква = экспортируется (публично)**
**Строчная буква = не экспортируется (приватно)**

```go
package user

var privateCounter = 0      // только внутри пакета
var PublicCounter = 100     // доступно всем

type privateUser struct {   // скрыт
    Name string
}

type PublicUser struct {    // доступен всем
    Name string
}
```

---

# Правило работает для ВСЕГО

| Тип | Приватный | Публичный |
|-----|-----------|-----------|
| Переменная | `var count int` | `var Count int` |
| Константа | `const pi = 3.14` | `const Pi = 3.14` |
| Функция | `func help() {}` | `func Help() {}` |
| Тип (struct) | `type point struct {}` | `type Point struct {}` |
| Поле структуры | `point.x` | `point.X` |
| Метод | `func (p *point) move()` | `func (p *Point) Move()` |

---

# Пример: структура с приватными полями

```go
package models

type User struct {
    ID   string  // публичное
    name string  // приватное!
    Age  int     // публичное
}

func (u *User) GetName() string {
    return u.name  // геттер для приватного поля
}

func NewUser(id, name string) *User {
    return &User{ID: id, name: name}
}
```

---

# Использование в main.go

```go
user := models.NewUser("1", "Alice")
fmt.Println(user.ID)        // OK
fmt.Println(user.name)      // ОШИБКА: name не экспортируется
fmt.Println(user.GetName()) // OK
```

---

# Функция init()

Специальная функция, выполняемая автоматически при импорте пакета.

```go
package database

var connection *sql.DB

func init() {
    connection = connectToDB()
    fmt.Println("Database initialized")
}
```

---

# Ваши вопросы?

---

# Go модули

---

# Что дают модули?

| Возможность | Как помогает |
|-------------|--------------|
| Проект в любой папке | Не нужен GOPATH |
| Фиксация версий | Точный список зависимостей |
| Семантическое версионирование | v1.2.3, v2.0.0 |
| Воспроизводимая сборка | go.sum гарантирует неизменность |

---

# Что такое модуль?

**Модуль** = набор Go-пакетов, версионируемых вместе.

```
Модуль github.com/myuser/myproject
├── go.mod          (описание модуля и зависимостей)
├── go.sum          (контрольные суммы зависимостей)
├── main.go
└── utils/
    └── helper.go   (пакет utils внутри модуля)
```

---

# Инициализация модуля

```bash
mkdir myapp
cd myapp

go mod init github.com/username/myapp
```

**Результат — файл go.mod:**

```go
module github.com/username/myapp

go 1.21
```

---

# Структура go.mod

```go
module github.com/username/myapp   // имя модуля

go 1.21                            // минимальная версия Go

require (
    github.com/go-chi/chi/v5 v5.0.10  // прямая зависимость
    github.com/google/uuid v1.3.0     // прямая зависимость
    github.com/foo/bar v1.2.3 // indirect  // косвенная
)
```

---

# Файл go.sum

```go
github.com/go-chi/chi/v5 v5.0.10 h1:r... (хеш)
github.com/go-chi/chi/v5 v5.0.10/go.mod h1:DslCQbL2O+kFtnQdT5RmKxGob5/L5Zc7H/jwRgmVIM=
```

**Назначение:**
- Контрольные суммы (SHA-256) для каждой версии зависимости
- Гарантия, что зависимости не подменили

---

# Импорты в Go

```go
import (
    "fmt"                              // стандартная библиотека
    "net/http"                         // стандартная библиотека
    "github.com/username/myapp/utils"  // из нашего модуля
    "github.com/go-chi/chi/v5"         // из внешнего модуля
)
```

**Правило:** Путь импорта = имя модуля + путь к пакету внутри модуля.

---

# Относительные импорты НЕ работают!

```go
// ОШИБКА
import "./utils"

// ПРАВИЛЬНО — полный путь от имени модуля
import "github.com/username/myapp/utils"
```

> В Go нет относительных импортов (в отличие от Python/Node.js)

---

# Добавление зависимости

```bash
# Способ 1: go get
go get github.com/go-chi/chi/v5

# Способ 2: добавить import и запустить go mod tidy
go mod tidy
```

---

# Команды для работы с модулями

| Команда | Что делает |
|---------|------------|
| `go mod init <name>` | Создаёт новый модуль |
| `go get <package>` | Добавляет зависимость |
| `go get <package>@v1.2.3` | Устанавливает конкретную версию |
| `go mod tidy` | Удаляет неиспользуемые зависимости |
| `go mod download` | Скачивает все зависимости |
| `go list -m all` | Показывает все зависимости |

---

# Ваши вопросы?

---

# Веб-сервисы: стандартная библиотека

---

# Пакет net/http

Содержит функции для разработки веб-серверов и клиентов.

**Ключевые типы:**
- `http.Request` — HTTP-запрос
- `http.Response` — HTTP-ответ

**Ключевые функции:**
- `http.ListenAndServe()` — запуск сервера
- `http.HandleFunc()` — регистрация обработчиков
- `http.Get()`, `http.Post()` — клиентские запросы

---

# Минимальный HTTP-сервер

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
    })
    
    http.ListenAndServe(":8080", nil)
}
```

---

# Сервер с несколькими обработчиками

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Serving: %s\n", r.URL.Path)
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
    t := time.Now().Format(time.RFC1123)
    fmt.Fprintf(w, "<h1>%s</h1>", "The current time is:")
    fmt.Fprintf(w, "<h2>%s</h2>", t)
}

func main() {
    http.HandleFunc("/time", timeHandler)
    http.HandleFunc("/", myHandler)
    
    http.ListenAndServe(":8001", nil)
}
```

---

# Проблемы стандартного ServeMux

| Проблема | Пример |
|----------|--------|
| Нет параметров в пути | `/tasks/1` — нельзя достать `1` |
| Не различает HTTP-методы | Придётся писать `if r.Method == "GET"` |
| Нет встроенной поддержки JSON | Вручную декодировать и кодировать |

**Вывод:** Для REST API нужен роутер получше!

---

# Ваши вопросы?

---

# Chi роутер

---

# Что такое Chi?

**Chi** — легковесный роутер для Go:
- Добавляет параметры в пути: `/tasks/{id}`
- Различает HTTP-методы (GET/POST/PUT/DELETE)
- Совместим со стандартным `http.Handler`
- Имеет удобные middleware из коробки

```bash
go get -u github.com/go-chi/chi/v5
```

---

# Первый сервер на Chi

```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
)

func main() {
    r := chi.NewRouter()
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello Chi!"))
    })
    
    http.ListenAndServe(":8080", r)
}
```

---

# Параметры в пути

```go
r.Get("/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    w.Write([]byte("Task ID: " + id))
})
```

**Примеры:**
- `GET /tasks/123` → `Task ID: 123`
- `GET /tasks/abc` → `Task ID: abc`

---

# Разные HTTP-методы

```go
r.Get("/tasks", listTasks)          // GET    /tasks
r.Post("/tasks", createTask)        // POST   /tasks
r.Get("/tasks/{id}", getTask)       // GET    /tasks/1
r.Put("/tasks/{id}", updateTask)    // PUT    /tasks/1
r.Delete("/tasks/{id}", deleteTask) // DELETE /tasks/1
```

**Преимущество:** Не нужно писать `switch r.Method`

---

# Структура для CRUD

```go
type Task struct {
    ID        string `json:"id"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}

var tasks = make(map[string]Task)
var tasksMutex sync.RWMutex
```

---

# CRUD: Create

```go
func createTask(w http.ResponseWriter, r *http.Request) {
    var newTask Task
    if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    newTask.ID = uuid.New().String()
    
    tasksMutex.Lock()
    tasks[newTask.ID] = newTask
    tasksMutex.Unlock()
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(newTask)
}
```

---

# CRUD: Read

```go
func listTasks(w http.ResponseWriter, r *http.Request) {
    tasksMutex.RLock()
    defer tasksMutex.RUnlock()
    
    taskList := make([]Task, 0, len(tasks))
    for _, task := range tasks {
        taskList = append(taskList, task)
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(taskList)
}

func getTask(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    
    tasksMutex.RLock()
    task, exists := tasks[id]
    tasksMutex.RUnlock()
    
    if !exists {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    
    json.NewEncoder(w).Encode(task)
}
```

---

# CRUD: Update

```go
func updateTask(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    
    var updatedTask Task
    if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    tasksMutex.Lock()
    defer tasksMutex.Unlock()
    
    if _, exists := tasks[id]; !exists {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    
    updatedTask.ID = id
    tasks[id] = updatedTask
    
    json.NewEncoder(w).Encode(updatedTask)
}
```

---

# CRUD: Delete

```go
func deleteTask(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    
    tasksMutex.Lock()
    defer tasksMutex.Unlock()
    
    if _, exists := tasks[id]; !exists {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    
    delete(tasks, id)
    w.WriteHeader(http.StatusNoContent)
}
```

---

# Middleware в Chi

**Middleware** — функция, которая оборачивает обработчик:

```
Запрос → Logger → Recoverer → Обработчик → Ответ
```

```go
import "github.com/go-chi/chi/v5/middleware"

r := chi.NewRouter()
r.Use(middleware.Logger)      // логирует все запросы
r.Use(middleware.Recoverer)   // не падает при панике
```

---

# Свой middleware

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("[%s] %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

r.Use(loggingMiddleware)
```

---

# Ваши вопросы?

---

# Go Standard Project Layout

---

# Проблема: как организовать код?

**Антипаттерн «плоский» проект:**

```
myproject/
├── main.go
├── handlers.go
├── models.go
├── database.go
├── utils.go
└── config.go
```

**Проблемы:**
- 1000+ строк в одном файле
- Непонятно, где что искать
- Невозможно переиспользовать код

---

# Standard Project Layout

Неофициальный стандарт для организации Go-проектов.

**Когда нужен:**
- Продуктовый сервис (6+ месяцев разработки)
- >2 разработчика в проекте
- Open-source библиотека

**Когда НЕ нужен:**
- Один файл (скрипт)
- Pet-проект из 2-3 пакететов

```
myproject/
├── cmd/                 # Точки входа
├── internal/            # Приватный код
├── pkg/                 # Публичный код
├── api/                 # API спецификации
├── configs/             # Конфигурации
├── go.mod
└── README.md
```

---

# cmd/ — точки входа

```
cmd/
├── server/
│   └── main.go      # HTTP-сервер
└── cli/
    └── main.go      # CLI-утилита
```

**Правила:**
- Каждая подпапка = отдельный бинарник
- `main.go` минимальный (только запуск)
- Вся логика — в `internal/` или `pkg/`

---

# internal/ — приватный код

```
internal/
├── app/             # Бизнес-логика
│   └── task/
├── server/          # Настройка сервера
└── storage/         # Работа с БД
```

```go
// ДРУГОЙ ПРОЕКТ НЕ МОЖЕТ:
import "github.com/company/myproject/internal/..."  // ОШИБКА!
```

> Go запрещает импорт из `internal/` извне родительской директории.

---

# pkg/ — публичный код

```
pkg/
├── models/          # Общие структуры
└── logger/          # Обёртка над логгером
```

**Правило:**
- `pkg/` — то, что полезно другим проектам
- `internal/` — специфичное для этого проекта

---

# Чего НЕ должно быть

```
❌ src/          (пережиток GOPATH)
❌ bin/          (бинарники в .gitignore)
❌ common/       (слишком общее название)
❌ utils/        (мусорка, не используйте)
```

---

# Пример реального проекта

```
todo-service/
├── cmd/
│   └── server/main.go
├── internal/
│   ├── app/task/service.go
│   ├── server/handlers.go
│   └── storage/postgres/task.go
├── pkg/
│   ├── models/task.go
│   └── logger/logger.go
├── configs/config.yaml
└── go.mod
```

---

# Ваши вопросы?

---

# Сегодня мы узнали

- **Структуры** — составные типы, указатели, анонимные структуры
- **Пакеты** — видимость, правило одной буквы, init()
- **Модули** — go.mod, go.sum, управление зависимостями
- **net/http** — минимальный сервер, обработчики
- **Chi роутер** — параметры пути, CRUD, middleware
- **Project Layout** — cmd/, internal/, pkg/

---

# Полезные ссылки

- [Chi Router](https://github.com/go-chi/chi)
- [Go Standard Project Layout](https://github.com/golang-standards/project-layout)

---
