package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	token string
)

func init() {
	// Читаем токен из файла или переменной окружения
	token = os.Getenv("DISCORD_TOKEN")
	if token == "" {
		// Пытаемся прочитать из файла
		if data, err := os.ReadFile("token.txt"); err == nil {
			token = strings.TrimSpace(string(data))
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: serverbackup <команда> [аргументы]")
		os.Exit(1)
	}

	if token == "" {
		fmt.Println("Ошибка: токен не найден")
		fmt.Println("Установите токен через меню программы")
		os.Exit(1)
	}

	// Создаем сессию с пользовательским токеном
	session, err := discordgo.New(token)
	if err != nil {
		fmt.Println("Ошибка создания сессии:", err)
		os.Exit(1)
	}

	// Отключаем некоторые события, которые нам не нужны
	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessages

	// Игнорируем ошибки при обработке событий
	session.SyncEvents = false
	session.StateEnabled = false

	// Открываем соединение с Discord
	err = session.Open()
	if err != nil {
		fmt.Println("Ошибка подключения к Discord:", err)
		os.Exit(1)
	}
	defer session.Close()

	backupManager := NewBackupManager("./backups")
	backupManager.Session = session

	switch os.Args[1] {
	case "backup":
		if len(os.Args) < 4 {
			fmt.Println("Использование: serverbackup backup <server_id> <backup_name>")
			os.Exit(1)
		}
		serverID := os.Args[2]
		backupName := os.Args[3]

		// Проверяем доступ к серверу
		_, err := session.Guild(serverID)
		if err != nil {
			fmt.Printf("Ошибка доступа к серверу: %v\n", err)
			fmt.Println("Убедитесь, что:")
			fmt.Println("1. ID сервера указан правильно")
			fmt.Println("2. Токен действителен")
			fmt.Println("3. У вас есть доступ к серверу")
			os.Exit(1)
		}

		err = backupManager.CreateBackup(serverID, backupName)
		if err != nil {
			fmt.Printf("Ошибка создания бэкапа: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Бэкап '%s' успешно создан!\n", backupName)

	case "list":
		backups, err := backupManager.ListBackups()
		if err != nil {
			fmt.Printf("Ошибка получения списка: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Доступные бэкапы:")
		for _, name := range backups {
			backup, _ := backupManager.GetBackupInfo(name)
			fmt.Printf("- %s (создан: %s)\n", name, backup.CreatedAt.Format("02.01.2006 15:04:05"))
		}

	case "info":
		if len(os.Args) < 3 {
			fmt.Println("Использование: serverbackup info <backup_name>")
			os.Exit(1)
		}
		backup, err := backupManager.GetBackupInfo(os.Args[2])
		if err != nil {
			fmt.Printf("Ошибка получения информации: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Информация о бэкапе %s:\n", os.Args[2])
		fmt.Printf("Название сервера: %s\n", backup.ServerInfo.Name)
		fmt.Printf("Создан: %s\n", backup.CreatedAt.Format("02.01.2006 15:04:05"))
		fmt.Printf("Количество каналов: %d\n", len(backup.Channels))
		fmt.Printf("Количество ролей: %d\n", len(backup.Roles))

	case "restore":
		if len(os.Args) < 4 {
			fmt.Println("Использование: serverbackup restore <backup_name> <target_server_id>")
			os.Exit(1)
		}
		err := backupManager.RestoreBackup(os.Args[2], os.Args[3])
		if err != nil {
			fmt.Printf("Ошибка восстановления: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Сервер успешно восстановлен!")
	}
}
