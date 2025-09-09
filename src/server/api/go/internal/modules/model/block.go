package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	BlockTypePage  = "page"
	BlockTypeBlock = "block"
)

type Block struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	SpaceID uuid.UUID `gorm:"type:uuid;not null;index:idx_blocks_space;index:idx_blocks_space_type_archived,priority:1;uniqueIndex:ux_blocks_space_parent_sort,priority:1" json:"space_id"`
	Space   *Space    `gorm:"constraint:fk_blocks_space,OnUpdate:CASCADE,OnDelete:CASCADE;" json:"space"`

	Type string `gorm:"type:text;not null;index:idx_blocks_space_type;index:idx_blocks_space_type_archived,priority:2;check:chk_blocks_type,type IN ('page','block')" json:"type"`

	ParentID *uuid.UUID `gorm:"type:uuid;check:chk_blocks_parent_rule,(type = 'block' AND parent_id IS NOT NULL) OR (type = 'page');uniqueIndex:ux_blocks_space_parent_sort,priority:2" json:"parent_id"`
	Parent   *Block     `gorm:"constraint:fk_blocks_parent,OnUpdate:CASCADE,OnDelete:CASCADE;" json:"parent"`

	Title string                             `gorm:"type:text;not null;default:''" json:"title"`
	Props datatypes.JSONType[map[string]any] `gorm:"type:jsonb;not null;default:'{}'" swaggertype:"object" json:"props"`

	Sort       int64 `gorm:"not null;default:0;uniqueIndex:ux_blocks_space_parent_sort,priority:3" json:"sort"`
	IsArchived bool  `gorm:"not null;default:false;index:idx_blocks_space_type_archived,priority:3;index" json:"is_archived"`

	Children  []*Block  `gorm:"foreignKey:ParentID;constraint:fk_blocks_children,OnUpdate:CASCADE,OnDelete:CASCADE;" json:"children,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Block) TableName() string { return "blocks" }
