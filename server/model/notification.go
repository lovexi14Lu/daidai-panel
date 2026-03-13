package model

import (
	"time"
)

type NotifyChannel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Type      string    `gorm:"size:32;not null" json:"type"`
	Config    string    `gorm:"type:text;default:'{}'" json:"-"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (NotifyChannel) TableName() string {
	return "notify_channels"
}

func (n *NotifyChannel) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":         n.ID,
		"name":       n.Name,
		"type":       n.Type,
		"config":     n.Config,
		"enabled":    n.Enabled,
		"created_at": n.CreatedAt,
		"updated_at": n.UpdatedAt,
	}
}
