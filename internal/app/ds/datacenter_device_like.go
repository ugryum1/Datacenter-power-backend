package ds

type DatacenterDeviceLike struct {
	ID         uint `gorm:"primaryKey"`
	EngineerID uint `gorm:"not null;uniqueIndex:idx_datacenter_device_likes_engineer_device"`
	DeviceID   uint `gorm:"not null;uniqueIndex:idx_datacenter_device_likes_engineer_device"`

	Engineer DatacenterEngineer `gorm:"foreignKey:EngineerID"`
	Device   DatacenterDevice   `gorm:"foreignKey:DeviceID"`
}
