package ds

import (
	"time"
)

type User struct {
	ID_User       uint `gorm:"primaryKey"`
	Name          string
	Phone         string `gorm:"unique"`
	Email_Address string `gorm:"unique"`
	Password      string
	Role          string
}

type Dye struct {
	ID_Dye         uint `gorm:"primaryKey"`
	User_ID        uint
	User           User `gorm:"foreignKey:User_ID"`
	Name           string
	Status         string
	CreationDate   time.Time
	FormationDate  time.Time
	CompletionDate time.Time
	Moderator      uint
	ModeratorUser  User       `gorm:"foreignKey:Moderator"`
	Colorants      []Colorant `gorm:"many2many:Dye_Colorants;"`
}
type Colorant struct {
	ID_Colorant int64 `gorm:"primaryKey"`
	Name        string
	Image       string
	Link        string
	Description string
	Properties  string
	Status      string
}

type Dye_Colorants struct {
	ID_DyeColorant uint `gorm:"primaryKey"`
	Dye_ID         uint
	DyeColorant    Dye `gorm:"foreignKey:Dye_ID"`
	Colorant_ID    uint
	ColorantDye    Colorant `gorm:"foreignKey:Colorant_ID"`
}
