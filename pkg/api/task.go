// Пакет api "task.go" обрабатывает HTTP-запросы, относящиеся к задачам
package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sam39rus/go_final_project/pkg/db"
)

// Функция taskHandler распределяет запросы на ресурс /api/task по различным методам HTTP
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleCreateTask(w, r) // Создание новой задачи
	case http.MethodGet:
		handleGetTask(w, r) // Получение задачи по ID
	case http.MethodPut:
		handleUpdateTask(w, r) // Обновление задачи
	case http.MethodDelete:
		handleDeleteTask(w, r) // Удаление задачи
	default:
		http.Error(w, "The method is not supported", http.StatusMethodNotAllowed) // Неверный метод HTTP
	}
}

// Функция handleCreateTask обрабатывает создание новой задачи
func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	addTaskHandler(w, r) // Использует основную функцию для создания задачи
}

// Функция handleGetTask обрабатывает получение задачи по ID
func handleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id") // Получаем ID задачи из параметров запроса
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error": "ID not specified"}) // Требуется указать ID задачи
		return
	}

	task, err := db.GetTask(id) // Получаем задачу из базы данных
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJson(w, http.StatusNotFound, map[string]string{
				"error": "Issue not found"}) // Задача не найдена
			return
		}

		fmt.Printf("error when getting a task from the database: %v\n", err)                       // Логи ошибки
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сервера"}) // Внутренняя ошибка сервера
		return
	}

	writeJson(w, http.StatusOK, task) // Возвращаем задачу в формате JSON
}

// Функция handleUpdateTask обрабатывает обновление задачи
func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task) // Декодируем тело запроса в структуру Task
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("JSON decoding error: %v", err)}) // Ошибка декодирования JSON
		return
	}

	if task.ID == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error": "The issue ID is not specified"}) // Необходимо указать ID задачи
		return
	}

	if task.Title == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error": "The issue title is not specified"}) // Необходимо указать название задачи
		return
	}

	err = checkDate(&task) // Проверяем и корректируем дату задачи
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error": err.Error()}) // Неправильная дата или повторение
		return
	}

	err = db.UpdateTask(&task) // Обновляем задачу в базе данных
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error()}) // Ошибка обновления задачи
		return
	}

	writeJson(w, http.StatusOK, map[string]string{}) // Успех, возвращаем пустой объект
}

// Функция handleDeleteTask обрабатывает удаление задачи
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id") // Получаем ID задачи из параметров запроса
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error": "ID not specified"}) // Необходимо указать ID задачи
		return
	}

	err := db.DeleteTask(id) // Удаляем задачу из базы данных
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error()}) // Ошибка удаления задачи
		return
	}

	writeJson(w, http.StatusOK, map[string]string{}) // Успех, возвращаем пустой объект
}

// Функция doneTaskHandler помечает задачу как выполненную
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "The method is not supported", http.StatusMethodNotAllowed) // Только POST-метод поддерживается
		return
	}

	id := r.URL.Query().Get("id") // Получаем ID задачи из параметров запроса
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error": "ID not specified"}) // Необходимо указать ID задачи
		return
	}

	task, err := db.GetTask(id) // Получаем задачу из базы данных
	if err != nil {
		writeJson(w, http.StatusNotFound, map[string]string{
			"error": "Issue not found"}) // Задача не найдена
		return
	}

	if task.Repeat == "" {
		// Если задача одноразовая, удаляем её
		err := db.DeleteTask(id)
		if err != nil {
			writeJson(w, http.StatusInternalServerError, map[string]string{
				"error": err.Error()}) // Ошибка удаления задачи
			return
		}
		writeJson(w, http.StatusOK, map[string]string{}) // Успех, возвращаем пустой объект
		return
	}

	// Если задача повторяющаяся, вычисляем следующую дату
	layout := "20060102"
	baseDate, err := time.Parse(layout, task.Date) // Преобразуем дату задачи в тип time.Time
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error": "incorrect date format in the issue"}) // Некорректный формат даты
		return
	}

	nextDate, err := CalculateNextDate(baseDate, task.Date, task.Repeat) // Расчет следующей даты
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{
			"error": err.Error()}) // Ошибка вычисления следующей даты
		return
	}

	err = db.UpdateDate(nextDate, id) // Обновляем дату задачи в базе данных
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error()}) // Ошибка обновления даты
		return
	}

	writeJson(w, http.StatusOK, map[string]string{}) // Успех, возвращаем пустой объект
}
