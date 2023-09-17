package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func main() {

	initDB()
	e := echo.New()
	e.POST("/book", bookHandler)

	e.POST("/changepassword", changePasswordHandler)
	e.POST("/login", loginHandler)
	e.POST("/status", statusHandler)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.Use(middleware.Logger())

	e.Logger.Fatal(e.Start(":8080"))

	//New route path book a slot with data input in form with name, requested slot s.no and roll number

}
