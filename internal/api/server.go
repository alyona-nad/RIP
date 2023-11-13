package api

import (
	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/repository"
	"io"
	"log"
	"net/http"

	//"strings"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	//"github.com/kljensen/snowball/russian"
	//"gorm.io/gorm"
	"github.com/gin-contrib/cors"
)

//...

func singleton() uint {
	var user uint
	user = 3
	return user
}

type DyeWithColorants struct {
	ds.Dyes
	Colorants []ds.ColorantsAndOtheres
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
	// Чтение изображения из запроса
	/*image, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image"})
		return
	}
	*/
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

func StartServer() {

	log.Println("Server start up")

	r := gin.Default()
	r.LoadHTMLGlob("C:/Program Files/Go/src/RIP/templates/*")

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	r.Use(cors.New(config))

	r.Static("/styles", "./internal/css")
	r.Static("/image", "./resources")
	_ = godotenv.Load()

	repo, err := repository.New(dsn.FromEnv())
	if err != nil {
		panic("failed to connect database")
	}
	r.GET("/list_of_colorants", func(c *gin.Context) {

		filterValue := c.Query("filterValue")

		products, err := repo.FilterColorant(filterValue, singleton())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при выполнении запроса к базе данных"})
			return
		}

		/*c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"services":    products,
			"filterValue": filterValue,
		})*/
		c.JSON(http.StatusOK, products)
	})

	r.GET("/product/:id", func(c *gin.Context) {
		productName := c.Param("id")
		var product *ds.ColorantsAndOtheres
		product, err = repo.GetColorantByID(productName)
		if err != nil {
			panic("failed to get products from DB")
		}
		log.Println(product)
		//c.HTML(http.StatusOK, "product.tmpl", product)
		c.JSON(http.StatusOK, product)
	})

	r.DELETE("/delete-service/:id", func(c *gin.Context) {
		serviceID := c.Param("id")

		repo.DeleteColorant(serviceID)

		filterValue := c.Query("filterValue")
		if filterValue != "" {
			c.Redirect(http.StatusSeeOther, "/home?filterValue="+filterValue)
		} else {
			c.Redirect(http.StatusSeeOther, "/home")
		}
		products, err := repo.GetAllColorant()
		if err != nil {
			panic("failed to get products from DB")
		}

		c.JSON(http.StatusOK, products)

	})

	r.POST("/services", func(c *gin.Context) {
		var newService ds.ColorantsAndOtheres
		c.BindJSON(&newService)
		err := repo.CreateColorant(newService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
			return
		}

		products, err := repo.GetAllColorant()
		if err != nil {
			panic("failed to get products from DB")
		}

		c.JSON(http.StatusOK, products)

	})

	r.PUT("/update_services/:id", func(c *gin.Context) {
		serviceID := c.Param("id")
		var newService ds.ColorantsAndOtheres
		c.BindJSON(&newService)
		err := repo.UpdateColorant(serviceID, &newService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
			return
		}

		products, err := repo.GetAllColorant()
		if err != nil {
			panic("failed to get products from DB")
		}

		c.JSON(http.StatusOK, products)

	})

	r.GET("/list_of_dyes", func(c *gin.Context) {

		StartDate := c.Query("StartDate")
		EndDate := c.Query("EndDate")
		status := c.Query("status")
		date1, err1 := time.Parse("2006-01-02", StartDate)
		date2, err2 := time.Parse("2006-01-02", EndDate)
		if err1 != nil || err2 != nil {
		}
		dyes, err := repo.FilterDyesByDateAndStatus(date1, date2, status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при выполнении запроса к базе данных"})
			return
		}

		/*c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"services":    filteredServices,
			"filterValue": filterValue,
		})*/
		c.JSON(http.StatusOK, dyes)
	})

	r.GET("/dye/:id", func(c *gin.Context) {
		productName := c.Param("id")
		dye, err := repo.GetDyeByID(productName)
		if err != nil {
			panic("failed to get products from DB")
		}
		c.JSON(http.StatusOK, dye)
	})

	r.DELETE("/delete-dye/:id", func(c *gin.Context) {
		serviceID := c.Param("id")

		repo.DeleteDye(serviceID, singleton())

		filterDate := c.Query("filterDate")
		if filterDate != "" {
			c.Redirect(http.StatusSeeOther, "/list_of_dyes?filterDate="+filterDate)
		} else {
			c.Redirect(http.StatusSeeOther, "/list_of_dyes")
		}
		products, err := repo.GetAllDyes()
		if err != nil {
			panic("failed to get products from DB")
		}

		c.JSON(http.StatusOK, products)

	})

	r.POST("/colorant/:id", func(c *gin.Context) {
		productName := c.Param("id")
		err := repo.CreateDye(productName, singleton())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create service"})
			return
		}

		dyes, err := repo.GetAllDyes()
		if err != nil {
			panic("failed to get dyes from DB")
		}

		c.JSON(http.StatusOK, dyes)

	})

	r.PUT("/formation-dye/:id", func(c *gin.Context) {
		serviceID := c.Param("id")
		var User []ds.Users
		User, err := repo.GetAllUsers()
		if err != nil {
			panic("failed to get users from DB")
		}
		found := false
		for _, user := range User {
			if user.ID_User == singleton() {
				if user.Role == "Пользователь" {
					found = true
					break
				}
			}
		}

		if !found {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Неверный статус пользователя"})
			return
		}
		repo.StatusUser(serviceID, singleton())
		dyes, err := repo.GetDyeByID(serviceID)
		if err != nil {
			panic("failed to get products from DB")
		}
		c.JSON(http.StatusOK, dyes)

	})

	r.PUT("/dyeid/:id/status/:status", func(c *gin.Context) {
		serviceID := c.Param("id")
		Status := c.Param("status")
		var User []ds.Users
		User, err := repo.GetAllUsers()
		if err != nil {
			panic("failed to get users from DB")
		}
		found := false
		for _, user := range User {
			if user.ID_User == singleton() {
				if user.Role == "Модератор" {
					found = true
					break
				}
			}
		}

		if !found {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Неверный статус пользователя"})
			return
		}
		repo.StatusModerator(serviceID, singleton(), Status)
		dyes, err := repo.GetDyeByID(serviceID)
		if err != nil {
			panic("failed to get products from DB")
		}
		c.JSON(http.StatusOK, dyes)

	})

	r.PUT("/update_dyes/:id", func(c *gin.Context) {
		serviceID := c.Param("id")
		var dyes ds.Dyes
		c.BindJSON(&dyes)
		err := repo.UpdateDye(serviceID, &dyes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
			return
		}

		dye, err := repo.GetAllDyes()
		if err != nil {
			panic("failed to get products from DB")
		}

		c.JSON(http.StatusOK, dye)

	})

	r.POST("/users", func(c *gin.Context) {
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

	})

	r.PUT("/update_many_to_many/:idDye/colorant/:idColorant", func(c *gin.Context) {
		DyeID := c.Param("idDye")
		ColorantId := c.Param("idColorant")
		var new_many_to_many ds.Dye_Colorants
		c.BindJSON(&new_many_to_many)
		err := repo.UpdateManytoMany(DyeID, ColorantId, &new_many_to_many)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update service"})
			return
		}

		products, err := repo.GetAllMtM()
		if err != nil {
			panic("failed to get products from DB")
		}

		c.JSON(http.StatusOK, products)

	})

	r.DELETE("/delete-MtM/:idDye/colorant/:idColorant", func(c *gin.Context) {
		DyeID := c.Param("idDye")
		ColorantId := c.Param("idColorant")
		err := repo.DeleteMtM(DyeID, ColorantId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
			return
		}

		Mtm1, err := repo.GetAllMtM()
		if err != nil {
			panic("failed to get products from DB")
		}

		c.JSON(http.StatusOK, Mtm1)
	})
	r.PUT("/colorants/:id/addImage", func(c *gin.Context) {
		AddColorantImage(repo, c)
	})

	r.Run()

	log.Println("Server down")
}
