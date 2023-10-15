package api

import (
	"log"
	"net/http"

	//"strconv"
	"strings"

	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/dsn"
	"awesomeProject/internal/app/repository"

	//"awesomeProject/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kljensen/snowball/russian"

	//"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceProduct struct {
	ID_Colorant int64
	Name        string
	Image       string
	Link        string
	Description string
	Properties  string
	Status      string
}

type  adddye struct {
	ID_User uint
	//percent string
}

func GetProductsFromDB(db *gorm.DB) ([]ServiceProduct, error) {
	var products []ServiceProduct
	err := db.Table("colorants_and_otheres").Select("id_colorant, name, image, description, properties,status").Where("status = ?", "Действует").Scan(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func StartServer() {
	log.Println("Server start up")

	r := gin.Default()
	r.LoadHTMLGlob("C:/Program Files/Go/src/RIP/templates/*")

	r.Static("/styles", "./internal/css")
	r.Static("/image", "./resources")
	_ = godotenv.Load()

	repo, err := repository.New(dsn.FromEnv())
	if err != nil {
		panic("failed to connect database")
	}
	r.GET("/home", func(c *gin.Context) {

		var products []ds.ColorantsAndOtheres
		products, err := repo.GetAllColorant()
		if err != nil {
			panic("failed to get products from DB")
		}
		filterValue := c.Query("filterValue")

		var filteredServices []ds.ColorantsAndOtheres
		if filterValue != "" {
			filterValueNormalized := russian.Stem(filterValue, false)

			for _, service := range products {
				serviceNameNormalized := russian.Stem(service.Name, false)
				if strings.Contains(strings.ToLower(serviceNameNormalized), strings.ToLower(filterValueNormalized)) {
					filteredServices = append(filteredServices, service)
				}
			}
		} else {
			filteredServices = products
		}

		/*c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"services":    filteredServices,
			"filterValue": filterValue,
		})*/
		c.JSON(http.StatusOK, filteredServices)
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

	r.POST("/delete-service/:id", func(c *gin.Context) {
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

	r.POST("/update_services/:id", func(c *gin.Context) {
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

	r.GET("/home_dyes", func(c *gin.Context) {

		dyes, err := repo.GetAllDyes()
		if err != nil {
			panic("failed to get products from DB")
		}
		var filteredServices []ds.Dyes
		filterDate := c.Query("filterDate")
        if filterDate == "" {
        filteredServices = dyes
    }

    for _, dye := range dyes {
        creationDate := dye.CreationDate.Format("2006-01-02")
        if creationDate == filterDate {
            filteredServices = append(filteredServices, dye)
        }
    }
		

		/*c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"services":    filteredServices,
			"filterValue": filterValue,
		})*/
		c.JSON(http.StatusOK, filteredServices)
	})

	r.GET("/dye/:id", func(c *gin.Context) {
		productName := c.Param("id")
		var dye *ds.Dyes
		dye, err = repo.GetDyeByID(productName)
		if err != nil {
			panic("failed to get products from DB")
		}
		c.JSON(http.StatusOK, dye)
	})

	r.POST("/delete-dye/:id", func(c *gin.Context) {
		serviceID := c.Param("id")

		repo.DeleteDye(serviceID)

		filterDate := c.Query("filterDate")
		if filterDate != "" {
			c.Redirect(http.StatusSeeOther, "/home_dyes?filterDate="+filterDate)
		} else {
			c.Redirect(http.StatusSeeOther, "/home_dyes")
		}
		products, err := repo.GetAllDyes()
		if err != nil {
			panic("failed to get products from DB")
		}

		c.JSON(http.StatusOK, products)

	})

	r.POST("/colorant/:id", func(c *gin.Context) {
		productName := c.Param("id")
		var usercolorant adddye
		c.BindJSON(&usercolorant)
		err := repo.CreateDye(productName,usercolorant.ID_User)
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

	r.POST("/formation-dye/:id/user/:id1", func(c *gin.Context) {
		serviceID := c.Param("id")
		UserID := c.Param("id1")
		repo.StatusUser(serviceID,UserID)
		dyes, err := repo.GetDyeByID(serviceID)
		if err != nil {
			panic("failed to get products from DB")
		}
		c.JSON(http.StatusOK, dyes)

	})

	r.POST("/completion-dye/:id/user/:id1", func(c *gin.Context) {
		serviceID := c.Param("id")
		UserID := c.Param("id1")
		repo.StatusModeratorEnd(serviceID,UserID)
		dyes, err := repo.GetDyeByID(serviceID)
		if err != nil {
			panic("failed to get products from DB")
		}
		c.JSON(http.StatusOK, dyes)

	})

	r.POST("/rejection-dye/:id/user/:id1", func(c *gin.Context) {
		serviceID := c.Param("id")
		UserID := c.Param("id1")
		repo.StatusModeratorReject(serviceID,UserID)
		dyes, err := repo.GetDyeByID(serviceID)
		if err != nil {
			panic("failed to get products from DB")
		}
		c.JSON(http.StatusOK, dyes)

	})

	r.POST("/update_dyes/:id", func(c *gin.Context) {
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

	r.POST("/update_many_to_many/:idDye/colorant/:idColorant", func(c *gin.Context) {
		DyeID := c.Param("idDye")
		ColorantId:= c.Param("idColorant")
		var new_many_to_many ds.Dye_Colorants
		c.BindJSON(&new_many_to_many)
		err := repo.UpdateManytoMany(DyeID,ColorantId, &new_many_to_many)
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
		ColorantId:= c.Param("idColorant")
		err := repo.DeleteMtM(DyeID,ColorantId)
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

	r.Run()

	log.Println("Server down")
}
