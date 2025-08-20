package model

import (
	"time"

	"gorm.io/datatypes"
)

type Session struct {
	ID        datatypes.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ProjectID datatypes.UUID    `gorm:"type:uuid;not null;primaryKey;index"`
	SpaceID   datatypes.UUID    `gorm:"type:uuid"`
	Configs   datatypes.JSONMap `gorm:"type:jsonb" swaggertype:"object"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Project Project `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
}

func (Session) TableName() string { return "sessions" }
