package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	_ "github.com/lib/pq"
)

func bookHandler(c echo.Context) error {
	rollno := verifyAuthHeader(c)
	if rollno == "" {
		return c.NoContent(http.StatusUnauthorized)
	}

	slot, err := strconv.Atoi(c.FormValue("slot"))
	if err != nil || slot < 1 || slot > 42 {
		return c.NoContent(http.StatusBadRequest)
	}

	//get the data and check if slot is free by checking if the row is empty
	rows, err := db.Query("SELECT rollno FROM slots WHERE slot = $1", slot)
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		return c.NoContent(http.StatusAlreadyReported)
	}

	//check if it has been a week since student's 2nd last booked slot
	var lastBooked int
	err = db.QueryRow("SELECT date1 FROM students WHERE rollno = $1", rollno).Scan(&lastBooked)
	if err != nil {
		panic(err)
	}
	if time.Now().Sub(time.Unix(int64(lastBooked), 0)).Hours() > 168 {
		return c.NoContent(http.StatusForbidden)
	}

	//book slot
	_, err = db.Exec("UPDATE slots SET rollno = $1 WHERE slot = $2", rollno, slot)

	if err != nil {
		panic(err)
	}
	//Move date2 to date1 and add todays date to date2
	_, err = db.Exec("UPDATE students SET date1 = date2, date2 = date1 + 1 WHERE rollno = $1", rollno)

	return c.String(http.StatusOK, "Booked")
}
