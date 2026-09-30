package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

func (r *Repository) GetFirstPublishedDatacenterDevice() (ds.DatacenterDevice, error) {
	device := ds.DatacenterDevice{}
	err := r.db.Where("status = ?", ds.DatacenterDeviceStatusPublished).
		Order("id").
		First(&device).Error
	if err != nil {
		return ds.DatacenterDevice{}, err
	}
	return device, nil
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
		return r.GetFirstPublishedDatacenterDevice()
	}
	if err != nil {
		return ds.DatacenterDevice{}, err
	}
	return device, nil
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

func (r *Repository) GetDatacenterDeviceDraft(creatorID uint) (ds.DatacenterDevice, error) {
	device := ds.DatacenterDevice{}
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, ds.DatacenterDeviceStatusDraft).
		First(&device).Error
	if err != nil {
		return ds.DatacenterDevice{}, err
	}
	return device, nil
}

func (r *Repository) CreateDatacenterDeviceDraft(
	creatorID uint,
	name string,
) (ds.DatacenterDevice, error) {
	device := ds.DatacenterDevice{
		Name:       name,
		Status:     ds.DatacenterDeviceStatusDraft,
		DateCreate: time.Now(),
		CreatorID:  creatorID,
	}
	err := r.db.Create(&device).Error
	if err != nil {
		return ds.DatacenterDevice{}, err
	}
	return device, nil
}

type DatacenterMediaFile struct {
	Header      *multipart.FileHeader
	ContentType string
	Filename    string
}

func (r *Repository) AddDatacenterDeviceMedia(
	device *ds.DatacenterDevice,
	image DatacenterMediaFile,
	video DatacenterMediaFile,
) error {
	imageURL, err := r.putDatacenterMediaObject(image)
	if err != nil {
		return err
	}
	videoURL, err := r.putDatacenterMediaObject(video)
	if err != nil {
		return err
	}

	device.ImageURL = sql.NullString{String: imageURL, Valid: true}
	device.VideoURL = sql.NullString{String: videoURL, Valid: true}
	return r.db.Model(device).
		Select("image_url", "video_url").
		Updates(device).Error
}

func (r *Repository) putDatacenterMediaObject(media DatacenterMediaFile) (string, error) {
	file, err := media.Header.Open()
	if err != nil {
		return "", fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	_, err = r.minio.PutObject(
		context.Background(),
		r.minio_bucket_name,
		media.Filename,
		file,
		media.Header.Size,
		minio.PutObjectOptions{
			ContentType: media.ContentType,
		})
	if err != nil {
		return "", fmt.Errorf("не удалось добавить объект в хранилище minio: %w", err)
	}

	return fmt.Sprintf(
		"http://%s/%s/%s",
		r.minio_endpoint, r.minio_bucket_name, media.Filename,
	), nil
}

func (r *Repository) PublishDatacenterDevice(
	deviceID int,
	creatorID uint,
	request ds.DatacenterDevicePublishRequest,
) error {
	result := r.db.Model(&ds.DatacenterDevice{}).
		Where(
			"id = ? AND creator_id = ? AND status = ?",
			deviceID, creatorID, ds.DatacenterDeviceStatusDraft,
		).
		Updates(ds.DatacenterDevice{
			Name:               request.Name,
			ShortDescription:   sql.NullString{String: request.ShortDescription, Valid: true},
			Status:             ds.DatacenterDeviceStatusPublished,
			PowerConsumptionW:  sql.NullInt32{Int32: request.PowerConsumptionW, Valid: true},
			HeatDissipationBTU: sql.NullInt32{Int32: request.HeatDissipationBTU, Valid: true},
			DateForm:           sql.NullTime{Time: time.Now(), Valid: true},
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) DeleteDatacenterDevice(deviceID int, creatorID uint) error {
	result := r.db.Model(&ds.DatacenterDevice{}).
		Where(
			"id = ? AND creator_id = ? AND status = ?",
			deviceID, creatorID, ds.DatacenterDeviceStatusPublished,
		).
		Update("status", ds.DatacenterDeviceStatusDeleted)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) LikeDatacenterDevice(engineerID uint, deviceID uint) error {
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&ds.DatacenterDeviceLike{EngineerID: engineerID, DeviceID: deviceID}).Error
}

func (r *Repository) UnlikeDatacenterDevice(engineerID uint, deviceID uint) error {
	return r.db.Where("engineer_id = ? AND device_id = ?", engineerID, deviceID).
		Delete(&ds.DatacenterDeviceLike{}).Error
}
