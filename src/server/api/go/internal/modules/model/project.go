package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Project struct {
	ID        uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SecretKey string            `gorm:"type:varchar(64);uniqueIndex;not null" json:"secret_key"`
	Configs   datatypes.JSONMap `gorm:"type:jsonb" swaggertype:"object" json:"configs"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Project <-> Space
	Spaces []Space `gorm:"constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"spaces"`

	// Project <-> Session
	Sessions []Session `gorm:"constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"sessions"`
}

func (Project) TableName() string { return "projects" }
