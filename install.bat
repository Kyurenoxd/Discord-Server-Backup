@echo off
chcp 65001 >nul
title Discord Server Backup Installer
color 0b
cls

echo.
echo  ============================
echo    Discord Server Backup - Setup
echo    Author: Kyurenoxd
echo    GitHub: github.com/Kyurenoxd
echo    Website: kyureno.dev
echo  ============================
echo.

:: Проверяем наличие Go
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo  [!] Go не найден. Пожалуйста, установите Go с https://golang.org/
    echo.
    pause
    exit /b 1
) else (
    echo  [✓] Go уже установлен
)

:: Проверяем версию Go
echo  Версия Go:
go version
echo.

echo  [1/3] Инициализация модуля...
go mod init github.com/Kyurenoxd/ServerBackup
if %errorlevel% neq 0 (
    echo  [!] Модуль уже инициализирован, продолжаем...
)

echo  [2/3] Установка зависимостей...
go get github.com/bwmarrin/discordgo@v0.28.1
if %errorlevel% neq 0 (
    echo  [!] Ошибка при установке зависимостей
    pause
    exit /b 1
)

echo  [3/3] Проверка и обновление зависимостей...
go mod tidy
if %errorlevel% neq 0 (
    echo  [!] Ошибка при обновлении зависимостей
    pause
    exit /b 1
)

echo.
echo  [✓] Зависимости установлены успешно!
echo  [✓] Discord Go API установлен
echo  [✓] Все готово к работе!

echo.
echo  ============================
echo    Установка завершена!
echo    Запустите start.bat
echo  ============================
echo.
echo  Нажмите любую клавишу для выхода...
pause >nul 