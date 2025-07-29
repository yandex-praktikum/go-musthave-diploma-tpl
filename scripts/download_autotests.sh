#!/bin/bash

# Скрипт для скачивания утилит gophermarttest и random

set -e

mkdir -p .tools

# Определяем OS и архитектуру
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Маппинг архитектур
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Неподдерживаемая архитектура: $ARCH"
        exit 1
        ;;
esac

echo "Скачиваем автотесты для $OS-$ARCH..."

# Скачиваем последний релиз из Yandex-Practicum/go-autotests
LATEST_RELEASE=$(curl -s https://api.github.com/repos/Yandex-Practicum/go-autotests/releases/latest)

# Скачиваем gophermarttest
GOPHERMART_FILENAME="gophermarttest-${OS}-${ARCH}"
GOPHERMART_DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep -o "https://.*${GOPHERMART_FILENAME}[^\"]*" | head -1)

if [ -z "$GOPHERMART_DOWNLOAD_URL" ]; then
    echo "Не удалось найти ссылку для скачивания $GOPHERMART_FILENAME"
    echo "Доступные файлы в последнем релизе:"
    echo "$LATEST_RELEASE" | grep -o '"name":"[^"]*"' | sed 's/"name":"//g' | sed 's/"//g'
    exit 1
fi

echo "Скачиваем gophermarttest: $GOPHERMART_DOWNLOAD_URL"
curl -L -o ".tools/gophermarttest" "$GOPHERMART_DOWNLOAD_URL"
chmod +x .tools/gophermarttest

# Скачиваем random
RANDOM_FILENAME="random-${OS}-${ARCH}"
RANDOM_DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep -o "https://.*${RANDOM_FILENAME}[^\"]*" | head -1)

if [ -z "$RANDOM_DOWNLOAD_URL" ]; then
    echo "Не удалось найти ссылку для скачивания $RANDOM_FILENAME"
    echo "Доступные файлы в последнем релизе:"
    echo "$LATEST_RELEASE" | grep -o '"name":"[^"]*"' | sed 's/"name":"//g' | sed 's/"//g'
    exit 1
fi

echo "Скачиваем random: $RANDOM_DOWNLOAD_URL"
curl -L -o ".tools/random" "$RANDOM_DOWNLOAD_URL"
chmod +x .tools/random

echo "Автотесты успешно скачаны в .tools/"
echo "Теперь можно запустить тесты с помощью gophermarttest" 