package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"datacenter-power/internal/app/ds"
)

func (r *Repository) GetPublishedDatacenterDevices() ([]ds.DatacenterDevice, error) {
	var devices []ds.DatacenterDevice
	err := r.db.Where("status = ?", ds.DatacenterDeviceStatusPublished).
		Order("id").
		Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *Repository) SearchPublishedDatacenterDevicesByPower(
	minPowerW int,
) ([]ds.DatacenterDevice, error) {
	var devices []ds.DatacenterDevice
	err := r.db.Where(
		"status = ? AND power_consumption_w >= ?",
		ds.DatacenterDeviceStatusPublished, minPowerW,
	).Order("id").Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *Repository) GetPublishedDatacenterDeviceByID(id int) (ds.DatacenterDevice, error) {
	device := ds.DatacenterDevice{}
	err := r.db.Where("id = ? AND status = ?", id, ds.DatacenterDeviceStatusPublished).
		First(&device).Error
	if err != nil {
		return ds.DatacenterDevice{}, err
	}
	return device, nil
}

func (r *Repository) GetNextPublishedDatacenterDevice(afterID int) (ds.DatacenterDevice, error) {
	device := ds.DatacenterDevice{}
	err := r.db.Where("id > ? AND status = ?", afterID, ds.DatacenterDeviceStatusPublished).
		Order("id").
		First(&device).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = r.db.Where("status = ?", ds.DatacenterDeviceStatusPublished).
			Order("id").
			First(&device).Error
	}
	if err != nil {
		return ds.DatacenterDevice{}, err
	}
	return device, nil
}

func (r *Repository) GetFirstPublishedDatacenterDeviceID() uint {
	var deviceID uint
	err := r.db.Model(&ds.DatacenterDevice{}).
		Where("status = ?", ds.DatacenterDeviceStatusPublished).
		Order("id").
		Select("id").
		First(&deviceID).Error
	if err != nil {
		return 0
	}
	return deviceID
}

func (r *Repository) GetMaxPublishedPowerConsumptionW() int {
	var maxPowerW int
	err := r.db.Model(&ds.DatacenterDevice{}).
		Where("status = ?", ds.DatacenterDeviceStatusPublished).
		Select("COALESCE(MAX(power_consumption_w), 0)").
		Scan(&maxPowerW).Error
	if err != nil {
		return 0
	}
	return maxPowerW
}

func (r *Repository) GetDatacenterDeviceLikesCount(deviceID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ds.DatacenterDeviceLike{}).
		Where("device_id = ?", deviceID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) GetDatacenterDevicesLikesCounts(deviceIDs []uint) (map[uint]int64, error) {
	var rows []struct {
		DeviceID   uint
		LikesCount int64
	}
	err := r.db.Model(&ds.DatacenterDeviceLike{}).
		Select("device_id, COUNT(*) AS likes_count").
		Where("device_id IN ?", deviceIDs).
		Group("device_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	likesCounts := make(map[uint]int64, len(rows))
	for _, row := range rows {
		likesCounts[row.DeviceID] = row.LikesCount
	}
	return likesCounts, nil
}

func (r *Repository) GetDatacenterDeviceDraft(creatorID uint) (*ds.DatacenterDevice, error) {
	device := &ds.DatacenterDevice{}
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.DatacenterDeviceStatusDraft).
		First(device).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (r *Repository) CreateDatacenterDeviceDraft(
	creatorID uint,
	name string,
) (*ds.DatacenterDevice, error) {
	device := ds.DatacenterDevice{
		Name:       name,
		Status:     ds.DatacenterDeviceStatusDraft,
		DateCreate: time.Now(),
		CreatorID:  creatorID,
	}
	err := r.db.Create(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *Repository) PublishDatacenterDeviceDraft(creatorID uint,
	name,
	shortDescription string,
	powerConsumptionW,
	heatDissipationBTU int,
) (uint, error) {
	device, err := r.GetDatacenterDeviceDraft(creatorID)
	if err != nil {
		return 0, err
	}
	if device == nil {
		return 0, errors.New("черновик устройства дата-центра не найден")
	}

	err = r.db.Model(device).Updates(ds.DatacenterDevice{
		Name:               name,
		ShortDescription:   sql.NullString{String: shortDescription, Valid: true},
		Status:             ds.DatacenterDeviceStatusPublished,
		PowerConsumptionW:  sql.NullInt32{Int32: int32(powerConsumptionW), Valid: true},
		HeatDissipationBTU: sql.NullInt32{Int32: int32(heatDissipationBTU), Valid: true},
		DateForm:           sql.NullTime{Time: time.Now(), Valid: true},
	}).Error
	if err != nil {
		return 0, err
	}
	return device.ID, nil
}

func (r *Repository) DeleteDatacenterDevice(deviceID uint) error {
	query := "UPDATE datacenter_devices SET status = $1 WHERE id = $2 RETURNING id"

	row := r.db.Raw(query, ds.DatacenterDeviceStatusDeleted, deviceID).Row()

	var deletedID uint
	err := row.Scan(&deletedID)
	if err != nil {
		return fmt.Errorf("ошибка при удалении устройства дата-центра с id %d: %w", deviceID, err)
	}
	return nil
}
