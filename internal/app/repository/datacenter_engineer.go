package repository

import (
	"datacenter-power/internal/app/ds"
)

func (r *Repository) DatacenterEngineerLoginExists(login string) (bool, error) {
	var count int64
	err := r.db.Model(&ds.DatacenterEngineer{}).Where("login = ?", login).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) CreateDatacenterEngineer(engineer *ds.DatacenterEngineer) error {
	return r.db.Create(engineer).Error
}

func (r *Repository) GetDatacenterEngineerByLogin(login string) (ds.DatacenterEngineer, error) {
	engineer := ds.DatacenterEngineer{}
	err := r.db.Where("login = ?", login).First(&engineer).Error
	if err != nil {
		return ds.DatacenterEngineer{}, err
	}
	return engineer, nil
}
