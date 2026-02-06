package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

// Connect устанавливает соединение с PostgreSQL
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("БД не отвечает: %w", err)
	}

	// Настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("Подключение к PostgreSQL установлено")
	return db, nil
}

// RunMigrations применяет все .up.sql миграции из указанной директории
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Создаём таблицу для отслеживания применённых миграций
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("не удалось создать таблицу миграций: %w", err)
	}

	// Находим все .up.sql файлы
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("не удалось прочитать директорию миграций: %w", err)
	}

	var upFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	for _, fileName := range upFiles {
		version := strings.TrimSuffix(fileName, ".up.sql")

		// Проверяем, применена ли миграция
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("ошибка проверки миграции %s: %w", version, err)
		}
		if exists {
			continue
		}

		// Читаем и применяем миграцию
		content, err := os.ReadFile(filepath.Join(migrationsDir, fileName))
		if err != nil {
			return fmt.Errorf("не удалось прочитать миграцию %s: %w", fileName, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("не удалось начать транзакцию для %s: %w", fileName, err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка применения миграции %s: %w", fileName, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("не удалось записать версию %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("не удалось зафиксировать миграцию %s: %w", fileName, err)
		}

		log.Printf("Миграция %s применена", fileName)
	}

	return nil
}
