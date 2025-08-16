#!/bin/bash

# Скрипт для скачивания утилиты statictest

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

FILENAME="statictest-${OS}-${ARCH}"

echo "Скачиваем statictest для $OS-$ARCH..."

# Скачиваем последний релиз из Yandex-Practicum/go-autotests
LATEST_RELEASE=$(curl -s https://api.github.com/repos/Yandex-Practicum/go-autotests/releases/latest)
DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep -o "https://.*${FILENAME}[^\"]*" | head -1)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Не удалось найти ссылку для скачивания $FILENAME"
    echo "Доступные файлы в последнем релизе:"
    echo "$LATEST_RELEASE" | grep -o '"name":"[^"]*"' | sed 's/"name":"//g' | sed 's/"//g'
    exit 1
fi

echo "Скачиваем: $DOWNLOAD_URL"

curl -L -o ".tools/statictest" "$DOWNLOAD_URL"

chmod +x .tools/statictest

echo "statictest успешно скачан в .tools/statictest"
echo "Теперь можно запустить: go vet -vettool=./.tools/statictest ./..."
