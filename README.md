# go-musthave-diploma-tpl

Шаблон репозитория для индивидуального дипломного проекта курса «Go-разработчик»

# Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без
   префикса `https://`) для создания модуля

# Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

Затем добавьте полученные изменения в свой репозиторий.

## DB

### Подключение через psql
`postgres@63bafaa95eee:/$ psql -U yandex_metrics`

### Drop DB
`migrate -database="postgres://postgres:postgres@localhost:5432/yandex_metrics?sslmode=disable" -path internal/config/db/migrations drop`

### psql cli

``
\l - Display database
\c - Connect to database
\dn - List schemas
\dt - List tables inside public schemas
\dt gophermart. - For gophermart. Dot(.) require!
\dt schema1. - List tables inside particular schemas. For eg: 'schema1'.
``
