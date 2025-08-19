package model

import (
	"time"

	"gorm.io/datatypes"
)

type Space struct {
	ID        datatypes.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ProjectID datatypes.UUID    `gorm:"type:uuid;not null;primaryKey;index"`
	Configs   datatypes.JSONMap `gorm:"type:jsonb" swaggertype:"object"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Project Project `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
}

func (Space) TableName() string { return "spaces" }
