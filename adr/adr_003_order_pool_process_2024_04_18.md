# ADR_003

## Название
Описание процесса пула системы начислений

## Статус
proposed

## Дата
2024-04-18

## Описание принятого/предложенного решения
### Процесс пула системы начислений
Статусная модель:
  - при регистрации номера статус у номер domain.OrderStatusNew
  - при получении от системы начислений бонусов на данный номер:
       domain.AccrualStatusInvalid -> domain.OrderStratusInvalid
       domain.AccrualStatusProcessed -> domain.OrderStatusProcessing + сохраняются данные о начислении

Пулинг системы начислений запускается вызовом order.PoolAcrualSystem.
 order.PoolAcrualSystem 
   - запускает горутины:
     - пула системы начислений в количестве config.AcrualSystemPoolCount
     - обновления статусов начислений (1 штука), обновляющая статусы либо по config.BatchSize штук(по дефолту 10) либо через 2 секунды
     - получение заказов в статусе domain.OrderStatusNew
 
 При возникновении ошибки все горутины приостанавливают работу на 5 секунд.
 Остановка горутин по остановке контекста.
 Контекс должен содержать logger.

 Функции реализованы по мотивам [Advanced Go Concurrency Patterns](https://go.dev/blog/io2013-talk-concurrency)

## Недостатки решения


## Комментарии

