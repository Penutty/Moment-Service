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

	e.GET("/moment/shared/user/:user/", userShared)
	e.GET("/moment/found/user/:user/", userFound)
	e.GET("/moment/left/user/:user/", userLeft)

	e.GET("/moment/shared/location/:location", locationShared)
	e.GET("/moment/public/location/:location", locationPublic)

	e.GET("/moment/hidden/location/:location", locationHidden)
	e.GET("/moment/lost/location/:location", locationLost)

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
