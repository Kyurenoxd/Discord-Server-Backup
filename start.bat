@echo off
chcp 65001 >nul
title Discord Server Backup
cls

:: Компилируем программу при запуске
color 0B
echo Компиляция программы...
go build -o serverbackup.exe
if %errorlevel% neq 0 (
    color 0C
    echo [!] Ошибка при компиляции
    pause
    exit /b 1
)
color 0A
echo Компиляция завершена успешно!
timeout /t 2 >nul

:menu
cls
color 0B
echo.
echo  ╔════════════════════════════════╗
echo  ║     Discord Server Backup v1.0   ║
echo  ║        Author: Kyurenoxd        ║
echo  ║   GitHub: github.com/Kyurenoxd  ║
echo  ║     Website: kyureno.dev        ║
echo  ╚════════════════════════════════╝
echo.
echo  Выберите действие:
echo.
echo  [1] Инструкция по получению токена
echo  [2] Установить токен
echo  [3] Создать бэкап сервера
echo  [4] Список бэкапов
echo  [5] Восстановить сервер
echo  [6] Настройки расписания
echo  [7] Выход
echo.
set /p choice="  Выберите опцию (1-7): "

if "%choice%"=="1" goto instructions
if "%choice%"=="2" goto settoken
if "%choice%"=="3" goto backup
if "%choice%"=="4" goto list
if "%choice%"=="5" goto restore
if "%choice%"=="6" goto schedule
if "%choice%"=="7" exit

goto menu

:instructions
cls
color 0B
echo.
echo  ╔════════════════════════════════╗
echo  ║     Как получить токен:         ║
echo  ╚════════════════════════════════╝
echo.
color 0F
echo  1. Откройте Discord в браузере (discord.com)
echo  2. Нажмите F12 или Ctrl+Shift+I для открытия DevTools
echo  3. Перейдите во вкладку Network
echo  4. В Discord нажмите Ctrl+R для обновления страницы
echo  5. В поиске DevTools введите "science"
echo  6. Найдите запрос и откройте его
echo  7. В заголовках (Headers) найдите "authorization"
echo  8. Скопируйте значение токена (без кавычек)
echo.
color 0C
echo  ВНИМАНИЕ: Никогда никому не передавайте свой токен!
echo  Это может привести к взлому вашего аккаунта!
echo.
color 0F
echo  [1] Установить токен
echo  [2] Вернуться в меню
echo.
set /p choice="  Выберите опцию (1-2): "
if "%choice%"=="1" goto settoken
if "%choice%"=="2" goto menu
goto instructions

:settoken
cls
color 0B
echo.
echo  ╔════════════════════════════════╗
echo  ║     Установка токена           ║
echo  ╚════════════════════════════════╝
echo.
color 0F
if exist token.txt (
    echo  Токен уже установлен
    echo.
    echo  [1] Изменить токен
    echo  [2] Вернуться в меню
    echo.
    set /p choice="  Выберите опцию (1-2): "
    if "%choice%"=="1" goto entertoken
    if "%choice%"=="2" goto menu
    goto settoken
) else (
    goto entertoken
)

:entertoken
echo  Вставьте ваш токен:
set /p "usertoken="
echo.

:: Сохраняем токен в файл
echo %usertoken%> "token.txt"
color 0A
echo  Токен успешно сохранен!
timeout /t 2 >nul
goto menu

:backup
cls
color 0B
echo.
echo  ╔════════════════════════════════╗
echo  ║     Создание бэкапа сервера     ║
echo  ╚════════════════════════════════╝
echo.
color 0F
echo  Проверка токена...
if not exist token.txt (
    color 0C
    echo  [!] Токен не установлен!
    echo  Сначала установите токен через меню.
    pause
    goto menu
)

echo  Введите ID сервера для создания бэкапа:
set /p server_id=""
echo  Введите название бэкапа:
set /p backup_name=""
echo.
echo  Создание бэкапа сервера...
echo.
serverbackup.exe backup "%server_id%" "%backup_name%"
if %errorlevel% neq 0 (
    color 0C
    echo  [!] Ошибка при создании бэкапа
    pause
    goto menu
)
color 0A
echo  Бэкап успешно создан!
timeout /t 2 >nul
goto menu

:list
cls
echo.
serverbackup.exe list
if %errorlevel% neq 0 (
    color 0C
    echo  [!] Ошибка при получении списка
    pause
)
pause
goto menu

:restore
cls
echo.
echo  ╔════════════════════════════════╗
echo  ║     Восстановление сервера из бэкапа  ║
echo  ╚════════════════════════════════╝
echo.
echo  Введите название бэкапа:
set /p backup_name=""
echo  Введите ID сервера назначения:
set /p target_id=""
echo.
echo  Восстановление сервера из бэкапа...
echo.
serverbackup.exe restore "%backup_name%" "%target_id%"
if %errorlevel% neq 0 (
    color 0C
    echo  [!] Ошибка при восстановлении
    pause
)
goto menu

:schedule
cls
echo.
echo  ╔════════════════════════════════╗
echo  ║     Настройка расписания бэкапов  ║
echo  ╚════════════════════════════════╝
echo.
echo  Введите ID сервера для настройки расписания:
set /p server_id=""
echo  Введите интервал в часах (например: 24):
set /p interval=""
echo.
echo  Настройка расписания бэкапов...
echo.
serverbackup.exe schedule "%server_id%" "%interval%"
if %errorlevel% neq 0 (
    color 0C
    echo  [!] Ошибка при настройке расписания
    pause
)
goto menu 