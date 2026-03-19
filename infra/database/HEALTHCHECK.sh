#!/bin/sh
set -e

# Проверяем доступность через стандартную системную базу 'postgres'
# Используем переменные среды, которые передадим при запуске
export PGPASSWORD=${POSTGRES_PASSWORD:-postal}
pg_isready -h localhost --port=5550 --username="${POSTGRES_USER:-postgres}" --dbname=postgres || exit 1
