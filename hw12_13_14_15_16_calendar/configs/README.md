# Конфигурация Calendar Service

## Файлы конфигурации

- `config.yaml` - конфигурация для разработки (in-memory storage)
- `config.prod.yaml` - пример конфигурации для production (PostgreSQL)
- `config.toml` - альтернативный формат конфигурации (TOML)

## Формат конфигурации

Поддерживаются форматы: **YAML** (`.yaml`, `.yml`) и **TOML** (`.toml`)

### Пример YAML:

```yaml
logger:
  level: INFO  # DEBUG, INFO, WARN, ERROR

http:
  host: localhost
  port: 8080

database:
  type: memory  # "memory" или "db"
  dsn: postgres://user:password@host:port/dbname?sslmode=disable
```

### Пример TOML:

```toml
[logger]
level = "INFO"

[http]
host = "localhost"
port = "8080"

[database]
type = "memory"
dsn = "postgres://user:password@host:port/dbname?sslmode=disable"
```

## Параметры

### Logger
- `level` - уровень логирования: `DEBUG`, `INFO`, `WARN`, `ERROR` (по умолчанию: `INFO`)

### HTTP
- `host` - хост для HTTP сервера (по умолчанию: `localhost`)
- `port` - порт для HTTP сервера (по умолчанию: `8080`)

### Database
- `type` - тип хранилища:
  - `memory` - in-memory хранилище (для разработки/тестирования)
  - `db` - PostgreSQL база данных
- `dsn` - строка подключения к PostgreSQL (используется только при `type = "db"`)

## Запуск с конфигурацией

```bash
# С конфигом по умолчанию (./configs/config.yaml)
./calendar

# С указанием пути к конфигу
./calendar --config=/path/to/config.yaml

# С production конфигом
./calendar --config=./configs/config.prod.yaml

# С TOML конфигом
./calendar --config=./configs/config.toml
```

## Переменные окружения

Для production рекомендуется использовать переменные окружения для чувствительных данных:

```bash
export DB_DSN="postgres://user:password@host:port/dbname?sslmode=disable"
```

Затем в конфиге можно использовать:
```yaml
database:
  dsn: ${DB_DSN}
```
