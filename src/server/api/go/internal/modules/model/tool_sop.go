package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ToolSOP struct {
	ID              uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Order           int               `gorm:"not null;uniqueIndex:uq_sop_block_id_order,priority:2" json:"order"`
	Action          string            `gorm:"type:text;not null" json:"action"`
	ToolReferenceID uuid.UUID         `gorm:"type:uuid;not null;index:idx_tool_sop_tool_reference_id" json:"tool_reference_id"`
	ToolReference   *ToolReference    `gorm:"constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"-"`
	SOPBlockID      uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:uq_sop_block_id_order,priority:1" json:"sop_block_id"`
	SOPBlock        *Block            `gorm:"constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"-"`
	Props           datatypes.JSONMap `gorm:"type:jsonb" swaggertype:"object" json:"props"`

	CreatedAt time.Time `gorm:"autoCreateTime;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (ToolSOP) TableName() string { return "tool_sops" }
