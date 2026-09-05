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

const equipmentMediaBaseURL = "http://localhost:9000/datacenter-media/"

type DatacenterDevice struct {
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

func (d DatacenterDevice) LikesCount() int {
	return len(d.LikeUserIDs)
}

type DatacenterDeviceRepository struct {
	datacenterDevices []DatacenterDevice
}

func NewDatacenterDeviceRepository() (*DatacenterDeviceRepository, error) {
	return &DatacenterDeviceRepository{
		datacenterDevices: []DatacenterDevice{
			{
				ID:                 1,
				Name:               "Dell PowerEdge R750",
				ShortDescription:   "Стоечный сервер 2U для виртуализации и высоконагруженных сервисов дата-центра. Два процессора Intel Xeon Scalable, до 32 отсеков NVMe, резервируемые блоки питания и горячая замена дисков.",
				Status:             StatusPublished,
				ImageURL:           equipmentMediaBaseURL + "server-rack.jpg",
				VideoURL:           equipmentMediaBaseURL + "server-cpu-vertical.webm",
				PowerConsumptionW:  800,
				HeatDissipationBTU: 2730,
				LikeUserIDs:        []int64{101, 102, 103},
			},
			{
				ID:                 2,
				Name:               "Cisco Catalyst 9300",
				ShortDescription:   "Коммутатор ядра сети дата-центра с поддержкой StackWise-480 и модульным блоком питания. Обеспечивает связность серверных стоек и агрегацию трафика.",
				Status:             StatusPublished,
				ImageURL:           equipmentMediaBaseURL + "switch.jpg",
				VideoURL:           equipmentMediaBaseURL + "switch-panel-vertical.webm",
				PowerConsumptionW:  350,
				HeatDissipationBTU: 1194,
				LikeUserIDs:        []int64{201, 202},
			},
			{
				ID:                 3,
				Name:               "NetApp FAS8300",
				ShortDescription:   "Система хранения данных корпоративного класса с гибридными дисковыми полками. Обслуживает файловый и блочный доступ для серверов дата-центра.",
				Status:             StatusPublished,
				ImageURL:           equipmentMediaBaseURL + "storage.jpg",
				VideoURL:           equipmentMediaBaseURL + "hdd-storage-vertical.webm",
				PowerConsumptionW:  1200,
				HeatDissipationBTU: 4094,
				LikeUserIDs:        []int64{301},
			},
			{
				ID:                 4,
				Name:               "APC Smart-UPS SRT 10000VA",
				ShortDescription:   "Источник бесперебойного питания для серверного зала с двойным преобразованием напряжения и батарейными модулями горячей замены.",
				Status:             StatusDraft,
				ImageURL:           equipmentMediaBaseURL + "ups.jpg",
				VideoURL:           equipmentMediaBaseURL + "fan-cooling-vertical.webm",
				PowerConsumptionW:  900,
				HeatDissipationBTU: 3071,
			},
			{
				ID:                 5,
				Name:               "Stulz CyberAir 3PRO",
				ShortDescription:   "Прецизионный кондиционер для охлаждения серверного зала дата-центра с регулируемой производительностью вентиляторов.",
				Status:             StatusDeleted,
				ImageURL:           equipmentMediaBaseURL + "crac.jpg",
				VideoURL:           equipmentMediaBaseURL + "fan-cooling-vertical.webm",
				PowerConsumptionW:  15000,
				HeatDissipationBTU: 51180,
				LikeUserIDs:        []int64{401, 402, 403, 404},
			},
		},
	}, nil
}

func (r *DatacenterDeviceRepository) ListPublishedDevices(minPowerW int) ([]DatacenterDevice, error) {
	var result []DatacenterDevice
	for _, d := range r.datacenterDevices {
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

func (r *DatacenterDeviceRepository) GetDeviceByID(id int) (DatacenterDevice, error) {
	for _, d := range r.datacenterDevices {
		if d.ID == id && d.Status != StatusDeleted {
			return d, nil
		}
	}
	return DatacenterDevice{}, fmt.Errorf("устройство дата-центра с id %d не найдено", id)
}

func (r *DatacenterDeviceRepository) GetNextPublishedDevice(afterID int) (DatacenterDevice, error) {
	found := false
	for _, d := range r.datacenterDevices {
		if found && d.Status == StatusPublished {
			return d, nil
		}
		if d.ID == afterID {
			found = true
		}
	}
	if !found {
		return DatacenterDevice{}, fmt.Errorf("устройство дата-центра с id %d не найдено", afterID)
	}
	for _, d := range r.datacenterDevices {
		if d.Status == StatusPublished {
			return d, nil
		}
	}
	return DatacenterDevice{}, errors.New("нет опубликованных устройств дата-центра")
}

func (r *DatacenterDeviceRepository) GetDraftDevice() (DatacenterDevice, error) {
	for _, d := range r.datacenterDevices {
		if d.Status == StatusDraft {
			return d, nil
		}
	}
	return DatacenterDevice{}, errors.New("черновик устройства дата-центра не найден")
}

func (r *DatacenterDeviceRepository) FirstPublishedDeviceID() int {
	for _, d := range r.datacenterDevices {
		if d.Status == StatusPublished {
			return d.ID
		}
	}
	return 0
}

func (r *DatacenterDeviceRepository) MaxPowerConsumptionW() int {
	maxPower := 0
	for _, d := range r.datacenterDevices {
		if d.Status == StatusPublished && d.PowerConsumptionW > maxPower {
			maxPower = d.PowerConsumptionW
		}
	}
	return maxPower
}
