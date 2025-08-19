#!/bin/bash

# Выходим из скрипта, если любая команда завершится с ошибкой
set -e

echo "--- Starting Go backend in the background... ---"
# Запускаем Go сервер в фоновом режиме и сохраняем его PID
go run cmd/main.go &
GO_PID=$!
echo "Go backend started with PID: $GO_PID"

# Функция для остановки фонового процесса при выходе из скрипта (Ctrl+C)
trap "echo '--- Stopping Go backend... ---'; kill $GO_PID" EXIT

echo "--- Starting Vue frontend (Vite dev server)... ---"
echo "--- Press Ctrl+C to stop both servers. ---"
cd frontend
npm run dev

# После завершения `npm run dev` (по Ctrl+C), trap выполнит команду kill