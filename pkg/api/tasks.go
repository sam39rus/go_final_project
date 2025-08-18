// Пакет api "tasks.go" обрабатывает HTTP-запросы, связанные с задачами
package api

import (
	"net/http"

	"github.com/sam39rus/go_final_project/pkg/db"
)

// Структура TasksResp используется для формирования ответа с задачами
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"` // Массив задач
}

// Функция tasksHandler обрабатывает GET-запросы на ресурс /api/tasks
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что использован именно метод GET
	if r.Method != http.MethodGet {
		http.Error(w, "The method is not supported", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры из URL-запроса
	search := r.URL.Query().Get("search") // Фильтр поиска

	// Ограничение количества результатов
	limit := 50

	// Получаем список задач из базы данных
	tasks, err := db.Tasks(limit, search)
	if err != nil {
		// Если произошла ошибка при получении задач, отправляем ответ с ошибкой
		writeJson(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// Формируем и отправляем ответ с результатами
	writeJson(w, http.StatusOK, TasksResp{Tasks: tasks})
}
