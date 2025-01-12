package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"weight-challenge/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Помощна функция за създаване на тестов контекст
func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// TestGetUserSettings тества функцията за взимане на потребителски настройки
func TestGetUserSettings(t *testing.T) {
	// Подготовка на тестовата среда
	gin.SetMode(gin.TestMode)
	c, w := setupTestContext()

	// Симулираме автентикиран потребител
	c.Set("userID", 1)

	// Изпълняваме функцията
	getUserSettings(c)

	// Проверяваме резултата
	assert.Equal(t, http.StatusOK, w.Code)

	// Декодираме отговора
	var response models.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	
	// Проверяваме дали декодирането е успешно
	assert.NoError(t, err)
	
	// Проверяваме дали отговорът съдържа правилните данни
	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.Username)
}

// TestUpdateUserSettings тества функцията за обновяване на потребителски настройки
func TestUpdateUserSettings(t *testing.T) {
	// Подготовка на тестовата среда
	gin.SetMode(gin.TestMode)
	c, w := setupTestContext()

	// Симулираме автентикиран потребител
	c.Set("userID", 1)

	// Подготвяме тестови данни
	testSettings := models.User{
		FirstName: "Test",
		LastName:  "User",
		Age:       25,
		Height:    180,
		Gender:    "male",
		Email:     "test@example.com",
		Target:    75.5,
	}

	// Подготвяме request body
	jsonData, err := json.Marshal(testSettings)
	assert.NoError(t, err)

	// Симулираме HTTP заявка
	c.Request = httptest.NewRequest("PUT", "/settings", bytes.NewBuffer(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")

	// Изпълняваме функцията
	updateUserSettings(c)

	// Проверяваме резултата
	assert.Equal(t, http.StatusOK, w.Code)

	// Декодираме отговора
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	
	// Проверяваме дали декодирането е успешно
	assert.NoError(t, err)
	
	// Проверяваме съобщението за успех
	assert.Equal(t, "Settings updated successfully", response["message"])
}

// TestUpdateVisibility тества функцията за обновяване на видимостта на профила
func TestUpdateVisibility(t *testing.T) {
	// Подготовка на тестовата среда
	gin.SetMode(gin.TestMode)
	c, w := setupTestContext()

	// Симулираме автентикиран потребител
	c.Set("userID", 1)

	// Подготвяме тестови данни
	testVisibility := struct {
		IsVisible bool `json:"isVisible"`
	}{
		IsVisible: true,
	}

	// Подготвяме request body
	jsonData, err := json.Marshal(testVisibility)
	assert.NoError(t, err)

	// Симулираме HTTP заявка
	c.Request = httptest.NewRequest("PUT", "/settings/visibility", bytes.NewBuffer(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")

	// Изпълняваме функцията
	updateVisibility(c)

	// Проверяваме резултата
	assert.Equal(t, http.StatusOK, w.Code)

	// Декодираме отговора
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	
	// Проверяваме дали декодирането е успешно
	assert.NoError(t, err)
	
	// Проверяваме съобщението за успех
	assert.Equal(t, "Visibility updated", response["message"])
} 