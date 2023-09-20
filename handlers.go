package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

// set timezone,
var loc, _ = time.LoadLocation("Asia/Kolkata")

func loginHandler(c echo.Context) error {
	mailID := c.FormValue("mailID")
	password := c.FormValue("password")
	if !isSafe(mailID, password) {
		return c.NoContent(http.StatusBadRequest)
	}
	passwordBytes := []byte(password)
	mailID = strings.ToLower(mailID)

	var passwordInDB string

	db.QueryRow("SELECT password FROM students WHERE mailID = $1", mailID).Scan(&passwordInDB)
	err := bcrypt.CompareHashAndPassword([]byte(passwordInDB), passwordBytes)
	if err != nil {
		return c.NoContent(http.StatusUnauthorized)
	}
	jwt := generateJWT(mailID)
	return c.String(http.StatusOK, jwt)
}

func bookHandler(c echo.Context) error {
	mailID := verifyAuthHeader(c)
	if mailID == "" {
		return c.NoContent(http.StatusUnauthorized)
	}
	if !isSafe(c.FormValue("slot")) {
		return c.NoContent(http.StatusBadRequest)
	}

	slot, err := strconv.Atoi(c.FormValue("slot"))
	if err != nil || slot < 1 || slot > 46 {
		return c.NoContent(http.StatusBadRequest)
	}

	//get the data and check if slot is free by checking if the row is empty
	var bookedBy sql.NullString
	err = db.QueryRow("SELECT mailID FROM slots WHERE slotno = $1", slot).Scan(&bookedBy)
	if err != nil {
		panic(err)
	}
	if bookedBy.Valid {
		return c.NoContent(http.StatusAlreadyReported)
	}

	//check if it has been a week since student's 2nd last booked slot
	var lastBooked sql.NullInt64
	err = db.QueryRow("SELECT date1 FROM students WHERE mailID = $1", mailID).Scan(&lastBooked)
	if err != nil {
		panic(err)
	}
	var lastBookedInt int64 = 0
	if lastBooked.Valid {
		lastBookedInt = lastBooked.Int64
	}
	if time.Since(time.Unix(lastBookedInt, 0)).Hours() < 168 {
		return c.NoContent(http.StatusForbidden)
	}

	//book slot
	_, err = db.Exec("UPDATE slots SET mailID = $1 WHERE slotno = $2", mailID, slot)

	if err != nil {
		panic(err)
	}
	//Move date2 to date1 and set date 2 todays
	_, err = db.Exec("UPDATE students SET date1 = date2, date2 = $1::integer WHERE mailID = $2::text", time.Now().Unix(), mailID)
	if err != nil {
		panic(err)
	}

	return c.String(http.StatusOK, "Booked")
}

func cancelHandler(c echo.Context) error {
	mailID := verifyAuthHeader(c)
	if mailID == "" {
		return c.NoContent(http.StatusUnauthorized)
	}
	if !isSafe(c.FormValue("slot")) {
		return c.NoContent(http.StatusBadRequest)
	}

	slot, err := strconv.Atoi(c.FormValue("slot"))
	if err != nil || slot < 1 || slot > 46 {
		return c.NoContent(http.StatusBadRequest)
	}

	//get the data and check if slot is booked by mailID
	var bookedBy sql.NullString
	err = db.QueryRow("SELECT mailID FROM slots WHERE slotno = $1", slot).Scan(&bookedBy)
	if err != nil {
		panic(err)
	}
	if !bookedBy.Valid || bookedBy.String != mailID {
		return c.NoContent(http.StatusForbidden)
	}

	//check if the cancellation is on the same day
	if time.Now().In(loc).Weekday().String() == getDayFromSlotNo(slot) {
		if getSlotStartHour(slot)-time.Now().In(loc).Hour() < 4 {
			return c.String(http.StatusForbidden, "Cannot cancel now")
		}

	}

	//cancel slot
	_, err = db.Exec("UPDATE slots SET mailID = NULL WHERE slotno = $1", slot)

	if err != nil {
		panic(err)
	}

	//if date1 is null set date 2 to null
	var date1 sql.NullInt64
	err = db.QueryRow("SELECT date1 FROM students WHERE mailID = $1", mailID).Scan(&date1)
	if err != nil {
		panic(err)
	}
	if !date1.Valid {
		_, err = db.Exec("UPDATE students SET date2 = NULL WHERE mailID = $1::text", mailID)
		if err != nil {
			panic(err)
		}
	} else {
		_, err = db.Exec("UPDATE students SET date1 = NULL WHERE mailID = $1::text", mailID)
		if err != nil {
			panic(err)
		}
	}

	return c.String(http.StatusOK, "Slot cancelled")
}
func changePasswordHandler(c echo.Context) error {
	mailID := c.FormValue("mailID")
	password := c.FormValue("password")
	newPassword := c.FormValue("newPassword")
	if !isSafe(mailID, password, newPassword) {
		return c.NoContent(http.StatusBadRequest)
	}
	passwordBytes := []byte(password)
	newPasswordBytes := []byte(newPassword)
	mailID = strings.ToLower(mailID)

	var passwordInDB string

	db.QueryRow("SELECT password FROM students WHERE mailID = $1", mailID).Scan(&passwordInDB)
	err := bcrypt.CompareHashAndPassword([]byte(passwordInDB), passwordBytes)
	if err != nil {
		return c.NoContent(http.StatusUnauthorized)
	}
	newPasswordBytes, _ = bcrypt.GenerateFromPassword(newPasswordBytes, bcrypt.DefaultCost)
	_, err = db.Exec("UPDATE students SET password = $1 WHERE mailID = $2", string(newPasswordBytes[:]), mailID)
	if err != nil {
		log.Println(err)
	}
	return c.NoContent(http.StatusOK)
}

// show the entire slots table in json format
func statusHandler(c echo.Context) error {
	rows, err := db.Query("SELECT * FROM slots WHERE mailID IS NOT NULL")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	type Entry struct {
		Slotno int    `json:"slotno"`
		MailID string `json:"mailID"`
	}

	var entries []Entry

	for rows.Next() {
		var slotno int
		var mailID sql.NullString

		rows.Scan(&slotno, &mailID)

		entries = append(entries, Entry{slotno, mailID.String})
	}
	if len(entries) == 0 {
		return c.NoContent(http.StatusNoContent)
	}
	entryBytes, _ := json.MarshalIndent(&entries, "", "  ")

	return c.JSONBlob(http.StatusOK, entryBytes)

}

func isSafe(args ...string) bool {
	for _, arg := range args {
		if len(arg) > 25 || len(arg) < 1 {
			return false
		}
		for _, ch := range arg {
			if ch < ' ' || ch > '~' {
				return false
			}
		}
	}
	return true
}
