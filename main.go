package main

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
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

	e.Logger.Fatal(e.Start(":1323"))

	//New route path book a slot with data input in form with name, requested slot s.no and roll number
	e.POST("/book", bookHandler)

}
