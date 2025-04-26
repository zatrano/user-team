package models

import (
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	Name   string `gorm:"size:100;not null;index"`
	Status bool   `gorm:"default:true;index"`
	Agents []User `gorm:"foreignKey:TeamID;references:ID"`
}

func (t *Team) Manager(db *gorm.DB) (*User, error) {
	var manager User
	err := db.Where("team_id = ? AND type = ?", t.ID, "manager").First(&manager).Error
	if err != nil {
		return nil, err
	}
	return &manager, nil
}
