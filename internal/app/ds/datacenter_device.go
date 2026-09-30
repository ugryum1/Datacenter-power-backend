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

type DatacenterDeviceSerializer struct {
	ID                 uint    `json:"id"`
	Name               string  `json:"name"`
	ShortDescription   *string `json:"short_description"`
	ImageURL           *string `json:"image_url"`
	VideoURL           *string `json:"video_url"`
	PowerConsumptionW  *int32  `json:"power_consumption_w"`
	HeatDissipationBTU *int32  `json:"heat_dissipation_btu"`
	LikesCount         int64   `json:"likes_count"`
}

type DatacenterDeviceListSerializer struct {
	DatacenterDeviceSerializer
	IsCreator int `json:"is_creator"`
}

type DatacenterDevicePublishRequest struct {
	Name               string `json:"name" binding:"required"`
	ShortDescription   string `json:"short_description" binding:"required"`
	PowerConsumptionW  int32  `json:"power_consumption_w" binding:"required"`
	HeatDissipationBTU int32  `json:"heat_dissipation_btu" binding:"required"`
}

type DatacenterDeviceLikeRequest struct {
	Like *int `json:"like" binding:"required,oneof=0 1"`
}

func NewDatacenterDeviceSerializer(
	device DatacenterDevice,
	likesCount int64,
) DatacenterDeviceSerializer {
	serializer := DatacenterDeviceSerializer{
		ID:         device.ID,
		Name:       device.Name,
		LikesCount: likesCount,
	}
	if device.ShortDescription.Valid {
		serializer.ShortDescription = &device.ShortDescription.String
	}
	if device.ImageURL.Valid {
		serializer.ImageURL = &device.ImageURL.String
	}
	if device.VideoURL.Valid {
		serializer.VideoURL = &device.VideoURL.String
	}
	if device.PowerConsumptionW.Valid {
		serializer.PowerConsumptionW = &device.PowerConsumptionW.Int32
	}
	if device.HeatDissipationBTU.Valid {
		serializer.HeatDissipationBTU = &device.HeatDissipationBTU.Int32
	}
	return serializer
}

func NewDatacenterDeviceListSerializer(
	device DatacenterDevice,
	likesCount int64,
	currentEngineerID uint,
) DatacenterDeviceListSerializer {
	serializer := DatacenterDeviceListSerializer{
		DatacenterDeviceSerializer: NewDatacenterDeviceSerializer(device, likesCount),
	}
	if device.CreatorID == currentEngineerID {
		serializer.IsCreator = 1
	}
	return serializer
}
