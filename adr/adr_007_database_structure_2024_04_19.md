# ADR_007

## Название
Структура табилц в БД

## Статус
proposed

## Дата
2024-04-19

## Описание принятого/предложенного решения
```plantuml
@startuml
class userInfo
class balance

userInfo "1..1" <- balance

class orderData

userInfo "1..N" <- orderData

class withdrawal

balance "1..N" <-  withdrawal

@enduml

```


## Недостатки решения

## Связана
ADR_004, ADR_005

## Комментарии
