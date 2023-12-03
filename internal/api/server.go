package api

import (
	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/dsn"
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
	//"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	//"github.com/google/uuid"
	//"github.com/kljensen/snowball/russian"
	//"gorm.io/gorm"
	"github.com/gin-contrib/cors"
	//"encoding/json"
)

//...

/*func singleton() uint {
	var user uint
	user = 3
	return user
}*/

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

func (a *Application)StartServer() {

	log.Println("Server start up")

	r := gin.Default()
	r.LoadHTMLGlob("C:/Program Files/Go/src/RIP/templates/*")

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	r.Use(cors.New(config))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Static("/styles", "./internal/css")
	r.Static("/image", "./resources")
	_ = godotenv.Load()

	repo, err := repository.New(dsn.FromEnv())
	if err != nil {
		panic("failed to connect database")
	}

	AuthGroup := r.Group("/auth")
	{
		AuthGroup.POST("/registration", a.Register)
		AuthGroup.POST("/login", a.Login)
		AuthGroup.GET("/logout", a.Logout)

	}
	//ColorantGroup := r.Group("/colorants")
	
		listofroleswithuser:=[2]role.Role{role.User,"l"}
		listofroleswithmoderator:=[2]role.Role{role.Moderator,"l"}
		//listofroles:=[2]role.Role{role.User,role.Moderator}
		r.GET("/list_of_colorants",func(c *gin.Context) {

			filterValue := c.Query("filterValue")
			userID := a.ParseUserID(c)
			products, err := repo.FilterColorant(filterValue, userID/*singleton()*/)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при выполнении запроса к базе данных"})
				return
			}
			c.JSON(http.StatusOK, products)
		})
		r.GET("/:id", func(c *gin.Context) {
			productName := c.Param("id")
			var product *ds.ColorantsAndOtheres
			product, err = repo.GetColorantByID(productName)
			if err != nil {
				panic("failed to get products from DB")
			}
			log.Println(product)
			c.JSON(http.StatusOK, product)
		})
		//ColorantGroup.Use(a.WithAuthCheck(role.User)).GET("/request/:id", c.GetConsultationsByRequestID)
		r.Use(a.WithAuthCheck(listofroleswithmoderator)).DELETE("/delete-service/:id", func(c *gin.Context) {
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
		r.Use(a.WithAuthCheck(listofroleswithmoderator)).PUT("/update_colorants/:id", func(c *gin.Context) {
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
		r.Use(a.WithAuthCheck(listofroleswithmoderator)).POST("/new_colorant", func(c *gin.Context) {
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
		r.Use(a.WithAuthCheck(listofroleswithuser)).POST("/colorant/:id", func(c *gin.Context) {
			productName := c.Param("id")
			userID := a.ParseUserID(c)
			err := repo.CreateDye(productName, userID)
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
		r.Use(a.WithAuthCheck(listofroleswithmoderator)).POST("/:id/addImage", func(c *gin.Context) {
			AddColorantImage(repo, c)
		})
	

	
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
		r.Use(a.WithAuthCheck(listofroleswithmoderator)).GET("/list_of_dyes", func(c *gin.Context) {

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
			c.JSON(http.StatusOK, dyes)
		})
		r.Use(a.WithAuthCheck(listofroleswithuser)).DELETE("/delete-dye/:id", func(c *gin.Context) {
			serviceID := c.Param("id")
			userID := a.ParseUserID(c)
			repo.DeleteDye(serviceID, userID)
	
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
		r.Use(a.WithAuthCheck(listofroleswithmoderator)).PUT("/update_dyes/:id", func(c *gin.Context) {
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
		r.Use(a.WithAuthCheck(listofroleswithuser)).PUT("/formation-dye/:id", func(c *gin.Context) {
			serviceID := c.Param("id")
			var User []ds.Users
			User, err := repo.GetAllUsers()
			if err != nil {
				panic("failed to get users from DB")
			}
			found := false
			userID := a.ParseUserID(c)
			for _, user := range User {
				if user.ID_User == userID {
					if user.Role == role.User {
						found = true
						break
					}
				}
			}
	
			if !found {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Неверный статус пользователя"})
				return
			}
			repo.StatusUser(serviceID, userID)
			dyes, err := repo.GetDyeByID(serviceID)
			if err != nil {
				panic("failed to get products from DB")
			}
			c.JSON(http.StatusOK, dyes)
	
		})
		r.Use(a.WithAuthCheck(listofroleswithmoderator)).PUT("/dyeid/:id/status/:status", func(c *gin.Context) {
			serviceID := c.Param("id")
			Status := c.Param("status")
			var User []ds.Users
			User, err := repo.GetAllUsers()
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
			repo.StatusModerator(serviceID, userID, Status)
			dyes, err := repo.GetDyeByID(serviceID)
			if err != nil {
				panic("failed to get products from DB")
			}
			c.JSON(http.StatusOK, dyes)
	
		})

	
		r.Use(a.WithAuthCheck(listofroleswithuser)).DELETE("/delete-MtM/:idDye/colorant/:idColorant", func(c *gin.Context) {
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
		r.Use(a.WithAuthCheck(listofroleswithmoderator)).PUT("/update_many_to_many/:idDye/colorant/:idColorant", func(c *gin.Context) {
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
		r.GET("/dye/:id", func(c *gin.Context) {
			productName := c.Param("id")
			dye, err := repo.GetDyeByID(productName)
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
		r.Run()

	log.Println("Server down")
}

	
	/*r.GET("/list_of_colorants", func(c *gin.Context) {

		filterValue := c.Query("filterValue")

		products, err := repo.FilterColorant(filterValue, singleton())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при выполнении запроса к базе данных"})
			return
		}

		/*c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"services":    products,
			"filterValue": filterValue,
		})
		c.JSON(http.StatusOK, products)
	})

	r.GET("/colorants/:id", func(c *gin.Context) {
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

	r.POST("/new_colorant", func(c *gin.Context) {
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

	r.PUT("/update_colorants/:id", func(c *gin.Context) {
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
		})
		c.JSON(http.StatusOK, dyes)
	})
*/
	
/*
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
*/
/*	r.POST("/colorant/:id", func(c *gin.Context) {
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
*/
	/*r.PUT("/formation-dye/:id", func(c *gin.Context) {
		serviceID := c.Param("id")
		var User []ds.Users
		User, err := repo.GetAllUsers()
		if err != nil {
			panic("failed to get users from DB")
		}
		found := false
		for _, user := range User {
			if user.ID_User == singleton() {
				if user.Role == role.User {
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

	})*/

/*	r.PUT("/dyeid/:id/status/:status", func(c *gin.Context) {
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
		repo.StatusModerator(serviceID, singleton(), Status)
		dyes, err := repo.GetDyeByID(serviceID)
		if err != nil {
			panic("failed to get products from DB")
		}
		c.JSON(http.StatusOK, dyes)

	})*/

/*	r.PUT("/update_dyes/:id", func(c *gin.Context) {
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

	})*/

	

	/*r.PUT("/update_many_to_many/:idDye/colorant/:idColorant", func(c *gin.Context) {
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
	})*/
	/*r.PUT("/colorants/:id/addImage", func(c *gin.Context) {
		AddColorantImage(repo, c)
	})

	r.POST("/login", func(c *gin.Context)  {
		Login(c *gin.Context)
	})*/

	
/*var secretKey = []byte("your-secret-key")

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Claims struct {
	UserID       string   `json:"user_id"`
	Scopes       []string `json:"scopes"`
	jwt.StandardClaims
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var credentials Credentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Проверка логина и пароля (пример)
	if credentials.Login == "login" && credentials.Password == "check123" {
		expirationTime := time.Now().Add(1 * time.Hour)
		claims := &Claims{
			UserID: "7d13fa21-89f5-41bb-a6ec-1c6112d4a3d1", // пример ID пользователя
			Scopes: []string{"read", "write"},
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "bitop-admin",
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(secretKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"expires_in":  expirationTime.Unix(),
			"access_token": tokenString,
			"token_type":   "Bearer",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Проверка наличия токена в заголовке Authorization
	tokenString := extractToken(r.Header.Get("Authorization"))
	if tokenString == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Валидация токена
	_, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	/*claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		w.WriteHeader(http.StatusForbidden)
		return
	}
*/
// В этом месте вы можете проверить, что у пользователя есть необходимые права (scopes)
// ...

/*	response := map[string]interface{}{
		"status": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func extractToken(authorizationHeader string) string {
	parts := strings.Split(authorizationHeader, "Bearer ")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}
type loginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResp struct {
	ExpiresIn   time.Duration `json:"expires_in"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
}
type JWT struct {
	SigningMethod jwt.SigningMethod
	ExpiresIn     time.Duration
	Token         string
}
func Login(gCtx *gin.Context) {
	req := &loginReq{}
	login := "login"
	password := "password"
	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	signingMethod := jwt.SigningMethodHS256
	expiresIn := time.Hour * 24
	secretKey := "test"
	if req.Login == login && req.Password == password {
		// значит проверка пройдена
		// генерируем ему jwt
		token := jwt.NewWithClaims(signingMethod, &ds.JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(expiresIn).Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "bitop-admin",
			},
			UserUUID: uuid.New(), // test uuid
			Scopes:   []string{}, // test data
		})

		if token == nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("token is nil"))
			return
		}

		strToken, err := token.SignedString([]byte(secretKey))
		if err != nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cant create str token"))
			return
		}

		gCtx.JSON(http.StatusOK, loginResp{
			ExpiresIn:   expiresIn,
			AccessToken: strToken,
			TokenType:   "Bearer",
		})
		return
	}

	gCtx.AbortWithStatus(http.StatusForbidden)
}


/*var signingKey = []byte("your-secret-key")

// User структура, представляющая пользователя
type User struct {
	Username string
	Password string
}

// Credentials структура для передачи данных в запросе на вход
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CustomClaims пользовательские утверждения для JWT
type CustomClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Имитация базы данных пользователей
var users = map[string]User{
	"john": User{"john", "password123"},
	"jane": User{"jane", "password456"},
}

// Создание токена при входе
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	/*if err := decodeJSON(r.Body, &creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, exists := users[creds.Username]
	if !exists || user.Password != creds.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(creds.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"token": "%s"}`, token)))
}

// Защищенный маршрут, который требует валидный токен
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := extractToken(r)
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := validateToken(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"message": "Hello, %s!"}`, claims.Username)))
}

// Функция для генерации токена
func generateToken(username string) (string, error) {
	claims := CustomClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Токен действителен в течение 24 часов
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

// Функция для проверки валидности токена
func validateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("Invalid token")
}

// Извлечение токена из заголовка Authorization
func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(bearerToken) > 7 && bearerToken[:7] == "Bearer " {
		return bearerToken[7:]
	}
	return ""
}*/

// Обработка запросов JSON
/*func decodeJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}*/
// @Summary      Показывает список красителей
// @Produce      json
// @Success      200  {object}
// @Router       /list_of_colorants [get]
