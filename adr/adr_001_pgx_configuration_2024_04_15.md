# ADR 001

## Название
Настрока соединения с БД

## Статус
proposed

## Дата
2024-04-15

## Описание принятого/предложенного решения
На основе статьи [Как работать с Postgres в Go: практики, особенности, нюансы](https://habr.com/ru/companies/oleg-bunin/articles/461935/) настройка pgx v5 выглядит так:

```go
pConf.MaxConns = int32(...) // ограничение кол-ва соединений
pConf.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol // режим Simple Query
pConf.ConnConfig.RuntimeParams["standard_conforming_strings"] = "on" 

```

[Протокол Simple Query](https://www.postgresql.org/docs/15/protocol-flow.html#id-1.10.6.7.4)



