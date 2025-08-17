// Основной пакет приложения
package main

import (
	"go1f/pkg/server"
	"log"
	"os"

	"github.com/sam39rus/go_final_project/pkg/db"
)

// Главная точка входа программы
func main() {
	// Получаем путь к файлу базы данных из переменной окружения
	dbFile := os.Getenv("TODO_DBFILE")

	// Если переменная не определена, устанавливаем стандартный путь файла БД
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	// Инициализируем подключение к базе данных
	if err := db.Init(dbFile); err != nil {
		// Если произошла ошибка подключения, программа завершится с сообщением об ошибке
		log.Fatal("Error connecting to the database:", err)
	}

	// Запускаем сервер и проверяем наличие ошибок
	if err := server.Run(); err != nil {
		// Если возникла ошибка при запуске сервера, выводим сообщение об ошибке и прекращаем выполнение программы
		log.Fatal("Server creation error:", err)
	}
}
