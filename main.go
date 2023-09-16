package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}
}

func main() {

	initDB()
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.POST("/book", bookHandler)

	e.POST("/changepassword", changePasswordHandler)
	e.POST("/login", loginHandler)
	e.POST("/status", statusHandler)

	e.Logger.Fatal(e.Start(":8080"))

	//New route path book a slot with data input in form with name, requested slot s.no and roll number

}
