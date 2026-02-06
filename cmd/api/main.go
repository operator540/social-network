package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"social-network/internal/config"
	"social-network/internal/database"
	"social-network/internal/handler"
	"social-network/internal/repository"
	"social-network/internal/service"
)

func main() {

	// Загружаем конфигурацию
	cfg := config.Load()


	// Подключаемся к PostgreSQL
	db, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	// Применяем миграции
	if err := database.RunMigrations(db, "migrations"); err != nil {
		log.Fatal("Ошибка миграций: ", err)
	}


	// Создаём директорию для аватарок
	os.MkdirAll("web/uploads", 0755)


	// === Инициализация слоёв (DI через конструкторы) ===


	// Репозитории
	userRepo := repository.NewUserRepo(db)
	postRepo := repository.NewPostRepo(db)
	commentRepo := repository.NewCommentRepo(db)
	followRepo := repository.NewFollowRepo(db)
	likeRepo := repository.NewLikeRepo(db)
	tokenRepo := repository.NewTokenRepo(db)


	// Сервисы
	authService := service.NewAuthService(userRepo, tokenRepo, cfg.JWTSecret)
	userService := service.NewUserService(userRepo)
	postService := service.NewPostService(postRepo)
	commentService := service.NewCommentService(commentRepo)
	followService := service.NewFollowService(followRepo)
	likeService := service.NewLikeService(likeRepo)


	// Хендлер + роутер
	h := handler.NewHandler(authService, userService, postService, commentService, followService, likeService)
	router := h.Routes()


	// HTTP-сервер
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}


	// Запуск в горутине
	go func() {
		log.Printf("Сервер запущен на порту %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()


	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Завершение работы сервера...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Ошибка при остановке сервера: ", err)
	}

	log.Println("Сервер остановлен")
}
