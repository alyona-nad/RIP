package api

import (
	"log"
	"net/http"
	"strings"

	"awesomeProject/internal/app/dsn"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kljensen/snowball/russian"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceProduct struct {
	ID          int64
	Name        string
	Image       string
	Link        string
	Description string
	Properties  string
	Status      string
}

func GetProductsFromDB(db *gorm.DB) ([]ServiceProduct, error) {
	var products []ServiceProduct
	var products1 []int64
	err := db.Table("colorants").Select("id_colorant, name, image, description, properties,status").Where("status = ?", "Действует").Scan(&products).Error
	if err != nil {
		return nil, err
	}
	err1 := db.Table("colorants").Select("id_colorant").Where("status = ?", "Действует").Scan(&products1).Error
	for i, id := range products1 {
		products[i].ID = id
	}
	if err1 != nil {
		return nil, err1
	}
	return products, nil
}

func StartServer() {
	log.Println("Server start up")

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.Static("/styles", "./internal/css")
	r.Static("/image", "./resources")
	_ = godotenv.Load()

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	r.GET("/home", func(c *gin.Context) {

		products, err := GetProductsFromDB(db)
		if err != nil {
			panic("failed to get products from DB")
		}

		filterValue := c.Query("filterValue")

		var filteredServices []ServiceProduct
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

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":       "Производство красок",
			"services":    filteredServices,
			"filterValue": filterValue,
		})
	})

	r.GET("/product/:name", func(c *gin.Context) {
		productName := c.Param("name")
		products, err := GetProductsFromDB(db)
		if err != nil {
			panic("failed to get products from DB")
		}

		var product ServiceProduct
		for _, p := range products {
			if p.Name == productName {
				product = p
				break
			}
		}
		c.HTML(http.StatusOK, "product.tmpl", gin.H{
			"title":       "Производство красок",
			"Name":        product.Name,
			"Properties":  product.Properties,
			"Image":       product.Image,
			"Description": product.Description,
		})
	})

	r.POST("/delete-service/:id", func(c *gin.Context) {
		serviceID := c.Param("id")
		var count int64
		db.Table("colorants").Where("id_colorant = ?", serviceID).Count(&count)

		if count == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		result := db.Exec("UPDATE colorants SET status = ? WHERE id_colorant = ?", "удалено", serviceID)
		if result.Error != nil {
			log.Println("Failed to delete service:", result.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		log.Println("Service deleted successfully")

		filterValue := c.Query("filterValue")
		if filterValue != "" {
			c.Redirect(http.StatusSeeOther, "/home?filterValue="+filterValue)
		} else {
			c.Redirect(http.StatusSeeOther, "/home")
		}

	})

	r.Run()

	log.Println("Server down")
}
