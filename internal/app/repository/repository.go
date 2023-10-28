package repository

import (
	"awesomeProject/internal/app/ds"
	"strconv"
	"time"
	"github.com/kljensen/snowball/russian"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"awesomeProject/internal/MinioClient"
)

type Repository struct {
	db *gorm.DB
	minioClient *minioclient.MinioClient
}
type ServiceProduct struct {
	ID_Colorant int64
	Name        string
	Image       string
	Link        string
	Description string
	Properties  string
	Status      string
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
minioClient, err :=minioclient.NewMinioClient()
if err != nil {
	return nil, err
}
	return &Repository{
		db: db,
		minioClient: minioClient,
	}, nil
}

func (r *Repository) GetAllColorant() ([]ds.ColorantsAndOtheres, error) {
	var colorants []ds.ColorantsAndOtheres
	err := r.db.Table("colorants_and_otheres").Select("id_colorant, name, image, description, properties,status").Where("status = ?", "Действует").Scan(&colorants).Error
	if err != nil {
		return nil, err
	}

	return colorants, nil
}

func (r *Repository) GetColorantByID(id string) (*ds.ColorantsAndOtheres, error) {
	colorants := &ds.ColorantsAndOtheres{}

	err := r.db.First(colorants, "id_colorant = ?", id).Error
	if err != nil {
		return nil, err
	}

	return colorants, nil
}

func (r *Repository) DeleteColorant(id string) error {
	return r.db.Exec("UPDATE colorants_and_otheres SET status = ? WHERE id_colorant = ?", "удалено", id).Error
}

func (r *Repository) CreateColorant(colorants ds.ColorantsAndOtheres) error {
	err := r.db.Table("colorants_and_otheres").Create(&colorants).Error
	return err
}

func (r *Repository) UpdateColorant(id string, colorants *ds.ColorantsAndOtheres) error {
	err := r.db.Model(&colorants).Where("id_colorant = ?", id).Updates(colorants).Error
	return err
}

type DyeWithColorants struct {
	*ds.Dyes
	Colorants []ds.ColorantsAndOtheres
}
type Colorants struct {
	*ds.ColorantsAndOtheres
	Dyes []uint
}

// func (r *Repository) GetAllDyes() ([]ds.Dyes, error) {
func (r *Repository) GetAllDyes() ([]DyeWithColorants, error) {
	var dyes []ds.Dyes
	err := r.db.Preload("User").Preload("ModeratorUser").Find(&dyes).Where("status = ?", "Действует").Scan(&dyes).Error
	if err != nil {
		return nil, err
	}
	var dyeWithColorants []DyeWithColorants
	for i := range dyes {
		r.db.Preload("User").Preload("ModeratorUser").Find(&dyes[i])

		var colorantIDs []uint
		r.db.Table("dye_colorants").
			Where("id_dye = ?", dyes[i].ID_Dye).
			Pluck("id_colorant", &colorantIDs)

		var colorants []ds.ColorantsAndOtheres
		r.db.Where("id_colorant IN ?", colorantIDs).Find(&colorants)
		dyeWithColorant := DyeWithColorants{
			Dyes:      &dyes[i],
			Colorants: colorants,
		}
		dyeWithColorants = append(dyeWithColorants, dyeWithColorant)

		//dyes[i].Colorants = colorants
	}
	//return dyes, nil
	return dyeWithColorants, nil
}

// func (r *Repository) GetDyeByID(id string) (ds.Dyes, error) {
func (r *Repository) GetDyeByID(id string) (DyeWithColorants, error) {
	dyes := &ds.Dyes{}

	err := r.db.First(dyes, "id_dye = ?", id).Error
	if err != nil {
		DWC := &DyeWithColorants{}
		DWC = nil
		return *DWC, err
	}

	r.db.Preload("User").Preload("ModeratorUser").Find(&dyes)

	var colorantIDs []uint
	r.db.Table("dye_colorants").
		Where("id_dye = ?", dyes.ID_Dye).
		Pluck("id_colorant", &colorantIDs)

	var colorants []ds.ColorantsAndOtheres
	r.db.Where("id_colorant IN ?", colorantIDs).Find(&colorants)
	//dyes.Colorants = colorants
	//r.db.Model(&dyes).Association("ID_Dye").Append(colorants)
	dyeWithColorants := DyeWithColorants{
		Dyes:      dyes,
		Colorants: colorants,
	}
	return dyeWithColorants, nil
	//return dye, nil
}

func (r *Repository) DeleteDye(id string, idUser uint) error {
	var User ds.Users
	err := r.db.Where("id_user = ? AND Role = ?", idUser, "Пользователь").First(&User).Error
	if err != nil {
		panic("Неверный статус пользователя")
	} else {
		return r.db.Exec("UPDATE dyes SET status = ? WHERE id_dye = ?", "удалено", id).Error
	}
}

func (r *Repository) CreateDye(idcolorant string, idUser uint) error {

	var dye ds.Dyes
	var dye_colorants ds.Dye_Colorants
	err := r.db.Where("User_ID = ? AND Status = ?", idUser, "Действует").First(&dye).Error
	id, err1 := strconv.Atoi(idcolorant)
	if err1 != nil {
		panic("failed to get products from DB")
	}
	if err != nil {
		newDye := ds.Dyes{
			User_ID:      idUser,
			Status:       "Действует",
			Name:         "Гуашь",
			CreationDate: time.Now(),
			Moderator:    idUser,
		}
		if err := r.db.Create(&newDye).Error; err != nil {
			return err
		}
		dye_colorants.ID_Dye = newDye.ID_Dye
		dye_colorants.ID_Colorant = uint(id)
		//dye_colorants.Percent_Content=percent

	} else {
		dye_colorants.ID_Dye = dye.ID_Dye
		dye_colorants.ID_Colorant = uint(id)
		//dye_colorants.Percent_Content=percent
	}
	err = r.db.Table("dye_colorants").Create(&dye_colorants).Error
	return nil
}

func (r *Repository) UpdateDye(id string, dye *ds.Dyes) error {
	err := r.db.Model(&dye).Where("id_dye = ?", id).Updates(dye).Error
	return err
}

func (r *Repository) StatusUser(id string, idUser uint) error {
	var User ds.Users
	err := r.db.Where("id_user = ? AND Role = ?", idUser, "Пользователь").First(&User).Error
	if err != nil {
		panic("Неверный статус пользователя")
	} else {
		return r.db.Exec("UPDATE dyes SET status = ?, formation_date= ? WHERE id_dye = ? and status=?", "Сформирован", time.Now(), id, "Действует").Error
	}
}

/*func (r *Repository) StatusModeratorEnd(id string,idUser string) error {
	var User ds.Users
	err := r.db.Where("id_user = ? AND Role = ?", idUser, "Модератор").First(&User).Error
	if err != nil {
		panic("Неверный статус пользователя")
	} else {
	return r.db.Exec("UPDATE dyes SET status = ?, completion_date= ? WHERE id_dye = ? and status=?", "Завершён",time.Now(), id, "Сформирован").Error
	}
}

func (r *Repository) StatusModeratorReject(id string,idUser string) error {
	var User ds.Users
	err := r.db.Where("id_user = ? AND Role = ?", idUser, "Модератор").First(&User).Error
	if err != nil {
		panic("Неверный статус пользователя")
	} else {
	return r.db.Exec("UPDATE dyes SET status = ? WHERE id_dye = ? and status=?", "Отклонено", id, "Сформирован").Error
	}
}*/

func (r *Repository) StatusModerator(id string, idUser uint, status string) error {
	var User ds.Users
	err := r.db.Where("id_user = ? AND Role = ?", idUser, "Модератор").First(&User).Error
	if err != nil {
		panic("Неверный статус пользователя")
	} else {
		if status == "reject" {
			return r.db.Exec("UPDATE dyes SET status = ? WHERE id_dye = ? and status=?", "Отклонено", id, "Сформирован").Error
		} else {
			return r.db.Exec("UPDATE dyes SET status = ?, completion_date= ? WHERE id_dye = ? and status=?", "Завершён", time.Now(), id, "Сформирован").Error
		}
	}
}

func (r *Repository) GetAllUsers() ([]ds.Users, error) {
	var users []ds.Users
	err := r.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *Repository) GetUserByID(id string) (*ds.Users, error) {
	users := &ds.Users{}

	err := r.db.First(users, "id_user = ?", id).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) CreateUser(user ds.Users) error {
	err := r.db.Table("users").Create(&user).Error
	return err
}

func (r *Repository) UpdateUser(id string, user *ds.Users) error {
	err := r.db.Model(&user).Where("id_user = ?", id).Updates(user).Error
	return err
}

func (r *Repository) UpdateManytoMany(idDye string, idColorant string, dye *ds.Dye_Colorants) error {
	err := r.db.Model(&dye).Where("id_dye = ? and id_colorant=?", idDye, idColorant).Updates(dye).Error
	return err
}

func (r *Repository) GetAllMtM() ([]ds.Dye_Colorants, error) {
	var dye_colorant []ds.Dye_Colorants
	//err := r.db.Find(&dye_colorant).Error
	err := r.db.Preload("DyeColorant").Preload("ColorantDye").Find(&dye_colorant).Error
	if err != nil {
		return nil, err
	}
	for i := range dye_colorant {
		r.db.Preload("User").Preload("ModeratorUser").Find(&dye_colorant[i].DyeColorant)
		r.db.Preload("User").Preload("ModeratorUser").Find(&dye_colorant[i].ColorantDye)
		var colorantIDs []uint
		r.db.Table("dye_colorants").
			Where("id_dye = ?", dye_colorant[i].DyeColorant.ID_Dye).
			Pluck("id_colorant", &colorantIDs)

		var colorants []ds.ColorantsAndOtheres
		r.db.Where("id_colorant IN ?", colorantIDs).Find(&colorants)
		//dye_colorant[i].DyeColorant.Colorants = colorants
		//r.db.Model(&dye_colorant[i].DyeColorant).Association("ColorantDye").Append(colorants)
	}

	return dye_colorant, nil
}

func (r *Repository) DeleteMtM(idDye string, idColorant string) error {
	return r.db.Where("id_dye = ? and id_colorant = ?", idDye, idColorant).Delete(&ds.Dye_Colorants{}).Error
}

func (r *Repository) FilterDyesByDateAndStatus(date1, date2 time.Time, status string) ([]DyeWithColorants, error) {
	var dyeWithColorants []DyeWithColorants
	var dyes []ds.Dyes
	query := r.db.
		Preload("User").
		Preload("ModeratorUser")

	if !date1.IsZero() {
		query = query.Where("formation_date >= ?", date1)
	}

	if !date2.IsZero() {
		query = query.Where("formation_date <= ?", date2)
	}

	if status != "" {
		query = query.Where("Status = ?", status)
	}

	err := query.Find(&dyes).Error

	if err != nil {
		return nil, err
	}

	for i := range dyes {
		r.db.Preload("User").Preload("ModeratorUser").Find(&dyes[i])

		var colorantIDs []uint
		r.db.Table("dye_colorants").
			Where("id_dye = ?", dyes[i].ID_Dye).
			Pluck("id_colorant", &colorantIDs)

		var colorants []ds.ColorantsAndOtheres
		r.db.Where("id_colorant IN ?", colorantIDs).Find(&colorants)
		dyeWithColorant := DyeWithColorants{
			Dyes:      &dyes[i],
			Colorants: colorants,
		}
		dyeWithColorants = append(dyeWithColorants, dyeWithColorant)
	}

	return dyeWithColorants, nil
}


func (r *Repository) FilterColorant(name string) (/*[]ds.ColorantsAndOtheres*/[]Colorants, error) {//черновик
	var ColorantsDyes []Colorants
	var colorant []ds.ColorantsAndOtheres
	if name != "" {
		filterValueNormalized := russian.Stem(name, false)

	if err := r.db.Where("name ILIKE ?", "%"+filterValueNormalized+"%").Find(&colorant).Error; err != nil {
		panic("failed to get products from DB")
	}
	}  else {
		if err := r.db.Find(&colorant).Error; err != nil {
			panic("failed to get products from DB")
		}
	}

	for i := range colorant {
		//r.db.Preload("User").Preload("ModeratorUser").Find(&dyes[i])

		var DyesIDs []uint
		r.db.Table("colorants_and_otheres").
			Where("id_colorant = ?", colorant[i].ID_Colorant).
			Pluck("id_dye", &DyesIDs)

		ColorantDye := Colorants{
			Dyes: DyesIDs,
			ColorantsAndOtheres: &colorant[i],
		}
		ColorantsDyes = append(ColorantsDyes, ColorantDye)
	}

	return ColorantsDyes, nil
	//return colorant, nil
}

func (r *Repository) AddColorantImage(colorantID int, imageBytes []byte, contentType string) error {
    err := r.minioClient.RemoveServiceImage(colorantID)
    if err != nil {
        return err
    }

    imageURL, err := r.minioClient.UploadServiceImage(colorantID, imageBytes, contentType)
    if err != nil {
        return err
    }
	
	err =r.db.Exec("UPDATE colorants_and_otheres SET image = ? WHERE id_colorant = ?", imageURL, colorantID).Error
    
    if err != nil {
        return err
    }

    return nil
}





