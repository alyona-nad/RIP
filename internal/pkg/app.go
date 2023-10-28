package app

import (
	"fmt"
	"net/http"
)

type Application struct {
	Config       map[string]string
	Router       *http.ServeMux
	DatabaseConn string
}

func (app *Application) Run() {
	fmt.Println("Application is running...")

}

func New() *Application {

	application := &Application{
		Config: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Router:       http.NewServeMux(),
		DatabaseConn: "your_database_connection_string",
	}

	return application
}
