// Пакет db предназначен для управления соединениями с базой данных SQLite
package db

// Подключаемые пакеты
import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// Глобальная переменная DB хранит открытое соединение с базой данных
var DB *sql.DB

// Строковая константа schema содержит SQL-код для создания таблицы 'scheduler'
const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    repeat VARCHAR(128) NOT NULL DEFAULT '' 
);
CREATE INDEX idx_date ON scheduler(date); 
`

// Функция Init инициализирует базу данных, создавая новый файл и схему таблиц при первом запуске
func Init(dbFile string) error {
	// Переменная install принимает значение true, если база данных ещё не существует
	install := false
	// Проверяем существование файла базы данных
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		install = true
	}

	// Открываем соединение с базой данных SQLite
	var err error
	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		// Возвращаем ошибку, если не удалось открыть базу данных
		return fmt.Errorf("error connecting to the database: %w", err)
	}

	// Если требуется установка схемы базы данных
	if install {
		// Выполняем создание таблицы и индекса
		_, err := DB.Exec(schema)
		if err != nil {
			// Возврат ошибки, если таблица не была создана
			return fmt.Errorf("error when creating a table in the database: %w", err)
		}
		// Уведомление о успешной установке
		fmt.Println("The table was created successfully.")
	}

	// Без ошибок возвращаем nil
	return nil
}
