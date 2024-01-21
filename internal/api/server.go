package api

import (
	_ "RIP/docs"
	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/repository"
	"awesomeProject/internal/app/role"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:7000"}
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
		AuthGroup.POST("/logout", a.Logout)

	}

	r.GET("/list_of_colorants", a.Get_All_Colorant)
	r.GET("/:id", a.Colorant_by_ID)
	r.DELETE("/delete-service/:id", a.WithAuthCheck(role.Moderator), a.DeletionColorant)
	r.PUT("/update_colorants/:id", a.WithAuthCheck(role.Moderator), a.UpdationColorant)
	r.POST("/new_colorant", a.WithAuthCheck(role.Moderator), a.Creation)
	r.POST("/colorant/:id", a.WithAuthCheck(role.User, role.Moderator), a.AddColorantInDye)
	r.POST("/:id/addImage", a.WithAuthCheck(role.Moderator), a.Add_Image)

	r.GET("/list_of_dyes", a.WithAuthCheck(role.User, role.Moderator), a.FilterDyes)
	r.GET("/dye/:id", a.WithAuthCheck(role.User, role.Moderator), a.OneOfDyes)
	r.DELETE("/delete-dye/:id", a.WithAuthCheck(role.User, role.Moderator), a.DeletionDye)
	/*r.PUT("/update_dyes/:id", a.WithAuthCheck(role.Moderator), a.DyeUpdation)*/
	r.PUT("/update_dyes/:id/put", a.DyeUpdationPrice)
	r.PUT("/formation-dye/:id", a.WithAuthCheck(role.User, role.Moderator), a.Status_User)
	r.PUT("/dyeid/:id/status/:status", a.WithAuthCheck(role.Moderator), a.Status_Moderator)

	r.DELETE("/delete-MtM/:idDye/colorant/:idColorant", a.WithAuthCheck(role.User, role.Moderator), a.DeletionMtM)
	r.PUT("/update_many_to_many/:idDye/colorant/:idColorant", a.WithAuthCheck(role.Moderator), a.UpdationMtM)

	r.Run()

	log.Println("Server down")
}

// @Summary Вывести список заявок
// @Security ApiKeyAuth
// @Description Вывести список заявок
// @Tags Dyes
// @Produce json
// @Success 200 {object} ds.Dyes
// @Failure 400 {object} ds.Dyes "Некорректный запрос"
// @Failure 404 {object} ds.Dyes "Некорректный запрос"
// @Failure 500 {object} ds.Dyes "Ошибка сервера"
// @Router /list_of_dyes [get]
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

// @Summary Добавить изображение к красителю
// @Security ApiKeyAuth
// @Description Добавить изображение к красителю
// @Tags Colorants
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID красителя"
// @Param image formData file true "Файл изображения"
// @Success 200 {string} string
// @Failure 400 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 404 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 500 {object} ds.ColorantsAndOtheres "Ошибка сервера"
// @Router /{id}/addImage [post]
func (a *Application) Add_Image(c *gin.Context) {
	AddColorantImage(a.repository, c)
}

// @Summary Создать краситель
// @Security ApiKeyAuth
// @Description Создать краситель
// @Tags Colorants
// @Accept json
// @Produce json
// @Param input body ds.ColorantsAndOtheres true "Информация о красителе"
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

// @Summary Удалить краситель
// @Security ApiKeyAuth
// @Description Удалить краситель
// @Tags Colorants
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID красителя"
// @Success 200 {string} string
// @Failure 400 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 404 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 500 {object} ds.ColorantsAndOtheres "Ошибка сервера"
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

// @Summary Показать краситель по ID
// @Description Показать краситель по ID
// @Tags Colorants
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID красителя"
// @Success 200 {object} ds.ColorantsAndOtheres
// @Failure 400 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 404 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 500 {object} ds.ColorantsAndOtheres "Ошибка сервера"
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

// @Summary Вывести список красителей
// @Description Вывести список красителей
// @Tags Colorants
// @Produce json
// @Success 200 {object} ds.ColorantsAndOtheres
// @Failure 400 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 404 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 500 {object} ds.ColorantsAndOtheres "Ошибка сервера"
// @Router /list_of_colorants [get]
func (a *Application) Get_All_Colorant(c *gin.Context) {

	filterValue := c.Query("filterValue")
	userID := a.ParseUserID(c)
	products, err := a.repository.FilterColorant(filterValue, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при выполнении запроса к базе данных"})
		return
	}
	c.JSON(http.StatusOK, products)
}

// @Summary Обновить краситель
// @Security ApiKeyAuth
// @Description Обновить краситель
// @Tags Colorants
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID красителя"
// @Param input body ds.ColorantsAndOtheres true "Информация о красителе"
// @Success 200 {string} string
// @Failure 400 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 404 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 500 {object} ds.ColorantsAndOtheres "Ошибка сервера"
// @Router /update_colorants/{id} [put]
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

// @Summary Добавить краситель в заявку
// @Security ApiKeyAuth
// @Description Добавить краситель в заявку
// @Tags Colorants
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID красителя"
// @Success 200 {string} string
// @Failure 400 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 404 {object} ds.ColorantsAndOtheres "Некорректный запрос"
// @Failure 500 {object} ds.ColorantsAndOtheres "Ошибка сервера"
// @Router /colorant/{id} [post]
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

// @Summary Удалить заявку
// @Security ApiKeyAuth
// @Description Удалить заявку
// @Tags Dyes
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID заявки"
// @Success 200 {string} string
// @Failure 400 {object} ds.Dyes "Некорректный запрос"
// @Failure 404 {object} ds.Dyes "Некорректный запрос"
// @Failure 500 {object} ds.Dyes "Ошибка сервера"
// @Router /delete-dye/{id} [delete]
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

/*
// @Summary Обновить заявку
// @Security ApiKeyAuth
// @Description Обновить заявку
// @Tags Dyes
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID заявки"
// @Param input body ds.Dyes true "Информация о заявке"
// @Success 200 {string} string
// @Failure 400 {object} ds.Dyes "Некорректный запрос"
// @Failure 404 {object} ds.Dyes "Некорректный запрос"
// @Failure 500 {object} ds.Dyes "Ошибка сервера"
// @Router /update_dyes/{id} [put]
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
}*/

// @Summary Сформировать заявку
// @Security ApiKeyAuth
// @Description Сформировать заявку
// @Tags Dyes
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID заявки"
// @Success 200 {string} string
// @Failure 400 {object} ds.Dyes "Некорректный запрос"
// @Failure 404 {object} ds.Dyes "Некорректный запрос"
// @Failure 500 {object} ds.Dyes "Ошибка сервера"
// @Router /formation-dye/{id} [put]
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

// @Summary Обновить статус модератором
// @Security ApiKeyAuth
// @Description Обновить статус модератором
// @Tags Dyes
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID заявки"
// @Param        status   path      string  true  "Статус"
// @Param input body ds.StatusData true "Информация о статусе"
// @Success 200 {string} string
// @Failure 400 {object} ds.Dyes "Некорректный запрос"
// @Failure 404 {object} ds.Dyes "Некорректный запрос"
// @Failure 500 {object} ds.Dyes "Ошибка сервера"
// @Router /dyeid/{id}/status/{status} [put]
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

// @Summary Удалить краситель из заявки
// @Security ApiKeyAuth
// @Description Удалить краситель из заявки
// @Tags Colorant-Dye
// @Accept       json
// @Produce      json
// @Param        idColorant   path      int  true  "ID красителя"
// @Param        idDye   path      int  true  "ID заявки"
// @Success 200 {string} string "Краситель был удален из заявки"
// @Failure 400 {string} string "Некорректный запрос"
// @Failure 404 {string} string "Некорректный запрос"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /delete-MtM/{idDye}/colorant/{idColorant} [delete]
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

	contentType := image.Header.Get("Content-Type")

	err = repository.AddColorantImage(id, imageBytes, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image uploaded successfully"})

}
