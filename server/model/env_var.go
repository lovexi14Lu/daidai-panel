package model

import (
	"time"
)

type EnvVar struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `gorm:"size:128;index;not null" json:"name"`
	Value     string    `gorm:"type:text;default:''" json:"value"`
	Remarks   string    `gorm:"size:256;default:''" json:"remarks"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	Position  float64   `gorm:"default:10000.0;index" json:"position"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	Group     string    `gorm:"size:64;default:'';index" json:"group"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (EnvVar) TableName() string {
	return "env_vars"
}

func (e *EnvVar) ToDict() map[string]interface{} {
	return map[string]interface{}{
		"id":         e.ID,
		"name":       e.Name,
		"value":      e.Value,
		"remarks":    e.Remarks,
		"enabled":    e.Enabled,
		"position":   e.Position,
		"sort_order": e.SortOrder,
		"group":      e.Group,
		"created_at": e.CreatedAt,
		"updated_at": e.UpdatedAt,
	}
}
