package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Получаем параметры подключения из переменных окружения или используем значения по умолчанию
	host := getEnv("PGHOST", "master")
	port := getEnv("PGPORT", "5432")
	user := getEnv("PGUSER", "postgres")
	password := getEnv("PGPASSWORD", "postgres")
	dbname := getEnv("PGDATABASE", "app_db")
	sslmode := getEnv("PGSSLMODE", "disable")

	// Формируем строку подключения
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	fmt.Println("Attempting to connect to PostgreSQL...")
	fmt.Printf("Connection string: %s\n", connStr)

	// Пробуем подключиться с повторными попытками
	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			fmt.Printf("Error opening connection: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = db.Ping()
		if err == nil {
			break
		}

		fmt.Printf("Attempt %d: Error connecting to database: %v\n", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		fmt.Printf("Failed to connect to database after multiple attempts: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully connected to PostgreSQL!")

	// Выполняем простой запрос
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		fmt.Printf("Error executing query: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("PostgreSQL version: %s\n", version)

	// Выполняем запрос для получения списка таблиц
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
	`)
	if err != nil {
		fmt.Printf("Error listing tables: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("\nAvailable tables:")
	var tableName string
	for rows.Next() {
		err := rows.Scan(&tableName)
		if err != nil {
			fmt.Printf("Error scanning table name: %v\n", err)
			continue
		}
		fmt.Printf("- %s\n", tableName)
	}

	fmt.Println("\nConnection test completed successfully!")
}

// Вспомогательная функция для получения переменных окружения с значениями по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
