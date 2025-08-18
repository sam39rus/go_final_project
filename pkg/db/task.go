// Пакет db "task.go" предназначен для работы с базой данных
package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Задача (Task) представляет собой структурированную запись, включающую:
type Task struct {
	ID      string `json:"id"`      // Уникальный идентификатор задачи
	Date    string `json:"date"`    // Дата задачи
	Title   string `json:"title"`   // Название задачи
	Comment string `json:"comment"` // Комментарий к задаче
	Repeat  string `json:"repeat"`  // Правило повторения задачи (если есть)
}

// Функция AddTask добавляет новую задачу в таблицу scheduler
// Параметр task: указатель на структуру Task, представляющую новую задачу
// Возвращает: идентификатор вновь добавленной задачи и возможную ошибку
func AddTask(task *Task) (int64, error) {
	// Подготовленный SQL-запрос для вставки новой задачи
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`

	// Выполнение запроса с параметрами задачи
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		// В случае ошибки при выполнении запроса возвращаем ошибку
		return 0, err
	}

	// Получаем идентификатор последней вставленной записи
	return res.LastInsertId()
}

// Функция Tasks получает список задач из базы данных
// Параметры:
//
//	limit - максимальное количество возвращаемых задач
//	search - необязательное ключевое слово для поиска
//
// Возвращает: срез указателей на задачи (*Task) и возможную ошибку
func Tasks(limit int, search string) ([]*Task, error) {
	// Создаем пустой срез для хранения задач
	tasks := make([]*Task, 0, limit)

	// Определены форматы даты для работы с поиском
	layoutSearch := "02.01.2006" // формат даты для поиска из параметра search
	layoutDB := "20060102"       // формат даты в базе данных

	// Проверяем, является ли поисковый запрос датой
	searchDate, err := time.Parse(layoutSearch, search)
	isDateSearch := (err == nil)

	// Выбор нужной стратегии выборки задач в зависимости от поискового запроса
	var rows *sql.Rows
	if search == "" {
		// Если ничего не искали, выбираем все задачи с сортировкой по дате и ограничением limit
		query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`
		rows, err = DB.Query(query, limit)
	} else if isDateSearch {
		// Если поисковый запрос является датой, ищем задачи строго по этой дате
		searchDateStr := searchDate.Format(layoutDB)
		query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date ASC LIMIT ?`
		rows, err = DB.Query(query, searchDateStr, limit)
	} else {
		// Если поисковый запрос — текст, ищем по названию и комментарию
		likePattern := fmt.Sprintf("%%%s%%", search)
		query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date ASC LIMIT ?`
		rows, err = DB.Query(query, likePattern, likePattern, limit)
	}

	// Обработка возможных ошибок выполнения запроса
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Сбор результатов запроса
	for rows.Next() {
		t := new(Task)
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	// Проверка на наличие ошибок сканирования
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Если tasks равен nil, создаем пустой слайс (чтобы избежать null в JSON)
	if tasks == nil {
		tasks = make([]*Task, 0)
	}

	return tasks, nil
}
