package ds

import (
	"database/sql"
	"time"
)

const (
	DatacenterDeviceStatusDraft     = "черновик"
	DatacenterDeviceStatusPublished = "опубликован"
	DatacenterDeviceStatusDeleted   = "удалён"
)

type DatacenterDevice struct {
	ID                 uint           `gorm:"primaryKey"`
	Name               string         `gorm:"type:varchar(100);not null"`
	ShortDescription   sql.NullString `gorm:"type:varchar(500);default:null"`
	Status             string         `gorm:"type:varchar(15);not null"`
	ImageURL           sql.NullString `gorm:"type:varchar(255);default:null"`
	VideoURL           sql.NullString `gorm:"type:varchar(255);default:null"`
	PowerConsumptionW  sql.NullInt32  `gorm:"type:integer;default:null"`
	HeatDissipationBTU sql.NullInt32  `gorm:"type:integer;default:null"`
	DateCreate         time.Time      `gorm:"not null"`
	CreatorID          uint           `gorm:"not null"`
	DateForm           sql.NullTime   `gorm:"default:null"`

	Creator DatacenterEngineer `gorm:"foreignKey:CreatorID"`
}
