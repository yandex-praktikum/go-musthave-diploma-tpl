#!/bin/bash

# Скрипт для нагрузочного тестирования API сервера
# Запускает множество запросов к различным эндпоинтам

BASE_URL="http://localhost:8080"
USERS=10
REQUESTS_PER_USER=50

echo "Начинаем нагрузочное тестирование..."
echo "Базовый URL: $BASE_URL"
echo "Количество пользователей: $USERS"
echo "Запросов на пользователя: $REQUESTS_PER_USER"

# Функция для регистрации пользователя
register_user() {
    local user_id=$1
    local username="user$user_id"
    local password="password$user_id"
    
    curl -s -X POST "$BASE_URL/api/user/register" \
        -H "Content-Type: application/json" \
        -d "{\"login\":\"$username\",\"password\":\"$password\"}" > /dev/null
}

# Функция для входа пользователя
login_user() {
    local user_id=$1
    local username="user$user_id"
    local password="password$user_id"
    
    local token=$(curl -s -i -X POST "$BASE_URL/api/user/login" \
        -H "Content-Type: application/json" \
        -d "{\"login\":\"$username\",\"password\":\"$password\"}" | grep -o 'Bearer [^[:space:]]*' | cut -d' ' -f2)
    
    echo "$token"
}

# Функция для загрузки заказа
upload_order() {
    local token=$1
    local order_number=$2
    
    curl -s -X POST "$BASE_URL/api/user/orders" \
        -H "Content-Type: text/plain" \
        -H "Authorization: Bearer $token" \
        -d "$order_number" > /dev/null
}

# Функция для получения заказов
get_orders() {
    local token=$1
    
    curl -s -X GET "$BASE_URL/api/user/orders" \
        -H "Authorization: Bearer $token" > /dev/null
}

# Функция для получения баланса
get_balance() {
    local token=$1
    
    curl -s -X GET "$BASE_URL/api/user/balance" \
        -H "Authorization: Bearer $token" > /dev/null
}

# Функция для снятия средств
withdraw_balance() {
    local token=$1
    local amount=$2
    local order_number=$3
    
    curl -s -X POST "$BASE_URL/api/user/balance/withdraw" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{\"order\":\"$order_number\",\"sum\":$amount}" > /dev/null
}

# Функция для получения истории снятий
get_withdrawals() {
    local token=$1
    
    curl -s -X GET "$BASE_URL/api/user/withdrawals" \
        -H "Authorization: Bearer $token" > /dev/null
}

# Основной цикл нагрузочного тестирования
for ((user_id=1; user_id<=$USERS; user_id++)); do
    echo "Обрабатываем пользователя $user_id..."
    
    # Регистрируем пользователя
    register_user $user_id
    
    # Входим и получаем токен
    token=$(login_user $user_id)
    
    if [ -n "$token" ]; then
        echo "Пользователь $user_id авторизован, токен получен"
        
        # Выполняем запросы для этого пользователя
        for ((req=1; req<=$REQUESTS_PER_USER; req++)); do
            # Генерируем случайный номер заказа
            order_number="1234567890$user_id$req"
            
            # Выполняем различные операции
            case $((req % 6)) in
                0) upload_order "$token" "$order_number" ;;
                1) get_orders "$token" ;;
                2) get_balance "$token" ;;
                3) withdraw_balance "$token" 100 "$order_number" ;;
                4) get_withdrawals "$token" ;;
                5) get_orders "$token" ;;
            esac
            
            # Небольшая пауза между запросами
            sleep 0.01
        done
        
        echo "Пользователь $user_id: выполнено $REQUESTS_PER_USER запросов"
    else
        echo "Ошибка авторизации для пользователя $user_id"
    fi
done

echo "Нагрузочное тестирование завершено!"
echo "Всего запросов: $((USERS * REQUESTS_PER_USER))" 