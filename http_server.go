package main

import (
	"fmt"
	"net/http"
	"time"

	"./config"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type (
	Payload struct {
		Key   string      `json:"key" form:"key" query:"key"`
		Value interface{} `json:"value" form:"value" query:"value"`
		TTL   int64       `json:"ttl" form:"ttl" query:"ttl"`
	}

	Response struct {
		Status       string      `json:"status"`
		Value        interface{} `json:"value,omitempty"`
		ExpiredAt    string      `json:"expired_at,omitempty"`
		ErrorMessage string      `json:"error_message,omitempty"`
	}
)

func startHTTPServer() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == *config.AuthToken, nil
	}))

	// Routes
	e.GET("/", healthCheck)
	e.GET("/:key", getValue)
	e.POST("/", setValue)
	e.DELETE("/:key", deleteValue)
	e.GET("/keys", getAllKeys)

	// Start server
	address := fmt.Sprintf("%s:%s", *config.ServerIP, *config.ServerPort)
	e.Logger.Fatal(e.Start(address))
}

func healthCheck(c echo.Context) error {
	return successResponse(c, "")
}

func getAllKeys(c echo.Context) error {
	keys, err := cacheManager.GetKeys()
	if err != nil {
		return errorResponse(c, "Error occured while collecting cache keys.")
	}

	return c.JSON(http.StatusOK, Response{
		Status: "ok",
		Value:  keys,
	})
}

func getValue(c echo.Context) error {
	key := c.Param("key")
	value, expiredAt, found, err := cacheManager.Get(key)
	if err != nil {
		errorMessage := fmt.Sprintf("Error occured while Get value '%s' from cache.", key)
		return errorResponse(c, errorMessage)
	}
	fmt.Printf("Search for key = %s, %b", value, found)
	if !found {
		errorMessage := fmt.Sprintf("Key '%s' not found in cache.", key)
		return errorResponse(c, errorMessage)
	}

	response := Response{Status: "ok"}
	if value != "" {
		response.Value = value
		if expiredAt != 0 {
			response.ExpiredAt = time.Unix(expiredAt, 0).Format("2006-01-02 15:04:05")
		}

	}
	fmt.Printf("Response %+v", response)
	return c.JSON(http.StatusOK, response)
}

func setValue(c echo.Context) error {
	payload := new(Payload)
	if err := c.Bind(payload); err != nil {
		return errorResponse(c, "Unprocessable request payload. Error: "+err.Error())
	}

	error := cacheManager.Set(payload.Key, payload.Value, payload.TTL)
	if error != nil {
		// fmt.Printf("Error occured while adding new key/value pair: %s - %s", key, value)
		errorMessage := fmt.Sprintf("Error occured while adding new key/value pair: %s - %s", payload.Key, payload.Value)
		return errorResponse(c, errorMessage)
	}

	return successResponse(c, "")
}

func deleteValue(c echo.Context) error {
	key := c.Param("key")
	error := cacheManager.Delete(key)
	if error != nil {
		fmt.Printf("Error occured while deleting key: %s", key)
	}
	return successResponse(c, "")
}

func successResponse(c echo.Context, value string) error {
	response := Response{Status: "ok"}
	if value != "" {
		response.Value = value
	}
	return c.JSON(http.StatusOK, response)
}

func errorResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, Response{
		Status:       "error",
		ErrorMessage: message,
	})
}
