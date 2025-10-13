package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Task struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SessionID uuid.UUID `gorm:"type:uuid;not null;index:ix_session_session_id;index:ix_session_session_id_task_id,priority:1;index:ix_session_session_id_task_status,priority:1;uniqueIndex:uq_session_id_task_order,priority:1" json:"session_id"`

	TaskOrder      int               `gorm:"not null;uniqueIndex:uq_session_id_task_order,priority:2" json:"task_order"`
	TaskData       datatypes.JSONMap `gorm:"type:jsonb;not null" swaggertype:"object" json:"task_data"`
	TaskStatus     string            `gorm:"type:text;not null;default:'pending';check:task_status IN ('success','failed','running','pending');index:ix_session_session_id_task_status,priority:2" json:"task_status"`
	IsPlanningTask bool              `gorm:"not null;default:false" json:"is_planning_task"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Task <-> Session
	Session *Session `gorm:"foreignKey:SessionID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"session"`

	// Task <-> Message (one-to-many)
	Messages []Message `gorm:"constraint:OnDelete:SET NULL,OnUpdate:CASCADE;" json:"messages"`
}

func (Task) TableName() string { return "tasks" }
