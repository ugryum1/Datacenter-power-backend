package repository

import (
	"errors"
	"fmt"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusDeleted   = "deleted"
)

const mediaBaseURL = "http://localhost:9000/datacenter-media/"

type Device struct {
	ID                 int
	Name               string
	ShortDescription   string
	Status             string
	ImageURL           string
	VideoURL           string
	PowerConsumptionW  int
	HeatDissipationBTU int
	LikeUserIDs        []int64
}

func (d Device) LikesCount() int {
	return len(d.LikeUserIDs)
}

type Repository struct {
	devices []Device
}

func NewRepository() (*Repository, error) {
	return &Repository{
		devices: []Device{
			{
				ID:                 1,
				Name:               "Dell PowerEdge R750",
				ShortDescription:   "Стоечный сервер 2U для виртуализации и высоконагруженных сервисов дата-центра. Два процессора Intel Xeon Scalable, до 32 отсеков NVMe.",
				Status:             StatusPublished,
				ImageURL:           mediaBaseURL + "server-rack.jpg",
				VideoURL:           mediaBaseURL + "fan-cooling-vertical.webm",
				PowerConsumptionW:  800,
				HeatDissipationBTU: 2730,
				LikeUserIDs:        []int64{101, 102, 103},
			},
			{
				ID:                 2,
				Name:               "Cisco Catalyst 9300",
				ShortDescription:   "Коммутатор ядра сети дата-центра с поддержкой StackWise-480 и модульным блоком питания.",
				Status:             StatusPublished,
				ImageURL:           mediaBaseURL + "switch.jpg",
				VideoURL:           mediaBaseURL + "switch-panel-vertical.webm",
				PowerConsumptionW:  350,
				HeatDissipationBTU: 1194,
				LikeUserIDs:        []int64{201, 202},
			},
			{
				ID:                 3,
				Name:               "NetApp FAS8300",
				ShortDescription:   "Система хранения данных (СХД) корпоративного класса с гибридными дисковыми полками.",
				Status:             StatusPublished,
				ImageURL:           mediaBaseURL + "storage.jpg",
				VideoURL:           mediaBaseURL + "hdd-storage-vertical.webm",
				PowerConsumptionW:  1200,
				HeatDissipationBTU: 4094,
				LikeUserIDs:        []int64{301},
			},
			{
				ID:                 4,
				Name:               "APC Smart-UPS SRT 10000VA",
				ShortDescription:   "Источник бесперебойного питания для серверного зала с двойным преобразованием напряжения.",
				Status:             StatusDraft,
				ImageURL:           mediaBaseURL + "ups.jpg",
				VideoURL:           mediaBaseURL + "hdd-storage-vertical.webm",
				PowerConsumptionW:  900,
				HeatDissipationBTU: 3071,
			},
			{
				ID:                 5,
				Name:               "Stulz CyberAir 3PRO",
				ShortDescription:   "Прецизионный кондиционер (CRAC) для охлаждения серверного зала дата-центра.",
				Status:             StatusDeleted,
				ImageURL:           mediaBaseURL + "crac.jpg",
				VideoURL:           mediaBaseURL + "hdd-storage-vertical.webm",
				PowerConsumptionW:  15000,
				HeatDissipationBTU: 51180,
				LikeUserIDs:        []int64{401, 402, 403, 404},
			},
		},
	}, nil
}

func (r *Repository) GetDevices(minPowerW int) ([]Device, error) {
	var result []Device
	for _, d := range r.devices {
		if d.Status != StatusPublished {
			continue
		}
		if minPowerW > 0 && d.PowerConsumptionW < minPowerW {
			continue
		}
		result = append(result, d)
	}
	return result, nil
}

func (r *Repository) GetDevice(id int) (Device, error) {
	for _, d := range r.devices {
		if d.ID == id && d.Status != StatusDeleted {
			return d, nil
		}
	}
	return Device{}, fmt.Errorf("устройство с id %d не найдено", id)
}

func (r *Repository) GetNextDevice(afterID int) (Device, error) {
	found := false
	for _, d := range r.devices {
		if found && d.Status == StatusPublished {
			return d, nil
		}
		if d.ID == afterID {
			found = true
		}
	}
	if !found {
		return Device{}, fmt.Errorf("устройство с id %d не найдено", afterID)
	}
	for _, d := range r.devices {
		if d.Status == StatusPublished {
			return d, nil
		}
	}
	return Device{}, errors.New("нет опубликованных устройств")
}

func (r *Repository) GetDraft() (Device, error) {
	for _, d := range r.devices {
		if d.Status == StatusDraft {
			return d, nil
		}
	}
	return Device{}, errors.New("черновик устройства не найден")
}

func (r *Repository) GetFirstPublishedID() int {
	for _, d := range r.devices {
		if d.Status == StatusPublished {
			return d.ID
		}
	}
	return 0
}
