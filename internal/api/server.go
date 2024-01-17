package api

import (
	"awesomeProject/internal/app/ds"
	//"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/repository"

	"awesomeProject/internal/app/role"
	"io"
	"log"
	"net/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	//"github.com/golang-jwt/jwt"
	//"strings"
	"strconv"
	"time"
	_ "RIP/docs"

	//"fmt"
	//_ "awesomeProject/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	//"github.com/google/uuid"
	//"github.com/kljensen/snowball/russian"
	//"gorm.io/gorm"
	"github.com/gin-contrib/cors"
	//"encoding/json"
)

type DyeWithColorants struct {
	ds.Dyes
	Colorants []ds.ColorantsAndOtheres
}

func (a *Application) StartServer() {

	log.Println("Server start up")

	r := gin.Default()
	r.LoadHTMLGlob("C:/Program Files/Go/src/RIP/templates/*")

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	r.Use(cors.New(config))
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization")
		c.Next()
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Static("/styles", "./internal/css")
	r.Static("/image", "./resources")
	_ = godotenv.Load()

	AuthGroup := r.Group("/auth")
	{
		AuthGroup.POST("/registration", a.Register)
		AuthGroup.POST("/login", a.Login)
		AuthGroup.GET("/logout", a.Logout)

	}

	r.GET("/list_of_colorants", a.Get_All_Colorant)
	r.GET("/:id", a.Colorant_by_ID)
	//ColorantGroup.Use(a.WithAuthCheck(role.User)).GET("/request/:id", c.GetConsultationsByRequestID)
	r.DELETE("/delete-service/:id", a.WithAuthCheck(role.Moderator), a.DeletionColorant)
	r.PUT("/update_colorants/:id", a.WithAuthCheck(role.Moderator), a.UpdationColorant)
	r.POST("/new_colorant", a.WithAuthCheck(role.Moderator), a.Creation)
	r.POST("/colorant/:id", a.WithAuthCheck(role.User, role.Moderator), a.AddColorantInDye)
	r.POST("/:id/addImage", a.WithAuthCheck(role.Moderator), a.Add_Image)
	r.GET("/list_of_dyes", a.WithAuthCheck(role.User, role.Moderator), a.FilterDyes)
	r.DELETE("/delete-dye/:id", a.WithAuthCheck(role.User, role.Moderator), a.DeletionDye)
	r.PUT("/update_dyes/:id", a.WithAuthCheck(role.Moderator), a.DyeUpdation)
	r.PUT("/update_dyes/:id/put", a.DyeUpdationPrice)
	r.PUT("/formation-dye/:id", a.WithAuthCheck(role.User,role.Moderator), a.Status_User)
	r.PUT("/dyeid/:id/status/:status", a.WithAuthCheck(role.Moderator), a.Status_Moderator)
	r.DELETE("/delete-MtM/:idDye/colorant/:idColorant", a.WithAuthCheck(role.User, role.Moderator), a.DeletionMtM)
	r.PUT("/update_many_to_many/:idDye/colorant/:idColorant", a.WithAuthCheck(role.Moderator), a.UpdationMtM)
	r.GET("/dye/:id", a.WithAuthCheck(role.User, role.Moderator), a.OneOfDyes)
	/*r.POST("/users", func(c *gin.Context) {
		var newUser ds.Users
		c.BindJSON(&newUser)
		err := repo.CreateUser(newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
			return
		}

		users, err := repo.GetAllUsers()
		if err != nil {
			panic("failed to get products from DB")
		}

		c.JSON(http.StatusOK, users)

	})*/
	r.Run()

	log.Println("Server down")
}

// FilterDyes godoc
//
// @Summary Get Dyes
// @Security ApiKeyAuth
// @Description Get all Dyes
// @Tags Dyes
// @ID get-Dyes
// @Produce json
// @Success 200 {object} ds.Dyes
// @Failure 400 {object} ds.Dyes "Некорректный запрос"
// @Failure 404 {object} ds.Dyes "Некорректный запрос"
// @Failure 500 {object} ds.Dyes "Ошибка сервера"
// @Router /Dyes/list_of_dyes [get]
func (a *Application) FilterDyes(c *gin.Context) {

	StartDate := c.Query("StartDate")
	EndDate := c.Query("EndDate")
	status := c.Query("status")
	date1, err1 := time.Parse("2006-01-02", StartDate)
	date2, err2 := time.Parse("2006-01-02", EndDate)
	userID := a.ParseUserID(c)

	if err1 != nil || err2 != nil {
	}
	dyes, err := a.repository.FilterDyesByDateAndStatus(date1, date2, status, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при выполнении запроса к базе данных"})
		return
	}
	c.JSON(http.StatusOK, dyes)
}

func (a *Application) Add_Image(c *gin.Context) {
	AddColorantImage(a.repository, c)
}

// Creation godoc
//
// @Summary create colorant
// @Security ApiKeyAuth
// @Description create colorant
// @Tags Colorants
// @ID create-colorant
// @Accept json
// @Produce json
// @Success 200 {string} string
// @Failure 400 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 404 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 500 {object} ds.ColorantsAndOtheres "Ошибка сервера"
// @Router /new_colorant [post]
func (a *Application) Creation(c *gin.Context) {
	var newService ds.ColorantsAndOtheres
	c.BindJSON(&newService)
	err := a.repository.CreateColorant(newService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	products, err := a.repository.GetAllColorant()
	if err != nil {
		panic("failed to get products from DB")
	}

	c.JSON(http.StatusOK, products)
}

// DeletionColorant godoc
//
// @Summary Удалить краситель по ID
// @Security ApiKeyAuth
// @Description Удалить краситель по его уникальному идентификатору
// @Tags Colorants
// @ID delete-colorant
// @Produce json
// @Param id path string true "ID красителя"
// @Success 200 {object} ds.SuccessResponse
// @Failure 400 {object} ds.ErrorResponse "Некорректный запрос"
// @Failure 404 {object} ds.ErrorResponse "Краситель не найден"
// @Failure 500 {object} ds.ErrorResponse "Ошибка сервера"
// @Router /delete-service/{id} [delete]
func (a *Application) DeletionColorant(c *gin.Context) {
	serviceID := c.Param("id")

	a.repository.DeleteColorant(serviceID)

	filterValue := c.Query("filterValue")
	if filterValue != "" {
		c.Redirect(http.StatusSeeOther, "/home?filterValue="+filterValue)
	} else {
		c.Redirect(http.StatusSeeOther, "/home")
	}
	products, err := a.repository.GetAllColorant()
	if err != nil {
		panic("failed to get products from DB")
	}

	c.JSON(http.StatusOK, products)
}

// Colorant_by_ID godoc
//
// @Summary Получить информацию о красителе по ID
// @Security ApiKeyAuth
// @Description Получить информацию о красителе по его уникальному идентификатору
// @Tags Colorants
// @ID get-colorant-by-id
// @Produce json
// @Param id path string true "ID красителя"
// @Success 200 {object} ds.ColorantsAndOtheres
// @Failure 400 {object} ds.ErrorResponse "Некорректный запрос"
// @Failure 404 {object} ds.ErrorResponse "Краситель не найден"
// @Failure 500 {object} ds.ErrorResponse "Ошибка сервера"
// @Router /{id} [get]
func (a *Application) Colorant_by_ID(c *gin.Context) {
	productName := c.Param("id")
	var product *ds.ColorantsAndOtheres
	product, err := a.repository.GetColorantByID(productName)
	if err != nil {
		panic("failed to get products from DB")
	}
	log.Println(product)
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, product)
}

// Get_All_Colorant godoc
//
// @Summary Get All Colorant
// @Security ApiKeyAuth
// @Description Get all colorants
// @Tags Colorants
// @ID get-colorants
// @Produce json
// @Success 200 {object} ds.ColorantsAndOtheres
// @Failure 400 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 404 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 500 {object} ds.ColorantsAndOtheres "Ошибка сервера"
// @Router /list_of_colorants [get]
func (a *Application) Get_All_Colorant(c *gin.Context) {

	filterValue := c.Query("filterValue")
	userID := a.ParseUserID(c)
	//log.Println(userID)
	products, err := a.repository.FilterColorant(filterValue, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при выполнении запроса к базе данных"})
		return
	}
	c.JSON(http.StatusOK, products)
}

func (a *Application) UpdationColorant(c *gin.Context) {
	serviceID := c.Param("id")
	var newService ds.ColorantsAndOtheres
	c.BindJSON(&newService)
	err := a.repository.UpdateColorant(serviceID, &newService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
		return
	}

	products, err := a.repository.GetAllColorant()
	if err != nil {
		panic("failed to get products from DB")
	}

	c.JSON(http.StatusOK, products)

}

func (a *Application) AddColorantInDye(c *gin.Context) {
	productName := c.Param("id")
	userID := a.ParseUserID(c)
	err := a.repository.CreateDye(productName, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
		return
	}

	dyes, err := a.repository.GetAllDyes()
	if err != nil {
		panic("failed to get dyes from DB")
	}

	c.JSON(http.StatusOK, dyes)

}

func (a *Application) DeletionDye(c *gin.Context) {
	serviceID := c.Param("id")
	userID := a.ParseUserID(c)
	a.repository.DeleteDye(serviceID, userID)

	filterDate := c.Query("filterDate")
	if filterDate != "" {
		c.Redirect(http.StatusSeeOther, "/list_of_dyes?filterDate="+filterDate)
	} else {
		c.Redirect(http.StatusSeeOther, "/list_of_dyes")
	}
	products, err := a.repository.GetAllDyes()
	if err != nil {
		panic("failed to get products from DB")
	}

	c.JSON(http.StatusOK, products)
}

type UpdateDyeRequest struct {
	Price uint `json:"Price"`
	Key   string
}

func (a *Application) DyeUpdationPrice(c *gin.Context) {
	serviceID := c.Param("id")
	var request UpdateDyeRequest

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}

	log.Println("Price")
	log.Println(request.Price)
	if request.Key == "123456" {
		err := a.repository.UpdateDyePrice(serviceID, request.Price)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
			return
		}
		dye, err := a.repository.GetDyeByID(serviceID)
		if err != nil {
			panic("failed to get products from DB")
		}

		c.JSON(http.StatusOK, dye)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Incorrect key"})
		return
	}
}

func (a *Application) DyeUpdation(c *gin.Context) {
	serviceID := c.Param("id")
	var dyes ds.Dyes
	c.BindJSON(&dyes)
	err := a.repository.UpdateDye(serviceID, &dyes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
		return
	}

	dye, err := a.repository.GetAllDyes()
	if err != nil {
		panic("failed to get products from DB")
	}

	c.JSON(http.StatusOK, dye)
}

func (a *Application) Status_User(c *gin.Context) {
	serviceID := c.Param("id")
	var User []ds.Users
	User, err := a.repository.GetAllUsers()
	if err != nil {
		panic("failed to get users from DB")
	}
	found := false
	userID := a.ParseUserID(c)
	for _, user := range User {
		if user.ID_User == userID {
			
				found = true
				break
			
		}
	}

	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Неверный статус пользователя"})
		return
	}
	a.repository.StatusUser(serviceID, userID)
	dyes, err := a.repository.GetDyeByID(serviceID)
	if err != nil {
		panic("failed to get products from DB")
	}
	c.JSON(http.StatusOK, dyes)
}

func (a *Application) Status_Moderator(c *gin.Context) {
	serviceID := c.Param("id")
	Status := c.Param("status")
	var User []ds.Users
	User, err := a.repository.GetAllUsers()
	if err != nil {
		panic("failed to get users from DB")
	}
	found := false
	userID := a.ParseUserID(c)
	for _, user := range User {
		if user.ID_User == userID {
			if user.Role == role.Moderator {
				found = true
				break
			}
		}
	}

	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Неверный статус пользователя"})
		return
	}
	a.repository.StatusModerator(serviceID, userID, Status)
	dyes, err := a.repository.GetDyeByID(serviceID)
	if err != nil {
		panic("failed to get products from DB")
	}
	c.JSON(http.StatusOK, dyes)
}

func (a *Application) DeletionMtM(c *gin.Context) {
	DyeID := c.Param("idDye")
	ColorantId := c.Param("idColorant")
	err := a.repository.DeleteMtM(DyeID, ColorantId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}

	Mtm1, err := a.repository.GetAllMtM()
	if err != nil {
		panic("failed to get products from DB")
	}

	c.JSON(http.StatusOK, Mtm1)
}

func (a *Application) UpdationMtM(c *gin.Context) {
	DyeID := c.Param("idDye")
	ColorantId := c.Param("idColorant")
	var new_many_to_many ds.Dye_Colorants
	c.BindJSON(&new_many_to_many)
	err := a.repository.UpdateManytoMany(DyeID, ColorantId, &new_many_to_many)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
		return
	}

	products, err := a.repository.GetAllMtM()
	if err != nil {
		panic("failed to get products from DB")
	}
	c.JSON(http.StatusOK, products)
}

func (a *Application) OneOfDyes(c *gin.Context) {
	productName := c.Param("id")
	dye, err := a.repository.GetDyeByID(productName)
	if err != nil {
		panic("failed to get products from DB")
	}
	//c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, dye)
}

func AddColorantImage(repository *repository.Repository, c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	if id < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"Status":  "Failed",
			"Message": "неверное значение id",
		})
		return
	}
	image, err := c.FormFile("image")
	if err != nil || image == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image"})
		return
	}

	// Чтение содержимого изображения в байтах
	file, err := image.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при открытии"})
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения"})
		return
	}
	// Получение Content-Type из заголовков запроса
	contentType := image.Header.Get("Content-Type")

	// Вызов функции репозитория для добавления изображения
	err = repository.AddColorantImage(id, imageBytes, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image uploaded successfully"})

}
