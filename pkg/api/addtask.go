// Пакет api "addtask.go" содержит логику обработки HTTP-запросов для работы с задачами
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sam39rus/go_final_project/pkg/db"
)

// Функция addTaskHandler обрабатывает POST-запросы на создание новой задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task // Создаем новую задачу

	// Декодируем тело запроса в структуру Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		// Если произошла ошибка декодирования, отправляем соответствующий ответ
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("JSON decoding error: %v", err)})
		return
	}

	// Проверяем обязательное поле Title
	if task.Title == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error": "The issue title is not specified"})
		return
	}

	// Проверяем корректность даты и правила повторения
	err = checkDate(&task)
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error": err.Error()})
		return
	}

	// Добавляем задачу в базу данных
	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("error writing to the database: %v", err)})
		return
	}

	// Возвращаем идентификатор добавленной задачи
	writeJson(w, http.StatusOK, map[string]string{
		"id": fmt.Sprintf("%d", id)})
}

// Функция checkDate проверяет дату и правило повторения задачи и при необходимости меняет дату
func checkDate(task *db.Task) error {
	// Текущая дата и время
	now := time.Now()
	// Оставляем только дату
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	layout := "20060102" // Формат даты (ГГГГММДД)

	// Если дата не указана, ставим текущую дату
	if task.Date == "" {
		task.Date = now.Format(layout)
	}

	// Проверяем корректность введённой даты
	t, err := time.Parse(layout, task.Date)
	if err != nil {
		return fmt.Errorf("incorrect date format")
	}

	// Если указано правило повторения, проверяем и вычисляем правильную дату
	if len(task.Repeat) > 0 {
		next, err := CalculateNextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("incorrect repetition rule: %w", err)
		}

		// Изменяем дату, если исходная дата уже прошла
		if t.Before(now) {
			task.Date = next
		}
	} else {
		// Если повторения нет и дата уже прошла, ставим текущую дату
		if t.Before(now) {
			task.Date = now.Format(layout)
		}
	}

	return nil
}

// Функция writeJson сериализует данные в JSON и записывает их в ответ клиента
func writeJson(w http.ResponseWriter, status int, data any) {
	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// Устанавливаем статус-код ответа
	w.WriteHeader(status)
	// Сериализуем данные в JSON и отправляем клиенту
	_ = json.NewEncoder(w).Encode(data)
}

// Функция taskHandler обрабатывает входящие HTTP-запросы на точку '/api/task', распределяя их по методам
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// Обработка POST-запроса на создание задачи
		addTaskHandler(w, r)
	case http.MethodGet:
		// Обработка GET-запроса на получение конкретной задачи по ID
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJson(w, http.StatusBadRequest, map[string]string{
				"error": "ID not specified"})
			return
		}
		task, err := db.GetTask(id)
		if err != nil {
			writeJson(w, http.StatusNotFound, map[string]string{
				"error": "Issue not found"})
			return
		}
		writeJson(w, http.StatusOK, task)
	case http.MethodPut:
		// Обработка PUT-запроса на обновление задачи
		var task db.Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			writeJson(w, http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("JSON decoding error: %v", err)})
			return
		}
		if task.ID == "" {
			writeJson(w, http.StatusBadRequest, map[string]string{
				"error": "No task identifier provided"})
			return
		}
		if task.Title == "" {
			writeJson(w, http.StatusBadRequest, map[string]string{
				"error": "No task title provided"})
			return
		}
		err = checkDate(&task)
		if err != nil {
			writeJson(w, http.StatusBadRequest, map[string]string{
				"error": err.Error()})
			return
		}
		err = db.UpdateTask(&task)
		if err != nil {
			writeJson(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error()})
			return
		}
		writeJson(w, http.StatusOK, map[string]string{})
	default:
		// Ответ на неподдерживаемые методы
		http.Error(w, "The method is not supported", http.StatusMethodNotAllowed)
	}
}
