package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/lucasarieta/gcs-scraper/domain"
)

func main() {
	app := echo.New()
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())

	routes := app.Group("")
	routes.GET("/", func(c echo.Context) error {
		return c.String(200, "Hello, World!")
	})

	scraperDomain := domain.ScraperDomain{}
	scraperDomain.SetupRoutes(routes)

	app.Logger.Fatal(app.Start(":8080"))
}
