package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kljensen/snowball/russian"
)


type ServiceProduct struct {
	Name        string
	Image       string
	Link        string
	Description string
	Properties  string
}

func StartServer() {
	log.Println("Server start up")

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	data := []ServiceProduct{
		{
			Name:        "Диоксид титана",
			Image:       "https://static.tildacdn.com/stor3836-6631-4636-b938-336534356332/69807646.jpg",
			Description: "Чистый диоксид титана — бесцветные кристаллы (желтеет при нагревании). Для технических целей применяется в раздробленном состоянии, представляя собой белый порошок. Не растворяется в воде и разбавленных минеральных кислотах (за исключением плавиковой).",
			Properties:  "Плотность: 4,235 г/см3\n/nТемпература плавления: 1843°C Температура кипения: 2972°C<br> Растворимость в воде: не растворяется<br>Температура разложения: 2900°C",
			Link:        "/product/TiO2",
		},
		{
			Name:        "Оксид(III) железа",
			Image:       "https://sc04.alicdn.com/kf/Hacb20d18e6274e53a197e9f380825000K.jpg",
			Description: "Амфотерный оксид с большим преобладанием основных свойств. Красно-коричневого цвета. Термически устойчив до высоких температур. Образуется при сгорании железа на воздухе. Не реагирует с водой. Медленно реагирует с кислотами и щелочами.",
			Properties:  "Плотность:	5,242 г/см3 Температура плавления:	1566°C Растворимость в воде:	нерастворим",
			Link:        "/product/FeO2",
		},
		{
			Name:        "Оксид(III) хрома",
			Image:       "https://ddek.ru/assets/images/products/225/prasino-tsimentou.jpg",
			Description: "Амфотерный оксид с большим преобладанием основных свойств. Красно-коричневого цвета. Термически устойчив до высоких температур. Образуется при сгорании железа на воздухе. Не реагирует с водой. Медленно реагирует с кислотами и щелочами.",
			Properties:  "Плотность:	5,242 г/см3 Температура плавления:	1566°C Растворимость в воде:	нерастворим",
			Link:        "/product/Cr2O3",
		},
		{
			Name:        "Оксид(V) ванадия",
			Image:       "https://chem.ru/uploads/posts/2020-04/medium/1586977299_oksid-vanadija-v.jpg",
			Description: "Амфотерный оксид с большим преобладанием основных свойств. Красно-коричневого цвета. Термически устойчив до высоких температур. Образуется при сгорании железа на воздухе. Не реагирует с водой. Медленно реагирует с кислотами и щелочами.",
			Properties:  "Плотность:	5,242 г/см3 Температура плавления:	1566°C Растворимость в воде:	нерастворим",
			Link:        "/product/V2O5",
		},
		{
			Name:        "Оксид(IV) марганца",
			Image:       "https://img.promportal.su/foto/good_fotos/49789/497896103/oksid-marganca-iv-dioksid-marganca-mno2_foto_largest.jpg",
			Description: "Амфотерный оксид с большим преобладанием основных свойств. Красно-коричневого цвета. Термически устойчив до высоких температур. Образуется при сгорании железа на воздухе. Не реагирует с водой. Медленно реагирует с кислотами и щелочами.",
			Properties:  "Плотность:	5,242 г/см3 Температура плавления:	1566°C Растворимость в воде:	нерастворим",
			Link:        "/product/MnO2",
		},
	}

	r.Static("/styles", "./internal/css")
	r.Static("/image", "./resources")

	r.GET("/home", func(c *gin.Context) {

		filterValue := c.Query("filterValue")

		var filteredServices []ServiceProduct

		if filterValue != "" {
			filterValueNormalized := russian.Stem(filterValue, false)

			for _, service := range data {
				serviceNameNormalized := russian.Stem(service.Name, false)
				if strings.Contains(strings.ToLower(serviceNameNormalized), strings.ToLower(filterValueNormalized)) {
					filteredServices = append(filteredServices, service)
				}
			}
		} else {
			filteredServices = data
		}

		log.Println(filteredServices)

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":    "Производство красок",
			"services": filteredServices,
			"filterValue":filterValue,
		})
	})

	r.GET("/product/:name", func(c *gin.Context) {
		productName := c.Param("name")

		var product ServiceProduct
		for _, p := range data {
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

	r.Run()

	log.Println("Server down")
}
