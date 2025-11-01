@echo off
REM Скрипт для налаштування debug утиліт (Windows)

echo 🔧 Налаштування debug утиліт...
echo.

REM Створюємо директорію debug
if not exist "debug" mkdir debug

REM Переміщуємо файли (якщо вони в корені)
if exist "test_auth.go" (
    move test_auth.go debug\
    echo ✓ Переміщено test_auth.go -^> debug\
)

if exist "reset_password.go" (
    move reset_password.go debug\
    echo ✓ Переміщено reset_password.go -^> debug\
)

echo.
echo ✓ Налаштування завершено!
echo.
echo Доступні команди:
echo   go run debug\test_auth.go      - Діагностика аутентифікації
echo   go run debug\reset_password.go - Скидання пароля користувача
echo.
pause