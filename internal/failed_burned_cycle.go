package internal

import (
	"time"

	"gorm.io/gorm"
)

type FailedBurnedCycle struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
	Amount    uint           `json:"amount"`
}
