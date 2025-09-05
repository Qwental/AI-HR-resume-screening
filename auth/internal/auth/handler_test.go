package auth

import (
	"ai-hr-service/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock для Service
type MockService struct {
	mock.Mock
}

func (m *MockService) Register(req RegisterRequest) (*LoginResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LoginResponse), args.Error(1)
}

func (m *MockService) Login(req LoginRequest) (*LoginResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LoginResponse), args.Error(1)
}

func (m *MockService) RefreshTokens(req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RefreshTokenResponse), args.Error(1)
}

func (m *MockService) Logout(req LogoutRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockService) LogoutAll(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockService) GetProfile(userID uint) (*models.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestHandler_Register_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)

	registerReq := RegisterRequest{
		Username: "testuser",
		Surname:  "TestSurname",
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedResponse := &LoginResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    1800,
		User: models.User{
			Username: registerReq.Username,
			Email:    registerReq.Email,
		},
	}

	mockService.On("Register", registerReq).Return(expectedResponse, nil)

	jsonData, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, expectedResponse.AccessToken, data["access_token"])
	assert.Equal(t, expectedResponse.RefreshToken, data["refresh_token"])

	mockService.AssertExpectations(t)
}

func TestHandler_Register_UserExists(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)

	registerReq := RegisterRequest{
		Username: "existinguser",
		Surname:  "TestSurname",
		Email:    "existing@example.com",
		Password: "password123",
	}

	mockService.On("Register", registerReq).Return(nil, errors.New("user with this email already exists"))

	jsonData, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Register(c)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "user with this email already exists", response["error"])

	mockService.AssertExpectations(t)
}

func TestHandler_Login_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)

	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedResponse := &LoginResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    1800,
		User: models.User{
			Email: loginReq.Email,
		},
	}

	mockService.On("Login", loginReq).Return(expectedResponse, nil)

	jsonData, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, expectedResponse.AccessToken, data["access_token"])

	mockService.AssertExpectations(t)
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)

	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	mockService.On("Login", loginReq).Return(nil, errors.New("invalid credentials"))

	jsonData, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid credentials", response["error"])

	mockService.AssertExpectations(t)
}

func TestHandler_LogoutAll_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)

	userID := uint(1)

	mockService.On("LogoutAll", userID).Return(nil)

	req := httptest.NewRequest("POST", "/logout-all", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.LogoutAll(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Successfully logged out from all devices", response["message"])

	mockService.AssertExpectations(t)
}

func TestHandler_LogoutAll_NoUserID(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest("POST", "/logout-all", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Не устанавливаем user_id

	handler.LogoutAll(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "user ID not found", response["error"])
}

func TestHandler_GetProfile_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)

	userID := uint(1)
	expectedUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockService.On("GetProfile", userID).Return(expectedUser, nil)

	req := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data := response["data"].(map[string]interface{})
	user := data["user"].(map[string]interface{})
	assert.Equal(t, expectedUser.Username, user["username"])
	assert.Equal(t, expectedUser.Email, user["email"])

	mockService.AssertExpectations(t)
}

func TestHandler_GetProfile_UserNotFound(t *testing.T) {
	mockService := new(MockService)
	handler := NewHandler(mockService)

	userID := uint(999)

	mockService.On("GetProfile", userID).Return(nil, errors.New("user not found"))

	req := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.GetProfile(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "user not found", response["error"])

	mockService.AssertExpectations(t)
}
