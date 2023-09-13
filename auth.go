package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var key []byte

func verifyAuthHeader(c echo.Context) string {
	token, err := jwt.Parse(c.Request().Header.Get("Authorization"), func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return key, nil
	})
	if err != nil {
		fmt.Println(err)
		return ""
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["sub"].(string)
	} else {
		return ""
	}

}

func generateJWT(rollno string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": rollno,
		"nbf": time.Now().Unix(), //not before current time
		"exp": time.Now().Add(time.Hour * 36).Unix(),
	})

	tokenString, err := token.SignedString(key)
	if err != nil {
		panic(err)
	}

	return tokenString
}
