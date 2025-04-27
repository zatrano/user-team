package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type ModelError string

func (e ModelError) Error() string {
	return string(e)
}

const (
	ErrSystemUserHasTeam        ModelError = "sistem kullanıcısının (system user) bir takımı olamaz (TeamID NULL olmalı)"
	ErrUserMissingTeam          ModelError = "yönetici (manager) veya temsilci (agent) kullanıcısının bir takımı olmalı (TeamID boş olamaz)"
	ErrInvalidUserType          ModelError = "geçersiz kullanıcı tipi (UserType)"
	ErrPasswordCannotBeEmpty    ModelError = "şifre boş olamaz"
	ErrInvalidUpdateTypeField   ModelError = "güncelleme verisinde geçersiz 'type' alanı tipi"
	ErrInvalidUpdateTeamIDField ModelError = "güncelleme verisinde geçersiz 'team_id' alanı tipi"
)

type UserType string

const (
	System  UserType = "system"
	Manager UserType = "manager"
	Agent   UserType = "agent"
)

func (UserType) GormDataType() string {
	return "user_type"
}
func (UserType) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	if db.Dialector.Name() == "postgres" {
		return "user_type"
	}
	return "varchar(10)"
}

type User struct {
	gorm.Model
	Name     string   `gorm:"size:100;not null;index"`
	Account  string   `gorm:"size:100;unique;not null"`
	Password string   `gorm:"size:255;not null"`
	Status   bool     `gorm:"default:true;index"`
	Type     UserType `gorm:"type:user_type;not null;default:'agent';index"`
	TeamID   *uint    `gorm:"index"`
	Team     *Team    `gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.Password == "" {
		return ErrPasswordCannotBeEmpty
	}
	hashed, bcryptErr := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if bcryptErr != nil {
		return bcryptErr
	}
	u.Password = string(hashed)

	validTypes := map[UserType]bool{System: true, Manager: true, Agent: true}
	if _, typeIsValid := validTypes[u.Type]; !typeIsValid {
		return ErrInvalidUserType
	}

	if u.Type == System {
		if u.TeamID != nil {
			return ErrSystemUserHasTeam
		}
	} else if u.Type == Manager || u.Type == Agent {
		if u.TeamID == nil {
			return ErrUserMissingTeam
		}
	}

	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	var userType UserType
	var teamID *uint
	knownType := false
	knownTeamID := false
	currentUserType := u.Type
	currentTeamID := u.TeamID

	if tx.Statement.Dest != nil {
		if destMap, ok := tx.Statement.Dest.(map[string]interface{}); ok {
			if t, exists := destMap["type"]; exists {
				knownType = true
				if typeStr, okStr := t.(string); okStr {
					userType = UserType(typeStr)
				} else if typeVal, okType := t.(UserType); okType {
					userType = typeVal
				} else {
					return ErrInvalidUpdateTypeField
				}
			}
			if tid, exists := destMap["team_id"]; exists {
				knownTeamID = true
				if tid == nil {
					teamID = nil
				} else if tidPtr, okPtr := tid.(*uint); okPtr {
					teamID = tidPtr
				} else if tidVal, okUint := tid.(uint); okUint {
					tempID := tidVal
					teamID = &tempID
				} else if tidFloat, okFloat := tid.(float64); okFloat {
					tempID := uint(tidFloat)
					teamID = &tempID
				} else {
					return ErrInvalidUpdateTeamIDField
				}
			}
		}
	}

	if !knownType {
		userType = currentUserType
	}
	if !knownTeamID {
		teamID = currentTeamID
	}

	validTypes := map[UserType]bool{System: true, Manager: true, Agent: true}
	if _, typeIsValid := validTypes[userType]; !typeIsValid {
		if string(userType) != "" {
			return ErrInvalidUserType
		}
	}

	if userType == System {
		if teamID != nil {
			return ErrSystemUserHasTeam
		}
	}

	return nil
}

func (u *User) SetPassword(password string) error {
	if password == "" {
		return ErrPasswordCannotBeEmpty
	}
	hashed, bcryptErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if bcryptErr != nil {
		return bcryptErr
	}
	u.Password = string(hashed)
	return nil
}
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
func (u *User) IsManager() bool {
	return u.Type == Manager
}
