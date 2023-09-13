package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/labstack/echo/v4"

	_ "github.com/lib/pq"
)

func genRandNum(min, max int64) int64 {
	// calculate the max we will be using
	bg := big.NewInt(max - min)

	// get big.Int between 0 and bg
	// in this case 0 to 20
	n, err := rand.Int(rand.Reader, bg)
	if err != nil {
		panic(err)
	}

	// add n to min to support the passed in range
	return n.Int64() + min
}

func bookHandler(c echo.Context) error {
	mailID := verifyAuthHeader(c)
	if mailID == "" {
		return c.NoContent(http.StatusUnauthorized)
	}

	slot, err := strconv.Atoi(c.FormValue("slot"))
	if err != nil || slot < 1 || slot > 42 {
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

func sendOTPHandler(c echo.Context) error {
	mailID := c.FormValue("mailID")
	mailID = strings.ToLower(mailID)
	if mailID == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	//check if mailID exists
	row, err := db.Query("SELECT mailID FROM students WHERE mailID = $1", mailID)
	if err != nil {
		panic(err)
	}
	if !row.Next() {
		return c.NoContent(http.StatusUnauthorized)
	}
	//generate and send OTP to mail
	otp := generateOTP(mailID)

	to := []string{mailID + "@" + os.Getenv("MAIL_TO_DOMAIN")}
	msg := []byte(
		"Subject: OTP for Booking Washing Machine Slot\r\n" +
			"\r\n" +
			"Your OTP is " + otp + "\n")
	err = smtp.SendMail(os.Getenv("SMTP_HOST"), auth, os.Getenv("MAIL_ID"), to, msg)
	if err != nil {
		log.Print(err)
	}
	log.Print(otp)

	return c.NoContent(http.StatusOK)
}

func generateOTP(mailID string) string {
	o := genRandNum(1000, 9999)
	//update in database
	_, err := db.Exec("UPDATE students SET otp = $1 WHERE mailID = $2", fmt.Sprint(o), mailID)
	if err != nil {
		panic(err)
	}
	return fmt.Sprint(o)

}

func loginHandler(c echo.Context) error {
	mailID := c.FormValue("mailID")
	otp := c.FormValue("otp")
	if mailID == "" || otp == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	var otpInDB string
	//check otp in database
	db.QueryRow("SELECT otp FROM students WHERE mailID = $1", mailID).Scan(&otpInDB)
	if otp != otpInDB {
		return c.NoContent(http.StatusUnauthorized)
	}
	jwt := generateJWT(mailID)
	return c.String(http.StatusOK, jwt)
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
