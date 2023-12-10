package api

import (
	"awesomeProject/internal/app/ds"
	"awesomeProject/internal/app/role"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
)

const jwtPrefix = "Bearer "

func (a *Application) WithAuthCheck(assignedRoles ... /*[2]*/ role.Role) func(ctx *gin.Context) {
	return func(gCtx *gin.Context) {
		jwtStr := gCtx.GetHeader("Authorization")
		fmt.Println(assignedRoles)
		fmt.Println("rfr")
		fmt.Println(jwtStr)
		if !strings.HasPrefix(jwtStr, jwtPrefix) { // если нет префикса то нас дурят!
			fmt.Println("ПЛОХО 1")
			gCtx.AbortWithStatus(http.StatusForbidden) // отдаем что нет доступа

			return // завершаем обработку
		}

		// отрезаем префикс
		jwtStr = jwtStr[len(jwtPrefix):]
		fmt.Println(jwtStr)

		err := a.redis.CheckJWTInBlacklist(gCtx.Request.Context(), jwtStr)
		if err == nil { // значит что токен в блеклисте
			gCtx.AbortWithStatus(http.StatusForbidden)

			return
		}
		if !errors.Is(err, redis.Nil) { // значит что это не ошибка отсуствия - внутренняя ошибка
			fmt.Println("Зашел сюда")
			fmt.Println(err)
			fmt.Println(redis.Nil)
			gCtx.AbortWithError(http.StatusInternalServerError, err)

			return
		}

		myClaims := a.ParseClaims(gCtx)

		ctxWithUserID := gCtx.Request.Context()
		ctxWithUserID = context.WithValue(ctxWithUserID, "userID", myClaims.UserID)
		gCtx.Set("userID", myClaims.UserID)

		userID, exists := gCtx.Get("userID")
		if exists {
			fmt.Println(userID.(uint))
		}

		ctxWithUserRole := gCtx.Request.Context()
		ctxWithUserRole = context.WithValue(ctxWithUserRole, "userRole", myClaims.Role)
		gCtx.Set("userRole", myClaims.Role)

		userRole, exists := gCtx.Get("userRole")
		if exists {
			fmt.Println(userRole.(role.Role))
		}

		fmt.Println("Сюда()")
		fmt.Println(myClaims)
		authorized := false
		fmt.Println(assignedRoles)
		for _, userRole := range assignedRoles {
			if myClaims.Role == userRole {
				authorized = true
				break
			}
		}

		if !authorized {
			gCtx.AbortWithStatus(http.StatusForbidden)
			log.Printf("role is not assigned")
			return
		}

	}

}
func (a *Application) ParseClaims(gCtx *gin.Context) *ds.JWTClaims {

	jwtStr := gCtx.GetHeader("Authorization")
	jwtStr = jwtStr[len(jwtPrefix):] // отрезаем префикс
	fmt.Println(jwtStr)
	token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.config.JWT.Token), nil
	})
	if err != nil {
		gCtx.AbortWithStatus(http.StatusForbidden)
		log.Println(err)

		return nil
	}

	myClaims := token.Claims.(*ds.JWTClaims)
	return myClaims
}

func (a *Application) ParseUserID(gCtx *gin.Context) uint {
	myClaims := a.ParseClaims(gCtx)
	return myClaims.UserID
}
