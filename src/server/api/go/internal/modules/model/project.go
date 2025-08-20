package model

import (
	"time"

	"gorm.io/datatypes"
)

type Project struct {
	ID      datatypes.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Configs datatypes.JSONMap `gorm:"type:jsonb" swaggertype:"object"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Spaces []Space `gorm:"constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
}

func (Project) TableName() string { return "projects" }
