package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
)

// BackupManager управляет бэкапами
type BackupManager struct {
	BackupDir string
	Session   *discordgo.Session
}

// NewBackupManager создает новый экземпляр менеджера бэкапов
func NewBackupManager(backupDir string) *BackupManager {
	// Создаем директорию для бэкапов, если её нет
	os.MkdirAll(backupDir, 0755)

	return &BackupManager{
		BackupDir: backupDir,
	}
}

// CreateBackup создает новый бэкап
func (bm *BackupManager) CreateBackup(guildID, backupName string) error {
	guild, err := bm.Session.Guild(guildID)
	if err != nil {
		return fmt.Errorf("не удалось получить информацию о сервере: %v", err)
	}

	backup := &ServerBackup{
		Name:      backupName,
		CreatedAt: time.Now(),
	}

	// Сохраняем основную информацию
	backup.ServerInfo.Name = guild.Name

	// Скачиваем и сохраняем аватарку
	if guild.Icon != "" {
		iconURL := fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.png", guild.ID, guild.Icon)
		resp, err := http.Get(iconURL)
		if err == nil && resp.StatusCode == 200 {
			backup.ServerInfo.IconURL = iconURL
			backup.ServerInfo.IconData, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}

	// Получаем каналы и сортируем их по позиции
	channels, err := bm.Session.GuildChannels(guildID)
	if err != nil {
		return fmt.Errorf("не удалось получить список каналов: %v", err)
	}

	// Сначала создаем карту категорий
	categoryMap := make(map[string]*ChannelBackup)
	backup.Channels = make([]ChannelBackup, 0, len(channels))

	// Сначала добавляем категории
	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildCategory {
			channel := ChannelBackup{
				ID:       ch.ID,
				Name:     ch.Name,
				Type:     ch.Type,
				Position: ch.Position,
			}
			categoryMap[ch.ID] = &channel
			backup.Channels = append(backup.Channels, channel)
		}
	}

	// Затем добавляем остальные каналы
	for _, ch := range channels {
		if ch.Type != discordgo.ChannelTypeGuildCategory {
			channel := ChannelBackup{
				ID:       ch.ID,
				Name:     ch.Name,
				Type:     ch.Type,
				Topic:    ch.Topic,
				Position: ch.Position,
				ParentID: ch.ParentID,
			}

			// Копируем права доступа
			channel.PermissionOverwrites = make([]PermissionOverwrite, len(ch.PermissionOverwrites))
			for i, perm := range ch.PermissionOverwrites {
				channel.PermissionOverwrites[i] = PermissionOverwrite{
					ID:    perm.ID,
					Type:  int(perm.Type),
					Allow: int64(perm.Allow),
					Deny:  int64(perm.Deny),
				}
			}

			backup.Channels = append(backup.Channels, channel)
		}
	}

	// Получаем и сортируем роли по позиции
	roles := make([]RoleBackup, len(guild.Roles))
	for i, role := range guild.Roles {
		roles[i] = RoleBackup{
			ID:          role.ID,
			Name:        role.Name,
			Color:       role.Color,
			Hoist:       role.Hoist,
			Position:    role.Position,
			Permissions: int64(role.Permissions),
			Mentionable: role.Mentionable,
		}
	}

	// Сортируем роли по позиции
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Position > roles[j].Position
	})

	backup.Roles = roles

	// Сохраняем бэкап
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %v", err)
	}

	filename := filepath.Join(bm.BackupDir, backupName+".backup")
	return os.WriteFile(filename, data, 0644)
}

// ListBackups возвращает список доступных бэкапов
func (bm *BackupManager) ListBackups() ([]string, error) {
	files, err := os.ReadDir(bm.BackupDir)
	if err != nil {
		return nil, err
	}

	var backups []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".backup" {
			backups = append(backups, file.Name()[:len(file.Name())-7]) // убираем .backup
		}
	}
	return backups, nil
}

// GetBackupInfo возвращает информацию о бэкапе
func (bm *BackupManager) GetBackupInfo(name string) (*ServerBackup, error) {
	data, err := os.ReadFile(filepath.Join(bm.BackupDir, name+".backup"))
	if err != nil {
		return nil, err
	}

	var backup ServerBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, err
	}
	return &backup, nil
}

// RestoreBackup восстанавливает сервер из бэкапа
func (bm *BackupManager) RestoreBackup(backupName, targetGuildID string) error {
	backup, err := bm.GetBackupInfo(backupName)
	if err != nil {
		return fmt.Errorf("ошибка чтения бэкапа: %v", err)
	}

	// Обновляем настройки сервера и аватарку
	guildData := &discordgo.GuildParams{
		Name: backup.ServerInfo.Name,
	}

	// Если есть аватарка, устанавливаем её
	if len(backup.ServerInfo.IconData) > 0 {
		// Конвертируем в base64
		iconB64 := base64.StdEncoding.EncodeToString(backup.ServerInfo.IconData)
		iconData := fmt.Sprintf("data:image/png;base64,%s", iconB64)
		guildData.Icon = iconData
	}

	_, err = bm.Session.GuildEdit(targetGuildID, guildData)
	if err != nil {
		return fmt.Errorf("ошибка обновления настроек сервера: %v", err)
	}

	// Удаляем существующие роли (кроме @everyone)
	roles, _ := bm.Session.GuildRoles(targetGuildID)
	for _, role := range roles {
		if role.Name != "@everyone" {
			bm.Session.GuildRoleDelete(targetGuildID, role.ID)
		}
	}

	// Создаем новые роли и сохраняем их ID
	roleMap := make(map[string]string) // старый ID -> новый ID
	for _, role := range backup.Roles {
		if role.Name != "@everyone" {
			newRole, err := bm.Session.GuildRoleCreate(targetGuildID, &discordgo.RoleParams{
				Name:        role.Name,
				Color:       &role.Color,
				Hoist:       &role.Hoist,
				Permissions: &role.Permissions,
				Mentionable: &role.Mentionable,
			})
			if err == nil {
				roleMap[role.ID] = newRole.ID
				// Устанавливаем позицию роли отдельно
				_, err = bm.Session.GuildRoleEdit(targetGuildID, newRole.ID, &discordgo.RoleParams{
					Name:        role.Name,
					Color:       &role.Color,
					Hoist:       &role.Hoist,
					Permissions: &role.Permissions,
					Mentionable: &role.Mentionable,
				})
				if err != nil {
					fmt.Printf("Ошибка установки настроек роли %s: %v\n", role.Name, err)
				}
			}
		} else {
			// Сохраняем ID роли @everyone
			roles, _ := bm.Session.GuildRoles(targetGuildID)
			for _, r := range roles {
				if r.Name == "@everyone" {
					roleMap[role.ID] = r.ID
					break
				}
			}
		}
	}

	// После создания всех ролей устанавливаем их позиции
	type RolePosition struct {
		ID       string
		Position int
	}

	positions := make([]RolePosition, len(backup.Roles))
	for i, role := range backup.Roles {
		if newID, ok := roleMap[role.ID]; ok {
			positions[i] = RolePosition{
				ID:       newID,
				Position: role.Position,
			}
		}
	}

	// Сортируем позиции
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].Position < positions[j].Position
	})

	// Применяем позиции по одной
	for _, pos := range positions {
		if pos.ID != "" {
			// Используем GuildRoleEdit для установки позиции
			_, err = bm.Session.GuildRoleEdit(targetGuildID, pos.ID, &discordgo.RoleParams{
				Name: "", // Оставляем пустым, чтобы не менять
			})
			if err != nil {
				fmt.Printf("Ошибка установки позиции роли: %v\n", err)
			}

			// Даем небольшую задержку между обновлениями
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Удаляем существующие каналы
	channels, _ := bm.Session.GuildChannels(targetGuildID)
	for _, channel := range channels {
		bm.Session.ChannelDelete(channel.ID)
	}

	// Сначала создаем категории
	categoryMap := make(map[string]string) // старый ID -> новый ID
	for _, channel := range backup.Channels {
		if channel.Type == discordgo.ChannelTypeGuildCategory {
			// Обновляем ID ролей в правах доступа
			perms := make([]*discordgo.PermissionOverwrite, 0)
			for _, perm := range channel.PermissionOverwrites {
				if newID, ok := roleMap[perm.ID]; ok {
					perms = append(perms, &discordgo.PermissionOverwrite{
						ID:    newID,
						Type:  discordgo.PermissionOverwriteType(perm.Type),
						Allow: int64(perm.Allow),
						Deny:  int64(perm.Deny),
					})
				}
			}

			createData := &discordgo.GuildChannelCreateData{
				Name:                 channel.Name,
				Type:                 channel.Type,
				Position:             channel.Position,
				PermissionOverwrites: perms,
			}

			newChannel, err := bm.Session.GuildChannelCreateComplex(targetGuildID, *createData)
			if err == nil {
				categoryMap[channel.ID] = newChannel.ID
			}
		}
	}

	// Затем создаем остальные каналы
	for _, channel := range backup.Channels {
		if channel.Type != discordgo.ChannelTypeGuildCategory {
			// Обновляем ID ролей в правах доступа
			perms := make([]*discordgo.PermissionOverwrite, 0)
			for _, perm := range channel.PermissionOverwrites {
				if newID, ok := roleMap[perm.ID]; ok {
					perms = append(perms, &discordgo.PermissionOverwrite{
						ID:    newID,
						Type:  discordgo.PermissionOverwriteType(perm.Type),
						Allow: int64(perm.Allow),
						Deny:  int64(perm.Deny),
					})
				}
			}

			createData := &discordgo.GuildChannelCreateData{
				Name:                 channel.Name,
				Type:                 channel.Type,
				Topic:                channel.Topic,
				Position:             channel.Position,
				ParentID:             categoryMap[channel.ParentID],
				PermissionOverwrites: perms,
			}

			_, err := bm.Session.GuildChannelCreateComplex(targetGuildID, *createData)
			if err != nil {
				fmt.Printf("Ошибка создания канала %s: %v\n", channel.Name, err)
			}
		}
	}

	return nil
}
