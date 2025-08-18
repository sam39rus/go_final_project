// Пакет db "task.go" предназначен для работы с базой данных
package db

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
