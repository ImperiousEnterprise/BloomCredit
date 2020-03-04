package main

import (
	"BloomCredit/db"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Env struct {
	DB db.Store
}

func main() {
	// Echo instance
	e := echo.New()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "pass", "postgres")

	db, err := db.NewDB(psqlInfo)
	if err != nil {
		log.Panic(err)
	}

	ev := Env{db}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/getId", ev.FetchCustomerId)
	e.GET("/customer", ev.GetCustomerTags)
	e.GET("/stats", ev.GetTagsStats)

	// Start server
	e.Logger.Fatal(e.Start(":7000"))
}

func (e Env) FetchCustomerId(c echo.Context) error {
	first := c.QueryParam("first_name")
	last := c.QueryParam("last_name")

	values := c.QueryParams()
	c.Logger().Print(values)

	if first == "" || last == "" {
		return WriteErrorJson(c, http.StatusBadRequest, errors.New("Need First and Last Name"))
	}

	id, err := e.DB.GetConsumerId(strings.ToLower(first), strings.ToLower(last))
	if err == sql.ErrNoRows {
		return WriteErrorJson(c, http.StatusBadRequest, errors.New("Customer Doesn't Exist"))
	} else if err != nil {
		return WriteErrorJson(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, id)
}

func (e Env) GetCustomerTags(c echo.Context) error {
	id := c.QueryParam("id")
	if !IsValidUUID(id) {
		return WriteErrorJson(c, http.StatusBadRequest, errors.New("Invalid Customer Id"))
	}

	resp, err := e.DB.GetCreditTags(id)
	if err == sql.ErrNoRows {
		return WriteErrorJson(c, http.StatusBadRequest, errors.New("Customer Doesn't Exist"))
	} else if err != nil {
		return WriteErrorJson(c, http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (e Env) GetTagsStats(c echo.Context) error {
	var err error
	tag := strings.Title(c.QueryParam("tag"))
	num, err := strconv.Atoi(tag[1:])
	if err != nil {
		return WriteErrorJson(c, http.StatusBadRequest, errors.New("invalid number"))
	} else if num < 1 || num > 200 {
		return WriteErrorJson(c, http.StatusBadRequest, errors.New("tag out of bounds"))
	} else if tag[0:1] != "X" {
		return WriteErrorJson(c, http.StatusBadRequest, errors.New("wrong credit tag format"))
	}

	stat, err := e.DB.GetStats(tag)

	if err != nil {
		return WriteErrorJson(c, http.StatusBadRequest, err)
	}
	return c.JSON(http.StatusOK, stat)
}

func WriteErrorJson(e echo.Context, errorNum int, err error) error {
	e.Logger().Error(err)
	return e.JSON(errorNum, err.Error())
}
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
