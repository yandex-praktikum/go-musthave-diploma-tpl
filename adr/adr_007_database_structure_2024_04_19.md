# ADR_007

## Название
Структура табилц в БД

## Статус
proposed

## Дата
2024-04-19

## Описание принятого/предложенного решения
[PlantUML в Visual Studio Code](https://logrocon.ru/news/plantuml_visualstudiocode)

Запустить prview в linux - Alt-D.

```plantuml
@startuml
class userInfo{
    userId serial
	login text not null
	hash text not null
	salt text not null
    primary key(userId)
    unique (login)
}
class balance {
    balanceId serial
	userId int not null
	current float8 not null default 0
	withdrawn float8 not null default 0
	release int not null default 0
	primary key(balanceId)
	unique (userId)
	foreign key (userId) references userInfo(userId)
}

userInfo "1..1" <- balance

class orderData {
    number varchar(255)
    userId int not null
    status varchar(255)
	accrual float8
    uploaded_at timestamptz
	score timestamptz not null
    primary key(number)
	foreign key (userId) references userInfo(userId)
}

userInfo "1..N" <- orderData

class withdrawal {
    withdrawalId serial
	balanceId int not null
	number varchar(255) not null
	sum float8 not null
	processed_at  timestamptz not null default now()
	primary key(withdrawalId)
	foreign key (balanceId) references balance(balanceId)
}

balance "1..N" <-  withdrawal

@enduml

```


## Недостатки решения

## Связана
ADR_004, ADR_005

## Комментарии
