# ADR 002

## Название
Защита от обращений к сервису начислений с одним и тем же номером заказа при работе приложения в несколько инстансов

## Статус
proposed

## Дата
2024-04-16

## Описание принятого/предложенного решения
В таблицу order добавлено поле score, содержащее время для следующего обращения. При получении запросов, по которым необходимо обратиться к сервису начислений проверяется значение этого поля. Если поле меньше текущего времени данное поле увеличивается на delta.

```sql
-- таблица
create table if not exists orderData(
		number varchar(255),
		userId int not null,
		status varchar(255),
		accrual float8,
		uploaded_at timestamptz,
		score timestamptz not null default now(),
		primary key(number),
		foreign key (userId) references userInfo(userId)
	);
```

```sql
-- запрос на получение данных
update orderData set score = $1 
		 where number in 
		   (select number from orderdata where status = ANY($2) and score < $3 limit $4) 
		 returning 
		    number, userId, status, accrual, uploaded_at;
```


## Недостатки решения
- 2023-04-17 рост размера БД при обновлении score

## Комментарии
- 2023-04-17 можно попробовать вынести score в отдельную таблицу c tablespace в оперативке; но надо будет сделать аккуратный select с left outer join на случай, если грохнется tablespace
- 2023-04-17 интересно - как происходит репликация таблиц в оперативке

