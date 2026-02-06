package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// ==================== АВТОРИЗАЦИЯ ====================

func TestRegister(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	if resp.Tokens.AccessToken == "" {
		t.Error("access_token пустой")
	}
	if resp.Tokens.RefreshToken == "" {
		t.Error("refresh_token пустой")
	}
	if resp.User["username"] != "testuser" {
		t.Errorf("username = %v, ожидали testuser", resp.User["username"])
	}
}

func TestRegisterDuplicate(t *testing.T) {
	app := setupTestApp(t)
	app.registerUser(t, "testuser", "test@test.com", "password123")

	// Пытаемся зарегистрировать с тем же email
	w := app.request("POST", "/v1/auth/register", map[string]string{
		"username": "other", "email": "test@test.com", "password": "password123",
	})
	if w.Code != http.StatusConflict {
		t.Errorf("Ожидали 409, получили %d", w.Code)
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	app := setupTestApp(t)
	app.registerUser(t, "testuser", "test@test.com", "password123")

	w := app.request("POST", "/v1/auth/register", map[string]string{
		"username": "testuser", "email": "other@test.com", "password": "password123",
	})
	if w.Code != http.StatusConflict {
		t.Errorf("Ожидали 409, получили %d", w.Code)
	}
}

func TestLogin(t *testing.T) {
	app := setupTestApp(t)
	app.registerUser(t, "testuser", "test@test.com", "password123")

	resp := app.loginUser(t, "test@test.com", "password123")
	if resp.Tokens.AccessToken == "" {
		t.Error("access_token пустой при логине")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	app := setupTestApp(t)
	app.registerUser(t, "testuser", "test@test.com", "password123")

	w := app.request("POST", "/v1/auth/login", map[string]string{
		"email": "test@test.com", "password": "wrongpassword",
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Ожидали 401, получили %d", w.Code)
	}
}

func TestLoginNonExistent(t *testing.T) {
	app := setupTestApp(t)

	w := app.request("POST", "/v1/auth/login", map[string]string{
		"email": "noone@test.com", "password": "password123",
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Ожидали 401, получили %d", w.Code)
	}
}

func TestRefresh(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	w := app.request("POST", "/v1/auth/refresh", map[string]string{
		"refresh_token": resp.Tokens.RefreshToken,
	})
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d: %s", w.Code, w.Body.String())
	}

	var tokens tokenPair
	json.NewDecoder(w.Body).Decode(&tokens)
	if tokens.AccessToken == "" {
		t.Error("Новый access_token пустой")
	}
}

func TestLogout(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	w := app.authRequest("POST", "/v1/auth/logout", resp.Tokens.AccessToken, map[string]string{
		"refresh_token": resp.Tokens.RefreshToken,
	})
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}

	// Повторный refresh должен провалиться
	w = app.request("POST", "/v1/auth/refresh", map[string]string{
		"refresh_token": resp.Tokens.RefreshToken,
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Ожидали 401 после логаута, получили %d", w.Code)
	}
}

// ==================== ПОСТЫ ====================

func TestCreatePost(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	w := app.authRequest("POST", "/v1/posts", resp.Tokens.AccessToken, map[string]string{
		"content": "Мой первый пост!",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("Ожидали 201, получили %d: %s", w.Code, w.Body.String())
	}

	var post map[string]any
	json.NewDecoder(w.Body).Decode(&post)
	if post["content"] != "Мой первый пост!" {
		t.Errorf("content = %v", post["content"])
	}
}

func TestCreatePostNoAuth(t *testing.T) {
	app := setupTestApp(t)

	w := app.request("POST", "/v1/posts", map[string]string{
		"content": "Без авторизации",
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Ожидали 401, получили %d", w.Code)
	}
}

func TestCreatePostEmptyContent(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	w := app.authRequest("POST", "/v1/posts", resp.Tokens.AccessToken, map[string]string{
		"content": "",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидали 400, получили %d", w.Code)
	}
}

func TestDeleteOwnPost(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	// Создаём пост
	w := app.authRequest("POST", "/v1/posts", resp.Tokens.AccessToken, map[string]string{
		"content": "Для удаления",
	})
	var post map[string]any
	json.NewDecoder(w.Body).Decode(&post)
	postID := int(post["id"].(float64))

	// Удаляем
	w = app.authRequest("DELETE", fmt.Sprintf("/v1/posts/%d", postID), resp.Tokens.AccessToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteOtherPost(t *testing.T) {
	app := setupTestApp(t)
	user1 := app.registerUser(t, "user1", "user1@test.com", "password123")
	user2 := app.registerUser(t, "user2", "user2@test.com", "password123")

	// user1 создаёт пост
	w := app.authRequest("POST", "/v1/posts", user1.Tokens.AccessToken, map[string]string{
		"content": "Пост юзера 1",
	})
	var post map[string]any
	json.NewDecoder(w.Body).Decode(&post)
	postID := int(post["id"].(float64))

	// user2 пытается удалить
	w = app.authRequest("DELETE", fmt.Sprintf("/v1/posts/%d", postID), user2.Tokens.AccessToken, nil)
	if w.Code != http.StatusForbidden {
		t.Errorf("Ожидали 403, получили %d", w.Code)
	}
}

func TestGetFeed(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	// Создаём 3 поста
	for i := 1; i <= 3; i++ {
		app.authRequest("POST", "/v1/posts", resp.Tokens.AccessToken, map[string]string{
			"content": fmt.Sprintf("Пост %d", i),
		})
	}

	w := app.request("GET", "/v1/feed", nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}

	var posts []map[string]any
	json.NewDecoder(w.Body).Decode(&posts)
	if len(posts) != 3 {
		t.Errorf("Ожидали 3 поста, получили %d", len(posts))
	}
}

// ==================== КОММЕНТАРИИ ====================

func TestCreateComment(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	// Создаём пост
	w := app.authRequest("POST", "/v1/posts", resp.Tokens.AccessToken, map[string]string{
		"content": "Пост для комментов",
	})
	var post map[string]any
	json.NewDecoder(w.Body).Decode(&post)
	postID := int(post["id"].(float64))

	// Добавляем комментарий
	w = app.authRequest("POST", fmt.Sprintf("/v1/posts/%d/comments", postID), resp.Tokens.AccessToken, map[string]string{
		"content": "Мой комментарий",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("Ожидали 201, получили %d: %s", w.Code, w.Body.String())
	}
}

func TestGetComments(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	// Создаём пост
	w := app.authRequest("POST", "/v1/posts", resp.Tokens.AccessToken, map[string]string{
		"content": "Пост для комментов",
	})
	var post map[string]any
	json.NewDecoder(w.Body).Decode(&post)
	postID := int(post["id"].(float64))

	// Добавляем 2 комментария
	for i := 1; i <= 2; i++ {
		app.authRequest("POST", fmt.Sprintf("/v1/posts/%d/comments", postID), resp.Tokens.AccessToken, map[string]string{
			"content": fmt.Sprintf("Коммент %d", i),
		})
	}

	w = app.request("GET", fmt.Sprintf("/v1/posts/%d/comments", postID), nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}

	var comments []map[string]any
	json.NewDecoder(w.Body).Decode(&comments)
	if len(comments) != 2 {
		t.Errorf("Ожидали 2 комментария, получили %d", len(comments))
	}
}

func TestGetCommentsEmpty(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	// Создаём пост без комментов
	w := app.authRequest("POST", "/v1/posts", resp.Tokens.AccessToken, map[string]string{
		"content": "Пост без комментов",
	})
	var post map[string]any
	json.NewDecoder(w.Body).Decode(&post)
	postID := int(post["id"].(float64))

	w = app.request("GET", fmt.Sprintf("/v1/posts/%d/comments", postID), nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}

	var comments []map[string]any
	json.NewDecoder(w.Body).Decode(&comments)
	if len(comments) != 0 {
		t.Errorf("Ожидали 0 комментариев, получили %d", len(comments))
	}
}

// ==================== ПОДПИСКИ ====================

func TestFollow(t *testing.T) {
	app := setupTestApp(t)
	user1 := app.registerUser(t, "user1", "user1@test.com", "password123")
	user2 := app.registerUser(t, "user2", "user2@test.com", "password123")

	user2ID := int(user2.User["id"].(float64))
	w := app.authRequest("POST", fmt.Sprintf("/v1/users/%d/follow", user2ID), user1.Tokens.AccessToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d: %s", w.Code, w.Body.String())
	}
}

func TestFollowSelf(t *testing.T) {
	app := setupTestApp(t)
	user := app.registerUser(t, "testuser", "test@test.com", "password123")

	userID := int(user.User["id"].(float64))
	w := app.authRequest("POST", fmt.Sprintf("/v1/users/%d/follow", userID), user.Tokens.AccessToken, nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидали 400, получили %d", w.Code)
	}
}

func TestUnfollow(t *testing.T) {
	app := setupTestApp(t)
	user1 := app.registerUser(t, "user1", "user1@test.com", "password123")
	user2 := app.registerUser(t, "user2", "user2@test.com", "password123")

	user2ID := int(user2.User["id"].(float64))

	// Подписываемся
	app.authRequest("POST", fmt.Sprintf("/v1/users/%d/follow", user2ID), user1.Tokens.AccessToken, nil)

	// Отписываемся
	w := app.authRequest("DELETE", fmt.Sprintf("/v1/users/%d/follow", user2ID), user1.Tokens.AccessToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}
}

func TestGetFollowers(t *testing.T) {
	app := setupTestApp(t)
	user1 := app.registerUser(t, "user1", "user1@test.com", "password123")
	user2 := app.registerUser(t, "user2", "user2@test.com", "password123")

	user2ID := int(user2.User["id"].(float64))
	app.authRequest("POST", fmt.Sprintf("/v1/users/%d/follow", user2ID), user1.Tokens.AccessToken, nil)

	w := app.request("GET", fmt.Sprintf("/v1/users/%d/followers", user2ID), nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}

	var followers []map[string]any
	json.NewDecoder(w.Body).Decode(&followers)
	if len(followers) != 1 {
		t.Errorf("Ожидали 1 подписчика, получили %d", len(followers))
	}
}

func TestGetFollowing(t *testing.T) {
	app := setupTestApp(t)
	user1 := app.registerUser(t, "user1", "user1@test.com", "password123")
	user2 := app.registerUser(t, "user2", "user2@test.com", "password123")

	user1ID := int(user1.User["id"].(float64))
	user2ID := int(user2.User["id"].(float64))
	app.authRequest("POST", fmt.Sprintf("/v1/users/%d/follow", user2ID), user1.Tokens.AccessToken, nil)

	w := app.request("GET", fmt.Sprintf("/v1/users/%d/following", user1ID), nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}

	var following []map[string]any
	json.NewDecoder(w.Body).Decode(&following)
	if len(following) != 1 {
		t.Errorf("Ожидали 1 подписку, получили %d", len(following))
	}
}

func TestFollowingFeed(t *testing.T) {
	app := setupTestApp(t)
	user1 := app.registerUser(t, "user1", "user1@test.com", "password123")
	user2 := app.registerUser(t, "user2", "user2@test.com", "password123")

	// user2 пишет пост
	app.authRequest("POST", "/v1/posts", user2.Tokens.AccessToken, map[string]string{
		"content": "Пост от user2",
	})

	// user1 подписывается на user2
	user2ID := int(user2.User["id"].(float64))
	app.authRequest("POST", fmt.Sprintf("/v1/users/%d/follow", user2ID), user1.Tokens.AccessToken, nil)

	// user1 проверяет ленту подписок
	w := app.authRequest("GET", "/v1/feed/following", user1.Tokens.AccessToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}

	var posts []map[string]any
	json.NewDecoder(w.Body).Decode(&posts)
	if len(posts) != 1 {
		t.Errorf("Ожидали 1 пост в ленте подписок, получили %d", len(posts))
	}
}

// ==================== ЛАЙКИ ====================

func TestLike(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	// Создаём пост
	w := app.authRequest("POST", "/v1/posts", resp.Tokens.AccessToken, map[string]string{
		"content": "Пост для лайка",
	})
	var post map[string]any
	json.NewDecoder(w.Body).Decode(&post)
	postID := int(post["id"].(float64))

	// Лайкаем
	w = app.authRequest("POST", fmt.Sprintf("/v1/posts/%d/like", postID), resp.Tokens.AccessToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}
}

func TestLikeIdempotent(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	w := app.authRequest("POST", "/v1/posts", resp.Tokens.AccessToken, map[string]string{
		"content": "Пост для двойного лайка",
	})
	var post map[string]any
	json.NewDecoder(w.Body).Decode(&post)
	postID := int(post["id"].(float64))

	// Лайкаем дважды — ошибки быть не должно (ON CONFLICT DO NOTHING)
	app.authRequest("POST", fmt.Sprintf("/v1/posts/%d/like", postID), resp.Tokens.AccessToken, nil)
	w = app.authRequest("POST", fmt.Sprintf("/v1/posts/%d/like", postID), resp.Tokens.AccessToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Повторный лайк: ожидали 200, получили %d", w.Code)
	}
}

func TestUnlike(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	w := app.authRequest("POST", "/v1/posts", resp.Tokens.AccessToken, map[string]string{
		"content": "Пост для анлайка",
	})
	var post map[string]any
	json.NewDecoder(w.Body).Decode(&post)
	postID := int(post["id"].(float64))

	// Лайкаем и убираем лайк
	app.authRequest("POST", fmt.Sprintf("/v1/posts/%d/like", postID), resp.Tokens.AccessToken, nil)
	w = app.authRequest("DELETE", fmt.Sprintf("/v1/posts/%d/like", postID), resp.Tokens.AccessToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}
}

func TestFeedWithLikeCount(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	// Создаём пост и лайкаем его
	w := app.authRequest("POST", "/v1/posts", resp.Tokens.AccessToken, map[string]string{
		"content": "Пост с лайками",
	})
	var post map[string]any
	json.NewDecoder(w.Body).Decode(&post)
	postID := int(post["id"].(float64))

	app.authRequest("POST", fmt.Sprintf("/v1/posts/%d/like", postID), resp.Tokens.AccessToken, nil)

	// Проверяем ленту с авторизацией
	w = app.authRequest("GET", "/v1/feed", resp.Tokens.AccessToken, nil)
	var posts []map[string]any
	json.NewDecoder(w.Body).Decode(&posts)

	if len(posts) == 0 {
		t.Fatal("Лента пустая")
	}

	likesCount := posts[0]["likes_count"].(float64)
	isLiked := posts[0]["is_liked"].(bool)

	if likesCount != 1 {
		t.Errorf("likes_count = %v, ожидали 1", likesCount)
	}
	if !isLiked {
		t.Error("is_liked = false, ожидали true")
	}
}

// ==================== ПРОФИЛИ ====================

func TestGetProfile(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	userID := int(resp.User["id"].(float64))
	w := app.request("GET", fmt.Sprintf("/v1/users/%d", userID), nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}

	var profile map[string]any
	json.NewDecoder(w.Body).Decode(&profile)
	if profile["username"] != "testuser" {
		t.Errorf("username = %v", profile["username"])
	}
}

func TestGetMe(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	w := app.authRequest("GET", "/v1/users/me", resp.Tokens.AccessToken, nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d: %s", w.Code, w.Body.String())
	}

	var user map[string]any
	json.NewDecoder(w.Body).Decode(&user)
	if user["username"] != "testuser" {
		t.Errorf("username = %v", user["username"])
	}
}

func TestUpdateBio(t *testing.T) {
	app := setupTestApp(t)
	resp := app.registerUser(t, "testuser", "test@test.com", "password123")

	w := app.authRequest("PUT", "/v1/users/me", resp.Tokens.AccessToken, map[string]string{
		"bio": "Привет, я разработчик!",
	})
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d: %s", w.Code, w.Body.String())
	}

	var user map[string]any
	json.NewDecoder(w.Body).Decode(&user)
	if user["bio"] != "Привет, я разработчик!" {
		t.Errorf("bio = %v", user["bio"])
	}
}

// ==================== HEALTHCHECK ====================

func TestHealthCheck(t *testing.T) {
	app := setupTestApp(t)

	w := app.request("GET", "/v1/health", nil)
	if w.Code != http.StatusOK {
		t.Errorf("Ожидали 200, получили %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %v", resp["status"])
	}
}

// ==================== ВАЛИДАЦИЯ ====================

func TestRegisterShortPassword(t *testing.T) {
	app := setupTestApp(t)

	w := app.request("POST", "/v1/auth/register", map[string]string{
		"username": "user", "email": "u@t.com", "password": "123",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидали 400, получили %d", w.Code)
	}
}

func TestRegisterMissingFields(t *testing.T) {
	app := setupTestApp(t)

	w := app.request("POST", "/v1/auth/register", map[string]string{
		"username": "", "email": "u@t.com", "password": "password123",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидали 400, получили %d", w.Code)
	}
}
