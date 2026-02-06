package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"social-network/internal/config"
	"social-network/internal/database"
	"social-network/internal/handler"
	"social-network/internal/repository"
	"social-network/internal/service"
)

// testApp — тестовое приложение
type testApp struct {
	handler http.Handler
}

// tokenPair — пара токенов из ответа
type tokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// authResponse — ответ на регистрацию/логин
type authResponse struct {
	User   map[string]any `json:"user"`
	Tokens tokenPair      `json:"tokens"`
}

// setupTestApp создаёт тестовое приложение с чистой БД
func setupTestApp(t *testing.T) *testApp {
	t.Helper()

	// Используем тестовую БД
	cfg := &config.Config{
		DBHost:     getTestEnv("TEST_DB_HOST", "localhost"),
		DBPort:     getTestEnv("TEST_DB_PORT", "5432"),
		DBUser:     getTestEnv("TEST_DB_USER", "postgres"),
		DBPassword: getTestEnv("TEST_DB_PASSWORD", "postgres"),
		DBName:     getTestEnv("TEST_DB_NAME", "social_network_test"),
		DBSSLMode:  "disable",
		JWTSecret:  "test-secret",
		ServerPort: "0",
	}

	db, err := database.Connect(cfg.DSN())
	if err != nil {
		t.Fatalf("Не удалось подключиться к тестовой БД: %v", err)
	}

	// Чистим все таблицы перед тестами
	tables := []string{"refresh_tokens", "likes", "follows", "comments", "posts", "users", "schema_migrations"}
	for _, table := range tables {
		db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE")
	}

	// Применяем миграции
	if err := database.RunMigrations(db, "../migrations"); err != nil {
		t.Fatalf("Ошибка миграций: %v", err)
	}

	// Создаём директорию для аватарок
	os.MkdirAll("../web/uploads", 0755)

	// Инициализация слоёв
	userRepo := repository.NewUserRepo(db)
	postRepo := repository.NewPostRepo(db)
	commentRepo := repository.NewCommentRepo(db)
	followRepo := repository.NewFollowRepo(db)
	likeRepo := repository.NewLikeRepo(db)
	tokenRepo := repository.NewTokenRepo(db)

	authService := service.NewAuthService(userRepo, tokenRepo, cfg.JWTSecret)
	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo)
	commentService := service.NewCommentService(commentRepo)
	followService := service.NewFollowService(followRepo)
	likeService := service.NewLikeService(likeRepo)

	h := handler.NewHandler(authService, userService, postService, commentService, followService, likeService)

	t.Cleanup(func() {
		db.Close()
	})

	return &testApp{handler: h.Routes()}
}

// registerUser регистрирует пользователя и возвращает authResponse
func (app *testApp) registerUser(t *testing.T, username, email, password string) authResponse {
	t.Helper()
	body := fmt.Sprintf(`{"username":"%s","email":"%s","password":"%s"}`, username, email, password)
	req := httptest.NewRequest("POST", "/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Ожидали 201 при регистрации, получили %d: %s", w.Code, w.Body.String())
	}

	var resp authResponse
	json.NewDecoder(w.Body).Decode(&resp)
	return resp
}

// loginUser логинит пользователя и возвращает authResponse
func (app *testApp) loginUser(t *testing.T, email, password string) authResponse {
	t.Helper()
	body := fmt.Sprintf(`{"email":"%s","password":"%s"}`, email, password)
	req := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Ожидали 200 при логине, получили %d: %s", w.Code, w.Body.String())
	}

	var resp authResponse
	json.NewDecoder(w.Body).Decode(&resp)
	return resp
}

// authRequest выполняет HTTP-запрос с токеном авторизации
func (app *testApp) authRequest(method, path, token string, body any) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	app.handler.ServeHTTP(w, req)
	return w
}

// request выполняет публичный HTTP-запрос
func (app *testApp) request(method, path string, body any) *httptest.ResponseRecorder {
	return app.authRequest(method, path, "", body)
}

func getTestEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
