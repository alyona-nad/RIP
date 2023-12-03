package ds

import (
	"github.com/golang-jwt/jwt"
	"awesomeProject/internal/app/role"
)

type JWTClaims struct {
	jwt.StandardClaims           // все что точно необходимо по RFC
	UserID           uint
	Role               role.Role
}