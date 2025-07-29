#!/bin/bash

# Пример использования API системы лояльности «Гофермарт»
# Сервер должен быть запущен на localhost:8080

BASE_URL="http://localhost:8080"

echo "=== Система лояльности «Гофермарт» - Пример использования ==="
echo

# 1. Регистрация пользователя
echo "1. Регистрация пользователя..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/user/register" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "testuser",
    "password": "password123"
  }')

echo "Ответ: $REGISTER_RESPONSE"
TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o 'Bearer [^"]*' | head -1)
echo "Токен: $TOKEN"
echo

# 2. Аутентификация пользователя
echo "2. Аутентификация пользователя..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/user/login" \
  -H "Content-Type: application/json" \
  -d '{
    "login": "testuser",
    "password": "password123"
  }')

echo "Ответ: $LOGIN_RESPONSE"
echo

# 3. Загрузка номера заказа
echo "3. Загрузка номера заказа..."
ORDER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/user/orders" \
  -H "Content-Type: text/plain" \
  -H "Authorization: $TOKEN" \
  -d "12345678903")

echo "Ответ: $ORDER_RESPONSE"
echo

# 4. Получение списка заказов
echo "4. Получение списка заказов..."
ORDERS_RESPONSE=$(curl -s -X GET "$BASE_URL/api/user/orders" \
  -H "Authorization: $TOKEN")

echo "Ответ: $ORDERS_RESPONSE"
echo

# 5. Получение баланса
echo "5. Получение баланса..."
BALANCE_RESPONSE=$(curl -s -X GET "$BASE_URL/api/user/balance" \
  -H "Authorization: $TOKEN")

echo "Ответ: $BALANCE_RESPONSE"
echo

# 6. Списание средств
echo "6. Списание средств..."
WITHDRAW_RESPONSE=$(curl -s -X POST "$BASE_URL/api/user/balance/withdraw" \
  -H "Content-Type: application/json" \
  -H "Authorization: $TOKEN" \
  -d '{
    "order": "2377225624",
    "sum": 100
  }')

echo "Ответ: $WITHDRAW_RESPONSE"
echo

# 7. Получение списка списаний
echo "7. Получение списка списаний..."
WITHDRAWALS_RESPONSE=$(curl -s -X GET "$BASE_URL/api/user/withdrawals" \
  -H "Authorization: $TOKEN")

echo "Ответ: $WITHDRAWALS_RESPONSE"
echo

echo "=== Пример завершен ===" 