# ADR 002

## Название
Обеспечение защиты от некорректного обновления баланса при работе в несколько потоков

## Статус
proposed

## Дата
2024-04-18

## Описание принятого/предложенного решения
```sql
-- таблица
create table if not exists balance(
    balanceId serial,
		userId int not null,
    current float8,
    withdrawn float8
  	release int not null default 0,
		primary key(number),
		foreign key (userId) references userInfo(userId)
	);
```
В таблицу balance добавлено поле release. При обновлении баланса 
```sql
update balancer set current=$1,withdrawn=$2,release=release+1 where userId=$3 and release=$3
```
если при обновлении не было обновлено ни одной записи - значит баланс уже был обновлен кем-то другим.
Необходимо перечитать текущие значения баланса.


## Недостатки решения


## Комментарии

