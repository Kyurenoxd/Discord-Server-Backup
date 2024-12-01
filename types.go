package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// ServerBackup представляет полный бэкап сервера
type ServerBackup struct {
	Name       string    `json:"name"`       // Имя бэкапа, которое задает пользователь
	CreatedAt  time.Time `json:"created_at"` // Дата создания
	ServerInfo struct {
		Name     string `json:"name"`      // Название сервера
		IconURL  string `json:"icon_url"`  // URL аватарки
		IconData []byte `json:"icon_data"` // Данные аватарки
	} `json:"server_info"`

	Channels []ChannelBackup `json:"channels"` // Каналы
	Roles    []RoleBackup    `json:"roles"`    // Роли
}

// ChannelBackup представляет бэкап канала
type ChannelBackup struct {
	ID                   string                `json:"id"` // ID канала
	Name                 string                `json:"name"`
	Type                 discordgo.ChannelType `json:"type"`
	Topic                string                `json:"topic"`
	Position             int                   `json:"position"`
	ParentID             string                `json:"parent_id"` // Для подканалов
	PermissionOverwrites []PermissionOverwrite `json:"permission_overwrites"`
}

// RoleBackup представляет бэкап роли
type RoleBackup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Hoist       bool   `json:"hoist"`
	Position    int    `json:"position"`
	Permissions int64  `json:"permissions"`
	Mentionable bool   `json:"mentionable"`
}

// PermissionOverwrite представляет права доступа
type PermissionOverwrite struct {
	ID    string `json:"id"`
	Type  int    `json:"type"`
	Allow int64  `json:"allow"`
	Deny  int64  `json:"deny"`
}
