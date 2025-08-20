// Пакет api "nextdate.go" предназначен для обработки HTTP-запросов и расчета дат повторного события
package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Функция afterNow проверяет, находится ли дата date позднее момента now
func afterNow(date, now time.Time) bool {
	return date.After(now)
}

// Функция convertToInts разбивает входную строку формата "1,2,3" на массив целых чисел
func convertToInts(input string) ([]int, error) {
	if input == "" {
		return nil, errors.New("empty input string") // Входная строка пустая
	}
	// Разделяем строку на отдельные элементы
	parts := strings.Split(input, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		// Конвертируем каждую часть в целое число
		num, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer from input: %w", err)
		}
		result = append(result, num)
	}
	return result, nil
}

// Функция validWeekdays проверяет, являются ли числа валидными днями недели (от 1 до 7)
func validWeekdays(days []int) bool {
	for _, d := range days {
		if d < 1 || d > 7 {
			return false
		}
	}
	return true
}

// Функция validMonthDays проверяет правильность указанных дней месяца
// Допустимы значения от 1 до 31, а также "-1", "-2" (для последних дней месяца)
func validMonthDays(days []int) bool {
	for _, d := range days {
		if d == 0 || d < -2 || d > 31 {
			return false
		}
	}
	return true
}

// Функция validMonths проверяет, что указанные месяцы находятся в диапазоне от 1 до 12
func validMonths(months []int) bool {
	for _, m := range months {
		if m < 1 || m > 12 {
			return false
		}
	}
	return true
}

// Функция lastDayOfMonth возвращает последний день указанного месяца
func lastDayOfMonth(t time.Time) int {
	// Переходим на следующий месяц и находим первый день
	nextMonth := t.AddDate(0, 1, 0)
	firstOfNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, t.Location())
	// Вычитаем один день, чтобы получить последний день текущего месяца
	lastDay := firstOfNextMonth.AddDate(0, 0, -1)
	return lastDay.Day()
}

// Функция CalculateNextDate рассчитывает ближайшую подходящую дату согласно указанному правилу повторения
func CalculateNextDate(current time.Time, startDateStr, repeatRule string) (string, error) {
	// Парсим начальную дату
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		return "", errors.New("incorrect start date format")
	}

	// Проверяем, что правило повторения указано
	if repeatRule == "" {
		return "", errors.New("repetition rule is empty")
	}

	// Парсим правила повторения
	parts := strings.Fields(repeatRule)
	if len(parts) == 0 {
		return "", errors.New("repetition rule is not specified")
	}

	// Обрабатываем различные типы правил повторения
	switch parts[0] {
	case "d":
		// Правила повторения по дням (например, каждые N дней)
		if len(parts) != 2 {
			return "", errors.New("format error for day-based rules")
		}
		interval, err := strconv.Atoi(parts[1]) // Количество дней между повторами
		if err != nil || interval <= 0 || interval > 400 {
			return "", errors.New("invalid interval value")
		}

		// Проверяем, не прошла ли уже начальная дата
		if !startDate.Before(current) {
			startDate = startDate.AddDate(0, 0, interval)
			return startDate.Format(layout), nil
		}

		// Циклически добавляем интервалы, пока не найдем будущую дату
		for {
			startDate = startDate.AddDate(0, 0, interval)
			if afterNow(startDate, current) {
				break
			}
		}
		return startDate.Format(layout), nil

	case "y":
		// Ежегодное повторение
		for {
			startDate = startDate.AddDate(1, 0, 0) // Добавляем год
			if afterNow(startDate, current) {
				break
			}
		}
		return startDate.Format(layout), nil

	case "w":
		// Повторение по дням недели (например, "w 1,3,5" означает понедельники, среду и пятницу)
		if len(parts) != 2 {
			return "", errors.New("format error for weekly rules")
		}
		daysOfWeek, err := convertToInts(parts[1])
		if err != nil {
			return "", errors.New("incorrect format for days of the week")
		}
		if !validWeekdays(daysOfWeek) {
			return "", errors.New("days of the week must be between 1 and 7")
		}

		// Поиск ближайшего подходящего дня недели
		for {
			startDate = startDate.AddDate(0, 0, 1)
			weekday := int(startDate.Weekday()) // День недели в формате Go (суббота = 6, воскресенье = 0)
			if weekday == 0 {
				weekday = 7 // Приводим воскресенье к номеру 7
			}
			for _, d := range daysOfWeek {
				if d == weekday && afterNow(startDate, current) {
					return startDate.Format(layout), nil
				}
			}
		}

	case "m":
		// Повторение по дням месяца (например, "m 10,20" или "m -1,-2 1,2,3")
		if len(parts) < 2 || len(parts) > 3 {
			return "", errors.New("incorrect format for monthly rules")
		}

		monthDays, err := convertToInts(parts[1])
		if err != nil {
			return "", errors.New("incorrect format for days")
		}
		if !validMonthDays(monthDays) {
			return "", errors.New("month days must be within acceptable values")
		}

		var months []int
		if len(parts) == 3 {
			months, err = convertToInts(parts[2])
			if err != nil {
				return "", errors.New("incorrect format for months")
			}
			if !validMonths(months) {
				return "", errors.New("months must be between 1 and 12")
			}
		} else {
			// Если не указаны конкретные месяцы, применяем ко всем месяцам
			months = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		}

		// Перебираем даты, пока не найдем подходящее событие
		for {
			startDate = startDate.AddDate(0, 0, 1)
			currMonth := int(startDate.Month())
			currDay := startDate.Day()

			for _, m := range months {
				if currMonth != m {
					continue
				}
				for _, md := range monthDays {
					// Обычные положительные дни месяца
					if md > 0 && currDay == md {
						if afterNow(startDate, current) {
							return startDate.Format(layout), nil
						}
					} else if md < 0 {
						// Специальные случаи для последнего (-1) и предпоследнего (-2) дней месяца
						last := lastDayOfMonth(startDate)
						if currDay == last+md+1 {
							if afterNow(startDate, current) {
								return startDate.Format(layout), nil
							}
						}
					}
				}
			}

			// Ограничение по количеству итераций для предотвращения бесконечного цикла
			if startDate.After(current.AddDate(100, 0, 0)) {
				return "", errors.New("unable to find an appropriate date within a reasonable amount of time")
			}
		}

	default:
		return "", errors.New("unsupported repetition rule")
	}
}

// nextDayHandler — обработчик HTTP-запроса /api/nextdate, возвращающий следующую дату события
func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что запрос выполнен методом GET
	if r.Method != http.MethodGet {
		http.Error(w, "Unsupported Method", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем параметры из запроса
	nowParam := r.FormValue("now")
	startParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")

	// Если параметр "now" не передан, используем текущую дату
	if nowParam == "" {
		nowParam = time.Now().Format(layout)
	}

	// Парсим значение параметра "now" в тип time.Time
	now, err := time.Parse(layout, nowParam)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing 'now' parameter: %v", err), http.StatusBadRequest)
		return
	}

	// Рассчитываем ближайшую дату
	next, err := CalculateNextDate(now, startParam, repeatParam)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error calculating next date: %v", err), http.StatusBadRequest)
		return
	}

	// Отправляем найденную дату обратно клиенту
	// Доработана обработка ошибок вызова WRITE
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	if _, err := w.Write([]byte(next)); err != nil {
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		return
	}
}
