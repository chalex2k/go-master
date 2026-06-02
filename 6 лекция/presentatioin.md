---
marp: true
theme: default
size: 16:9
paginate: true
---

# Подключение БД
## Леция 6

---

# Цель занятия

Сегодня добавляем состояние через базу данных,
чтобы сервис стал полноценным приложением.

---

# Сегодня разберём

- Как проектировать схему БД на SQL
- Как управлять изменениями схемы через `goose`
- Как ходить в БД из Go через `pgx`
- Зачем использовать query-builder (`squirrel`)
- Как правильно отключаться от бд

---

# Что такое база данных

База данных — организованная система хранения и доступа к данным.

БД позволяют:
- Сохранять данные между перезапусками приложения
- Валидировать и ограничивать данные на уровне схемы
- Быстро искать информацию в больших объёмах
- Безопасно обслуживать параллельные запросы

---

# Типы БД

- `SQL`: строгая схема, связи, транзакции (`PostgreSQL`, `MySQL`, `SQLite`)
- `NoSQL document`: гибкая схема, нет классических `JOIN` (`MongoDB`)
- `Key-value`: максимальная скорость, минимум структуры (`Redis`)
- `Column-oriented`: аналитика на больших данных (`ClickHouse`)
- `Graph`: связи как граф (`Neo4j`)

---

# PostgreSQL

PostgreSQL — объектно-реляционная СУБД для production-нагрузок.

Ключевые свойства:
- ACID и транзакции
- Расширяемость (типы, функции, операторы)
- `JSON/JSONB`
- Сложные запросы (`WITH`, window functions)

---

# Локальный запуск БД

Для разработки обычно поднимаем PostgreSQL в Docker:

- Быстро стартует
- Изолированное окружение
- Одинаковое поведение у всей команды

---

# Создание таблиц

```sql
CREATE TABLE directors (
    id SERIAL PRIMARY KEY,
    last_name TEXT NOT NULL,
    first_name TEXT NOT NULL
);

CREATE TABLE films (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    release_date DATE NOT NULL,
    director_id INTEGER NOT NULL REFERENCES directors(id) ON DELETE RESTRICT,
    uuid UUID UNIQUE NOT NULL,
    rating DECIMAL(3,1) CHECK (rating >= 0 AND rating <= 10),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

---

# Что важно в схеме

- `PRIMARY KEY` — уникальный идентификатор записи
- `NOT NULL` — обязательные поля
- `REFERENCES` — внешние ключи и целостность связей
- `UNIQUE` — уникальные значения
- `CHECK` — проверка диапазонов и инвариантов
- `DEFAULT` — значения по умолчанию

---

# Вставка данных

```sql
INSERT INTO directors (last_name, first_name) VALUES
('Нолан', 'Кристофер'),
('Кэмерон', 'Джеймс');

INSERT INTO films (title, release_date, director_id, uuid, rating) VALUES
('Начало', '2010-07-16', 1, 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 8.8),
('Интерстеллар', '2014-11-06', 1, 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 8.6);
```

---

# SELECT с JOIN и агрегацией

```sql
SELECT
    d.id AS director_id,
    d.first_name || ' ' || d.last_name AS director_full_name,
    COUNT(f.id) AS films_count
FROM directors d
JOIN films f ON d.id = f.director_id
WHERE f.release_date >= '2000-01-01'
GROUP BY d.id, d.first_name, d.last_name
ORDER BY films_count DESC;
```

---

# Зачем индексы

Без индексов PostgreSQL часто делает `Sequential Scan`:

- Чтение всей таблицы
- Время поиска `O(n)`
- Рост latency с ростом объёма данных

Индекс ускоряет доступ к данным по нужным полям.

---

# Виды индексов

- `B-tree` — основной выбор для большинства задач
- `Hash` — точные совпадения
- `GIN` — составные/документные данные
- `GiST` — геоданные и полнотекстовый поиск

```sql
CREATE INDEX idx_film_release_date ON films(release_date);
```

---

# Ваши вопросы?

---

# Миграции

---

# Что дают миграции

- Версионирование схемы БД
- Применение изменений в правильном порядке
- Откат через `down`-миграции
- Автоматизация для CI/CD и локальной разработки

---

# Goose

`goose` — популярный инструмент миграций для Go-проектов.

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
mkdir -p migrations
goose create init_schema sql
```

---

# Переменные окружения Goose

```bash
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="postgres://user:pass@localhost:5432/mydb"
```

---

# Пример миграции

```sql
-- up
CREATE TABLE IF NOT EXISTS films (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  release_date DATE NOT NULL
);
CREATE INDEX idx_film_release_date ON films(release_date);

-- down
DROP INDEX IF EXISTS idx_film_release_date;
DROP TABLE IF EXISTS films;
```

---

# Основные команды Goose

```bash
goose status
goose up
goose up-by-one
goose down
goose down-to 20240115120000
goose version
goose create add_feature sql
```

---

# Ваши вопросы?

---

# Подключаем приложение к БД

---

# Пакет pgx

`pgx` — современный и быстрый драйвер PostgreSQL для Go.

- Низкоуровневое и эффективное API
- Поддержка пулов соединений через `pgxpool`
- Удобная работа с параметризованными запросами

---

# Базовый запрос через pgx

```go
conn, err := pgx.Connect(ctx, "postgres://user:pass@localhost:5432/db")

var title string
err = conn.QueryRow(ctx,
    "SELECT title FROM films WHERE id = $1",
    123,
).Scan(&title)
```

---

# Почему нужен пул соединений

Один коннект на всё приложение — узкое место.
Новый коннект на каждый запрос — дорого.

Решение: `pgxpool`
- Переиспользует соединения
- Даёт параллелизм
- Ограничивает нагрузку на БД
- Пересоздаёт сломанные соединения

---

# Создание пула pgxpool

```go
config, err := pgxpool.ParseConfig(connString)
config.MaxConns = 20
config.MinConns = 5
config.MaxConnLifetime = time.Hour
config.MaxConnIdleTime = 30 * time.Minute
config.HealthCheckPeriod = time.Minute

pool, err := pgxpool.NewWithConfig(ctx, config)
err = pool.Ping(ctx)
```

---

# Graceful shutdown

---

# Зачем нужен graceful shutdown

При резком завершении процесса можно получить:
- Обрывы клиентских запросов
- Незавершённые операции в БД
- Утечки ресурсов (коннекты, файлы, воркеры)

Цель graceful shutdown: корректно остановить сервис без потери консистентности.

---

# Базовый паттерн завершения

1. Поймать сигнал `SIGTERM` / `SIGINT`
2. Остановить приём новых запросов
3. Дождаться завершения активных операций
4. Закрыть пул БД и освободить ресурсы

---

# Graceful shutdown — демо

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

pool, err := pgxpool.NewWithConfig(ctx, config)
if err != nil {
    log.Fatalf("pgxpool init: %v", err)
}
defer pool.Close()

<-ctx.Done()
log.Println("shutdown signal received")
```

---

# SQL-инъекции

Опасный вариант:

```go
query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)
```

Безопасный вариант: только параметры (`$1`, `$2`, ...).

```go
err := pool.QueryRow(ctx,
    "SELECT id FROM users WHERE username = $1",
    username,
).Scan(&id)
```

---

# Транзакции

Используем, когда несколько операций должны выполниться атомарно.

```go
tx, err := conn.Begin(ctx)
if err != nil { return err }
defer tx.Rollback(ctx)

// ... несколько запросов через tx ...

if err := tx.Commit(ctx); err != nil {
    return err
}
```

---

# Ваши вопросы?

---

# Query-builder Squirrel

---

# Зачем query-builder

Проблема длинных SQL-строк в коде:
- Тяжело читать
- Сложно безопасно расширять
- Легко ошибиться в позициях аргументов

`Squirrel` позволяет собирать запрос по частям.

---

# Пример Squirrel

```go
films := sq.Select("id", "title", "release_date", "director_id", "uuid", "rating", "updated_at").
    From("films").
    Where(sq.Eq{"rating": 8.5}).
    Where(sq.Gt{"release_date": "2010-01-01"}).
    OrderBy("release_date DESC").
    Limit(10)

sql, args, err := films.ToSql()
```

---

# Ваши вопросы?

---