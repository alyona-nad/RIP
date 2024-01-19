package api

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"golang.org/x/crypto/bcrypt"
	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/role"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type registerReq struct {
	Name        string `json:"name"`
	Login       string `json:"login"`
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
	Password    string `json:"pass"`
}

type registerResp struct {
	Ok          bool      `json:"ok"`
	AccessToken string    `json:"access_token"`
	Role        role.Role `json:"role"`
}


// @Summary Регистрация
// @Description Регистрация
// @Tags auth
// @Accept json
// @Produce json
// @Param input body ds.Users true "Информация о пользователе"
// @Success 200 {object} registerResp
// @Router /auth/registration [post]
func (a *Application) Register(gCtx *gin.Context) {
	req := &registerReq{}
	cfg := a.config
	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	/*if req.Password == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("password is empty"))
		return
	}

	if req.Login == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("login is empty"))
		return
	}

	if req.Name == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("name is empty"))
		return
	}

	if req.Email == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("email is empty"))
		return
	}

	if req.PhoneNumber == "" {
		gCtx.AbortWithError(http.StatusBadRequest, fmt.Errorf("phone number is empty"))
		return
	}*/

	err = a.repository.Register(&ds.Users{
		Role:        role.User,
		Name:        req.Name,
		Login:       req.Login,
		Email_Address:       req.Email,
		Phone: req.PhoneNumber,
		Password:   generateHashString(req.Password), // пароли делаем в хешированном виде и далее будем сравнивать хеши, чтобы их не угнали с базой вместе
	})
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	user, err := a.repository.GetUserByLogin(req.Login)
	fmt.Println(user)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	cfg.JWT.SigningMethod = jwt.SigningMethodHS256
	cfg.JWT.ExpiresIn = time.Hour
	token := jwt.NewWithClaims(cfg.JWT.SigningMethod, &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(cfg.JWT.ExpiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "bitop-admin",
		},
		UserID: user.ID_User, // test uuid
		Role:   user.Role,
	})
	if token == nil {
		gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("token is nil"))
		return
	}

	strToken, err := token.SignedString([]byte(cfg.JWT.Token))
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cant create str token"))
		return
	}

	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    strToken,
		Expires:  time.Now().Add(cfg.JWT.ExpiresIn),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	gCtx.Header("Authorization", "Bearer "+strToken)

	http.SetCookie(gCtx.Writer, cookie)

	gCtx.JSON(http.StatusOK, &registerResp{
		Ok:          true,
		AccessToken: strToken,
		Role:        user.Role,
	})
}

func generateHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	//hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	//if err!=nil{}
	//return string(hashedPassword)
	return hex.EncodeToString(h.Sum(nil))
}

/*func checkPassword(password, hashedPassword string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}
// @Summary Logout
// @Security ApiKeyAuth
// @Description Выход из аккаунта
// @Tags auth
// @Produce json
// @Success 200 {string} string
// @Router /auth/logout [get]
func (a *Application) Logout(gCtx *gin.Context) {
	// получаем заголовок
	jwtStr := gCtx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) { // если нет префикса то нас дурят!
		gCtx.AbortWithStatus(http.StatusBadRequest) // отдаем что нет доступа

		return // завершаем обработку
	}

	// отрезаем префикс
	jwtStr = jwtStr[len(jwtPrefix):]

	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.config.JWT.Token), nil
	})
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)

		return
	}

	// сохраняем в блеклист редиса
	err = a.redis.WriteJWTToBlacklist(gCtx.Request.Context(), jwtStr, a.config.JWT.ExpiresIn)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)

		return
	}
	userID := a.ParseUserID(gCtx)
	a.repository.DeleteActiveDye(userID)
	
	gCtx.Status(http.StatusOK)
}

type loginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResp struct {
	ExpiresIn   int       `json:"expires_in"`
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	Role        role.Role `json:"role"`
}

// @Summary Вход в аккаунт
// @Description Логин
// @Tags auth
// @Accept json
// @Produce json
// @Param input body loginReq true "login info"
// @Success 200 {object} loginResp
// @Router /auth/login [post]
func (a *Application) Login(gCtx *gin.Context) {
	cfg := a.config
	req := &loginReq{}

	err := json.NewDecoder(gCtx.Request.Body).Decode(req)
	if err != nil {
		gCtx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := a.repository.GetUserByLogin(req.Login)
	fmt.Println(user)
	if err != nil {
		gCtx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	 fmt.Println(generateHashString(req.Password))
	fmt.Println(user.Password)
	// fmt.Println(req.Login)
	// fmt.Println(user.Login)
	//if req.Login == user.Login && checkPassword(req.Password, user.Password){
	if req.Login == user.Login && user.Password == generateHashString(req.Password) {
		// значит проверка пройдена
		// генерируем ему jwt
		cfg.JWT.SigningMethod = jwt.SigningMethodHS256
		cfg.JWT.ExpiresIn = time.Hour
		token := jwt.NewWithClaims(cfg.JWT.SigningMethod, &ds.JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(cfg.JWT.ExpiresIn).Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "bitop-admin",
			},
			UserID: user.ID_User, // test uuid
			Role:   user.Role,
		})
		if token == nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("token is nil"))
			return
		}

		strToken, err := token.SignedString([]byte(cfg.JWT.Token))
		if err != nil {
			gCtx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cant create str token"))
			return
		}

		cookie := &http.Cookie{
			Name:     "access_token",
			Value:    strToken,
			Expires:  time.Now().Add(cfg.JWT.ExpiresIn),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		}
		gCtx.Header("Authorization", "Bearer "+strToken)

		http.SetCookie(gCtx.Writer, cookie)

		gCtx.JSON(http.StatusOK, loginResp{
			ExpiresIn:   int(cfg.JWT.ExpiresIn.Seconds()),
			AccessToken: strToken,
			TokenType:   "Bearer",
			Role:        user.Role,
		})
		return
	}
	fmt.Println("Response 1:")
	gCtx.AbortWithStatus(http.StatusForbidden) // отдаем 403 ответ в знак того что доступ запрещен
}
