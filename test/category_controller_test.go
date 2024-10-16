package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"golang-restful-api/app"
	"golang-restful-api/controller"
	"golang-restful-api/helper"
	"golang-restful-api/middleware"
	"golang-restful-api/model/domain"
	"golang-restful-api/repository"
	"golang-restful-api/service"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func setupTestDB() *sql.DB {
	db, err := sql.Open("mysql", "root:123@tcp(localhost:3306)/golang_restful_api_test")
	helper.PanicIfError(err)

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxIdleTime(10 * time.Minute)
	db.SetConnMaxLifetime(60 * time.Minute)

	return db
}

func setupRouter(db *sql.DB) http.Handler {
	validate := validator.New()
	categoryRepository := repository.NewCategoryRepository()
	categoryService := service.NewCategoryService(categoryRepository, db, validate)
	categoryController := controller.NewCategoryController(categoryService)
	router := app.NewRouter(categoryController)

	return middleware.NewAuthMiddleware(router)
}

func truncateCategory(db *sql.DB) {
	db.Exec("truncate category")
}

func TestCreateSuccess(t *testing.T) {
	db := setupTestDB()
	truncateCategory(db)

	router := setupRouter(db)
	requestBody := strings.NewReader(`{"name":"Windows"}`)
	request := httptest.NewRequest(http.MethodPost, "https://localhost:3000/api/categories", requestBody)
	recorder := httptest.NewRecorder()
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-API-KEY", "RAHASIA")

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 200, response.StatusCode)
}

func TestCreateFailed(t *testing.T) {
	db := setupTestDB()
	truncateCategory(db)

	router := setupRouter(db)
	requestBody := strings.NewReader(`{"name":""}`)
	request := httptest.NewRequest(http.MethodPost, "https://localhost:3000/api/categories", requestBody)
	recorder := httptest.NewRecorder()
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-API-KEY", "RAHASIA")

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 400, response.StatusCode)
}

func TestUpdateSuccess(t *testing.T) {
	db := setupTestDB()
	truncateCategory(db)

	tx, _ := db.Begin()
	categoryRepository := repository.NewCategoryRepository()
	category := categoryRepository.Save(context.Background(), tx, domain.Category{
		Name: "Linux",
	})
	tx.Commit()

	router := setupRouter(db)
	requestBody := strings.NewReader(`{"name":"MacOS"}`)
	request := httptest.NewRequest(http.MethodPut, "https://localhost:3000/api/categories/"+strconv.Itoa(category.Id), requestBody)
	recorder := httptest.NewRecorder()
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-API-KEY", "RAHASIA")

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 200, response.StatusCode)

	var responseBody map[string]interface{}
	body, _ := io.ReadAll(response.Body)
	json.Unmarshal(body, &responseBody)
	assert.Equal(t, "OK", responseBody["status"])
	assert.Equal(t, "MacOS", responseBody["data"].(map[string]interface{})["name"])
}

func TestUpdateFailed(t *testing.T) {
	db := setupTestDB()
	truncateCategory(db)

	tx, _ := db.Begin()
	categoryRepository := repository.NewCategoryRepository()
	category := categoryRepository.Save(context.Background(), tx, domain.Category{
		Name: "Linux",
	})
	tx.Commit()

	router := setupRouter(db)
	requestBody := strings.NewReader(`{"name":""}`)
	request := httptest.NewRequest(http.MethodPut, "https://localhost:3000/api/categories/"+strconv.Itoa(category.Id), requestBody)
	recorder := httptest.NewRecorder()
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-API-KEY", "RAHASIA")

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 400, response.StatusCode)

	var responseBody map[string]interface{}
	body, _ := io.ReadAll(response.Body)
	json.Unmarshal(body, &responseBody)
	assert.Equal(t, "BAD REQUEST", responseBody["status"])
	// assert.Equal(t, "MacOS", responseBody["data"].(map[string]interface{})["name"])
}

func TestDeleteSuccess(t *testing.T) {
	db := setupTestDB()
	truncateCategory(db)

	tx, _ := db.Begin()
	categoryRepository := repository.NewCategoryRepository()
	category := categoryRepository.Save(context.Background(), tx, domain.Category{
		Name: "Linux",
	})
	tx.Commit()

	router := setupRouter(db)
	request := httptest.NewRequest(http.MethodDelete, "https://localhost:3000/api/categories/"+strconv.Itoa(category.Id), nil)
	recorder := httptest.NewRecorder()
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-API-KEY", "RAHASIA")

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 200, response.StatusCode)

	var responseBody map[string]interface{}
	body, _ := io.ReadAll(response.Body)
	json.Unmarshal(body, &responseBody)
	assert.Equal(t, "OK", responseBody["status"])
	// assert.Equal(t, "MacOS", responseBody["data"].(map[string]interface{})["name"])
}

func TestDeleteFailed(t *testing.T) {
	db := setupTestDB()
	truncateCategory(db)

	tx, _ := db.Begin()
	categoryRepository := repository.NewCategoryRepository()
	category := categoryRepository.Save(context.Background(), tx, domain.Category{
		Name: "Linux",
	})
	tx.Commit()

	router := setupRouter(db)
	request := httptest.NewRequest(http.MethodDelete, "https://localhost:3000/api/categories/1"+strconv.Itoa(category.Id), nil)
	recorder := httptest.NewRecorder()
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-API-KEY", "RAHASIA")

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 404, response.StatusCode)

	var responseBody map[string]interface{}
	body, _ := io.ReadAll(response.Body)
	json.Unmarshal(body, &responseBody)
	assert.Equal(t, "NOT FOUND", responseBody["status"])
}

func TestGetCategorySuccess(t *testing.T) {
	db := setupTestDB()
	truncateCategory(db)

	tx, _ := db.Begin()
	categoryRepository := repository.NewCategoryRepository()
	category := categoryRepository.Save(context.Background(), tx, domain.Category{
		Name: "Linux",
	})
	tx.Commit()

	router := setupRouter(db)
	request := httptest.NewRequest(http.MethodGet, "https://localhost:3000/api/categories/"+strconv.Itoa(category.Id), nil)
	recorder := httptest.NewRecorder()
	request.Header.Add("X-API-KEY", "RAHASIA")

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 200, response.StatusCode)

	var responseBody map[string]interface{}
	body, _ := io.ReadAll(response.Body)
	json.Unmarshal(body, &responseBody)
	assert.Equal(t, "OK", responseBody["status"])
	assert.Equal(t, "Linux", responseBody["data"].(map[string]interface{})["name"])
}

func TestGetCategoryFailed(t *testing.T) {
	db := setupTestDB()
	truncateCategory(db)

	tx, _ := db.Begin()
	categoryRepository := repository.NewCategoryRepository()
	category := categoryRepository.Save(context.Background(), tx, domain.Category{
		Name: "Linux",
	})
	tx.Commit()

	router := setupRouter(db)
	request := httptest.NewRequest(http.MethodGet, "https://localhost:3000/api/categories/1"+strconv.Itoa(category.Id), nil)
	recorder := httptest.NewRecorder()
	request.Header.Add("X-API-KEY", "RAHASIA")

	router.ServeHTTP(recorder, request)

	response := recorder.Result()
	assert.Equal(t, 404, response.StatusCode)

	var responseBody map[string]interface{}
	body, _ := io.ReadAll(response.Body)
	json.Unmarshal(body, &responseBody)
	assert.Equal(t, "NOT FOUND", responseBody["status"])
	assert.Equal(t, 404, int(responseBody["code"].(float64)))
	// assert.Equal(t, "Linux", responseBody["data"].(map[string]interface{})["name"])
}

func TestGetAllCategorySuccess(t *testing.T) {
	db := setupTestDB()
	truncateCategory(db)

	tx, _ := db.Begin()
	categoryRepository := repository.NewCategoryRepository()
	category := categoryRepository.Save(context.Background(), tx, domain.Category{
		Name: "Linux",
	})
	tx.Commit()

	router := setupRouter(db)
	request := httptest.NewRequest(http.MethodGet, "https://localhost:3000/api/categories", nil)
	recorder := httptest.NewRecorder()
	request.Header.Add("X-API-KEY", "RAHASIA")

	router.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, 200, response.StatusCode)

	var responseBody map[string]interface{}
	body, _ := io.ReadAll(response.Body)
	json.Unmarshal(body, &responseBody)

	assert.Equal(t, "OK", responseBody["status"])
	temp := responseBody["data"].([]interface{})
	categories := temp[0].(map[string]interface{})
	assert.Equal(t, category.Id, int(categories["id"].(float64)))
	assert.Equal(t, category.Name, categories["name"])

}

func TestAuthorized(t *testing.T) {
	db := setupTestDB()
	truncateCategory(db)

	router := setupRouter(db)
	request := httptest.NewRequest(http.MethodGet, "https://localhost:3000/api/categories", nil)
	recorder := httptest.NewRecorder()
	request.Header.Add("X-API-KEY", "")

	router.ServeHTTP(recorder, request)

	response := recorder.Result()

	assert.Equal(t, 401, response.StatusCode)

	var responseBody map[string]interface{}
	body, _ := io.ReadAll(response.Body)
	json.Unmarshal(body, &responseBody)

	assert.Equal(t, "UNAUTHORIZED", responseBody["status"])
}
