package ds

import (
	"time"
	"awesomeProject/internal/app/role"
)

type Users struct {
	ID_User       uint `gorm:"primaryKey"`
	Name          string
	Login       string    `json:"login"`
	Phone         string `gorm:"unique"`
	Email_Address string `gorm:"unique"`
	Password      string
	Role        role.Role
}

type Dyes struct {
	ID_Dye         uint `gorm:"primaryKey"`
	User_ID        uint
	User           Users `gorm:"foreignKey:User_ID"`
	Name           string
	Status         string
	CreationDate   time.Time
	FormationDate  time.Time
	CompletionDate time.Time
	Moderator      uint
	ModeratorUser  Users                 `gorm:"foreignKey:Moderator"`
	Price uint
}
type ColorantsAndOtheres struct {
	ID_Colorant int64 `gorm:"primaryKey;autoIncrement"`
	Name        string
	Image       string
	Description string
	Properties  string
	Status      string
}

type Dye_Colorants struct {
	ID_Dye          uint 
	DyeColorant     Dyes `gorm:"primaryKey;foreignKey:ID_Dye"`
	ID_Colorant     uint 
	ColorantDye     ColorantsAndOtheres `gorm:"primaryKey;foreignKey:ID_Colorant"`
	Percent_Content float64
}
