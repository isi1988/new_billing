#!/bin/bash

# Выходим из скрипта, если любая команда завершится с ошибкой
set -e

# --- КОНФИГУРАЦИЯ ---
# Имя исполняемого файла по умолчанию
DEFAULT_APP_NAME="new-billing"
# Получаем имя из первого аргумента скрипта, если он есть, иначе используем значение по умолчанию
APP_NAME=${1:-$DEFAULT_APP_NAME}
# --------------------

# Определяем ОС для правильного расширения файла
GOOS=$(go env GOOS)
if [ "$GOOS" = "windows" ]; then
    APP_NAME="${APP_NAME}.exe"
fi

echo "🚀 Starting production build..."
echo "Output file will be: $APP_NAME"
echo ""

# --- ШАГ 1: Сборка фронтенда ---
echo "📦 [1/2] Building Vue.js frontend..."
# Переходим в папку фронтенда
cd frontend

# Устанавливаем/обновляем зависимости
echo "  -> Running 'npm install'..."
npm install > /dev/null 2>&1 # Скрываем длинный вывод npm

# Собираем статические файлы в папку /dist
echo "  -> Running 'npm run build'..."
npm run build
echo "✅ Frontend build complete."
echo ""

# Возвращаемся в корень проекта
cd ..

# --- ШАГ 2: Сборка бэкенда ---
echo "🛠️  [2/2] Building Go backend and embedding frontend..."
# Используем флаги -ldflags "-s -w" для уменьшения размера исполняемого файла
# -s: убрать таблицу символов
# -w: убрать отладочную информацию
go build -ldflags="-s -w" -o "$APP_NAME" ./cmd/main.go
echo "✅ Go backend build complete."
echo ""

# --- ЗАВЕРШЕНИЕ ---
echo "✨ Build finished successfully!"
echo "Run your application with: ./$APP_NAME"