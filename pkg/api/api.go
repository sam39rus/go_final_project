// Пакет api реализует обработку маршрутов API
package api

import (
	"net/http"
)

// Функция Init осуществляет регистрацию обработчиков для различных конечных точек API
func Init() {
	// Настройка маршрута для получения следующего рабочего дня
	http.HandleFunc("/api/nextdate", nextDayHandler)

	// Маршрут для операций с отдельной задачей
	http.HandleFunc("/api/task", taskHandler)

	// Маршрут для операций с несколькими задачами одновременно
	http.HandleFunc("/api/tasks", tasksHandler)

	// Маршрут для пометки задачи как выполненной
	http.HandleFunc("/api/task/done", doneTaskHandler)
}
