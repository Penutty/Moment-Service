package main

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
	"os"
)

func main() {
	e := echo.New()
	f, err := os.OpenFile("/home/tjp/go/log/momentservice.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	e.Logger.SetOutput(f)

	e.Pre(validateJwt)

	e.GET("/moment/:moment/user/:user/", userMoment)
	e.GET("/moment/:moment/location/:location", locationMoment)
	e.POST("/moment/:moment/:action", setMoment)

	e.Logger.Fatal(e.Start(":8081"))
}

func userMoment(c echo.Context) error {

}

func locationMoment(c echo.Context) error {

}

func setMoment(c echo.Context) error {

}

func validateJwt(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Printf("\nc:\n%v\n", c)
		fmt.Printf("c.Request().Header\n = %v\n", c.Request().Header)
		if err := c.Redirect(http.StatusUseProxy, "http://localhost:8080/auth"); err != nil {
			c.Logger().Print(err)
		}
		return next(c)
	}
}
