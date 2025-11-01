#!/bin/bash
# Скрипт для налаштування debug утиліт

echo "🔧 Налаштування debug утиліт..."

# Створюємо директорію debug, якщо її немає
mkdir -p debug

# Переміщуємо файли (якщо вони в корені)
if [ -f "test_auth.go" ]; then
    mv test_auth.go debug/
    echo "✓ Переміщено test_auth.go -> debug/"
fi

if [ -f "reset_password.go" ]; then
    mv reset_password.go debug/
    echo "✓ Переміщено reset_password.go -> debug/"
fi

echo ""
echo "✓ Налаштування завершено!"
echo ""
echo "Доступні команди:"
echo "  go run debug/test_auth.go      - Діагностика аутентифікації"
echo "  go run debug/reset_password.go - Скидання пароля користувача"