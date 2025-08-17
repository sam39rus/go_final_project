// Пакет api реализует обработку маршрутов API
package api

import (
	"net/http"
)

// Функция Init осуществляет регистрацию обработчиков для различных конечных точек API
func Init() {
	// Настройка маршрута для получения следующего рабочего дня
	http.HandleFunc("/api/nextdate", nextDayHandler)
}
