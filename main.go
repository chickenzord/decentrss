package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/chickenzord/decentrss/internal/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	bindPort = "8080"
)

func init() {
	if port := os.Getenv("BIND_PORT"); port != "" {
		bindPort = port
	}
}

func main() {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())
	e.GET("/ping", func(c echo.Context) error {
		return c.String(200, "pong")
	})
	e.GET("/feed", func(c echo.Context) error {
		url := c.QueryParam("url")
		if url == "" {
			return c.String(400, "url is required")
		}

		feed, err := database.GetFeed(url)
		if err != nil {
			if errors.Is(err, &database.ErrFeedNotFound{}) {
				return c.String(404, err.Error())
			}

			return c.String(500, err.Error())
		}

		c.Response().Header().Set("Content-Type", "application/json")
		if err := feed.WriteJSON(c.Response().Writer); err != nil {
			return c.String(500, err.Error())
		}

		return nil
	})
	e.POST("/feed", func(c echo.Context) error {
		url, itemCount, err := database.ParseAndSaveFeed(c.Request().Body)
		if err != nil {
			return c.String(500, err.Error())
		}

		return c.JSON(http.StatusAccepted, echo.Map{
			"success": true,
			"data": echo.Map{
				"item_count": itemCount,
				"url":        url,
			},
		})
	})

	if err := e.Start(fmt.Sprintf(":%s", bindPort)); err != nil {
		panic(err)
	}
}
