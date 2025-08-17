// Пакет server реализует функциональность HTTP-файлового сервера
package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// Run запускает HTTP-файл-сервер на заданном или стандартном порте
func Run() error {
	// Устанавливаем стандартный порт сервера
	port := 7540

	// Проверяем переменную окружения TODO_PORT
	if p := os.Getenv("TODO_PORT"); p != "" {
		// Пробуем преобразовать значение порта в число
		val, err := strconv.Atoi(p)
		if err == nil {
			// Если успешно, используем указанный порт
			port = val
		} else {
			// Иначе возвращаем ошибку преобразования порта
			return fmt.Errorf("invalid TODO_PORT: %w", err)
		}
	}

	// Создаем файловое хранилище для статического веб-контента из папки web
	fs := http.FileServer(http.Dir("web"))

	// Регистрируем обработчик запросов "/"
	http.Handle("/", fs)

	// Формируем адрес прослушивания (например ":7540")
	addr := fmt.Sprintf(":%d", port)

	// Сообщение о старте сервера
	fmt.Printf("The server is running at: http://localhost%s\n", addr)

	// Начинаем прослушивание и обслуживание входящих соединений
	return http.ListenAndServe(addr, nil)
}
