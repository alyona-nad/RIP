package api

import (
"github.com/joho/godotenv"
"awesomeProject/internal/app/dsn"
"awesomeProject/internal/app/repository"
"time"

"awesomeProject/internal/app/redis"
"github.com/golang-jwt/jwt"
"github.com/kelseyhightower/envconfig"
)

type Application struct {
	config     *Config
	repository *repository.Repository
	redis      *redis.Client
}
type Config struct {
	JWT struct {
		Token         string
		SigningMethod jwt.SigningMethod
		ExpiresIn     time.Duration
	}
}
func New() (Application, error) {
_ = godotenv.Load()
config := &Config{}
err := envconfig.Process("", config)
if err != nil {
return Application{}, err
}
repo, err := repository.New(dsn.FromEnv())
if err != nil {
return Application{}, err
}
redisClient, err := redis.New()
if err != nil {
	return Application{}, err
}

return Application{config: config, repository: repo, redis: redisClient}, nil
}

