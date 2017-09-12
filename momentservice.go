package main

import (
	"fmt"
	"github.com/labstack/echo"
	"net/http"
)

func main() {
	e := echo.New()

	e.Pre(validateJwt)
	e.GET("/home", getHome)

	e.Logger.Fatal(e.Start(":8081"))
}

func getHome(c echo.Context) error {
	return c.String(http.StatusOK, "Welcome to the Home page.")
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
