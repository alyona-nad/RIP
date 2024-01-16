package repository

import (
	"awesomeProject/internal/app/ds"
	"strconv"
	"time"
	"github.com/kljensen/snowball/russian"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"awesomeProject/internal/MinioClient"
	//"github.com/gin-gonic/gin"
	//"awesomeProject/internal/app/ds/jwt.go"
	//"encoding/json"
	//"net/http"
	//"github.com/golang-jwt/jwt"
	//"fmt"
	//"github.com/google/uuid"
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
	if colorants.Image == "" {
        colorants.Image = "http://localhost:9000/rip1/defoult.jpg"
    }
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
	Colorants []ds.ColorantsAndOtheres
	Dyes uint
}

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
	if (len(dyeWithColorants.Colorants)==0) {
		err=r.db.Exec("UPDATE dyes SET status = ? WHERE id_dye = ? and status=?", "удалено", id, "Действует").Error
	}
	return dyeWithColorants, nil
	//return dye, nil
}

func (r *Repository) DeleteDye(id string, idUser uint) error {
	var User ds.Users
	err := r.db.Where("id_user = ? AND Role = ?", idUser, 1/*"Пользователь"*/).First(&User).Error
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

/*func (r *Repository) CreateDye(idcolorant string, idUser uint) error {

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
*/
/*func (r *Repository) UpdateDye(id string, dye *ds.Dyes) error {
	err := r.db.Model(&dye).Where("id_dye = ?", id).Updates(dye).Error
	return err
}*/
func (r *Repository) UpdateDye(id string, dye *ds.Dyes) error {
    err := r.db.Where("id_dye = ?", id).Updates(dye).Error
    return err
}

func (r *Repository) UpdateDyePrice(id string, price uint) error {
	/*newPrice, err := strconv.ParseUint(price,10,0)
    if err != nil {
        return err
    }*/

    // Используйте Find для поиска красителя по ID
    var dye ds.Dyes
    if err := r.db.First(&dye, "id_dye = ?", id).Error; err != nil {
        return err
    }

    // Обновление цены
    dye.Price = price
	//dye.Price = uint(newPrice)
    // Сохранение изменений в базе данных
    if err := r.db.Save(&dye).Error; err != nil {
        return err
    }

    return nil
}

func (r *Repository) StatusUser(id string, idUser uint) error {
	var User ds.Users
	err := r.db.Where("id_user = ? AND Role = ?", idUser, 1/*"Пользователь"*/).First(&User).Error
	if err != nil {
		panic("Неверный статус пользователя")
	} else {
		return r.db.Exec("UPDATE dyes SET status = ?, formation_date= ? WHERE id_dye = ? and status=?", "Сформирован", time.Now(), id, "Действует").Error
	}
}


func (r *Repository) StatusModerator(id string, idUser uint, status string) error {
	var User ds.Users
	err := r.db.Where("id_user = ? AND Role = ?", idUser, 2/*"Модератор"*/).First(&User).Error
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

func (r *Repository) FilterColorant(name string,id uint) (/*[]ds.ColorantsAndOtheres*/Colorants, error) {//черновик
	var colorant []ds.ColorantsAndOtheres
	if name != "" {
		filterValueNormalized := russian.Stem(name, false)
	
	if err := r.db.Where("name ILIKE ? and status = ?", "%"+filterValueNormalized+"%","Действует").Find(&colorant).Error; err != nil {
		panic("failed to get products from DB")
	}
	}  else {
		if err := r.db.Where("status = ?","Действует").Find(&colorant).Error; err != nil {
			panic("failed to get products from DB")
		}
	}


		var DyesIDs uint
			r.db.Table("dyes").
			Where("user_id = ? and status=?", id,"Действует").
			Pluck("id_dye", &DyesIDs)
			
			ColorantsDyes := Colorants{
			Dyes: DyesIDs,
			Colorants: colorant,
	}

	return ColorantsDyes, nil
	//return colorant, nil
}
func (r *Repository) FilterDyesByDateAndStatus(date1, date2 time.Time, status string, id uint) ([]DyeWithColorants, error) {
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
	var User ds.Users
	
	err1 := r.db.Where("id_user = ? AND Role = ?", id, 2/*"Модератор"*/).First(&User).Error
	if err1 != nil {
		query = query.Where("user_id = ?", id)
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


func (r *Repository) Register(user *ds.Users) error {
	return r.db.Create(user).Error
}

func (r *Repository) GetUserByLogin(login string) (*ds.Users, error) {
	user := &ds.Users{
		Login: login,
	}

	err := r.db.Where(user).First(user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) DeleteActiveRequest(userID uint) error {

	dye := &ds.Dyes{}
	err := r.db.Find(dye, "status = 'Действует' AND user_id = ?", userID).Error
	if err != nil {
		return err
	}

	return r.db.Exec("UPDATE requests SET status = 'Удалено' WHERE id=? AND user_id=?", dye.ID_Dye, userID).Error

}








