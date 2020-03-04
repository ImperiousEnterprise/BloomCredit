package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockDb struct{}

//This method is just a placeholder so that I can use the db interface
func (m mockDb) AddCustomerAndTags([][]interface{}) error {
	return nil
}
func (m mockDb) GetConsumerId(first string, last string) (string, error) {
	if first == "" || last == "" {
		return "", errors.New("Empty Bro")
	}
	return "12345676", nil
}
func (m mockDb) GetCreditTags(tag string) (map[string]string, error) {
	mp := make(map[string]map[string]string)
	mp["6d8a3aa2-7b69-4253-9ffb-3b297c703396"] = map[string]string{"X001": "1", "X002": "2"}
	mp["6d8a3aa2-7b69-4253-9ffb-3b297c703397"] = map[string]string{"X001": "1", "X002": "2"}
	res := mp[tag]
	if res == nil {
		return nil, sql.ErrNoRows
	}

	return res, nil
}
func (m mockDb) GetStats(stat string) (map[string]interface{}, error) {
	mp := make(map[string]map[string]interface{})
	mp["X0001"] = map[string]interface{}{"mean": "1", "median": "2", "standard_deviation": "3"}
	mp["X0002"] = map[string]interface{}{"mean": "4", "median": "5", "standard_deviation": "6"}
	res := mp[stat]
	if res == nil {
		return nil, sql.ErrNoRows
	}

	return res, nil
}

func TestGetCustomerId_Good(t *testing.T) {
	e := echo.New()
	db := mockDb{}
	ev := Env{db}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	req := httptest.NewRequest(http.MethodGet, "/getId?first_name=test&last_name=me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, ev.FetchCustomerId(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		removedNewSpace := strings.TrimSuffix(rec.Body.String(), "\n")
		assert.Equal(t, "\"12345676\"", removedNewSpace)
	}
}
func TestGetCustomerTags_Good(t *testing.T) {
	e := echo.New()
	db := mockDb{}
	ev := Env{db}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	req := httptest.NewRequest(http.MethodGet, "/customer?id=6d8a3aa2-7b69-4253-9ffb-3b297c703396", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	j, _ := json.Marshal(map[string]string{"X001": "1", "X002": "2"})

	if assert.NoError(t, ev.GetCustomerTags(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		removedNewSpace := strings.TrimSuffix(rec.Body.String(), "\n")
		assert.Equal(t, string(j), removedNewSpace)
	}
}

func TestGetCustomerId_Bad(t *testing.T) {
	e := echo.New()
	db := mockDb{}
	ev := Env{db}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	req := httptest.NewRequest(http.MethodGet, "/getId?first_name=test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, ev.FetchCustomerId(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		removedNewSpace := strings.TrimSuffix(rec.Body.String(), "\n")
		assert.Equal(t, "\"Need First and Last Name\"", removedNewSpace)
	}
}
func TestGetCustomerTags_Bad(t *testing.T) {
	e := echo.New()
	db := mockDb{}
	ev := Env{db}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	req := httptest.NewRequest(http.MethodGet, "/customer?id=6d8a3aa2-7b69-4253-9ffb-3b", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, ev.GetCustomerTags(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		removedNewSpace := strings.TrimSuffix(rec.Body.String(), "\n")
		assert.Equal(t, "\"Invalid Customer Id\"", removedNewSpace)
	}
}
func TestGetCustomerTags_Bad_NonExistantCustomer(t *testing.T) {
	e := echo.New()
	db := mockDb{}
	ev := Env{db}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	req := httptest.NewRequest(http.MethodGet, "/customer?id=6d8a3aa2-7b69-4253-9ffb-3b297c703398", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, ev.GetCustomerTags(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		removedNewSpace := strings.TrimSuffix(rec.Body.String(), "\n")
		assert.Equal(t, "\"Customer Doesn't Exist\"", removedNewSpace)
	}
}

func TestGetStats_Good(t *testing.T) {
	e := echo.New()
	db := mockDb{}
	ev := Env{db}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	req := httptest.NewRequest(http.MethodGet, "/stats?tag=x0001", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	j, _ := json.Marshal(map[string]interface{}{"mean": "1", "median": "2", "standard_deviation": "3"})
	if assert.NoError(t, ev.GetTagsStats(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		removedNewSpace := strings.TrimSuffix(rec.Body.String(), "\n")
		assert.Equal(t, string(j), removedNewSpace)
	}
}

func TestGetStats_Bad(t *testing.T) {
	e := echo.New()
	db := mockDb{}
	ev := Env{db}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	req := httptest.NewRequest(http.MethodGet, "/stats?tag=x200000", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, ev.GetTagsStats(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		removedNewSpace := strings.TrimSuffix(rec.Body.String(), "\n")
		assert.Equal(t, "\"tag out of bounds\"", removedNewSpace)
	}

	req = httptest.NewRequest(http.MethodGet, "/stats?tag=UU3", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, ev.GetTagsStats(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		removedNewSpace := strings.TrimSuffix(rec.Body.String(), "\n")
		assert.Equal(t, "\"invalid number\"", removedNewSpace)
	}

	req = httptest.NewRequest(http.MethodGet, "/stats?tag=U0003", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	if assert.NoError(t, ev.GetTagsStats(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		removedNewSpace := strings.TrimSuffix(rec.Body.String(), "\n")
		assert.Equal(t, "\"wrong credit tag format\"", removedNewSpace)
	}
}
