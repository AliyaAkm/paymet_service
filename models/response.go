package models

// Структура для ответа сервера
type Response struct {
	Status  string `json:"status"`  // Статус ответа (success/fail)
	Message string `json:"message"` // Сообщение в ответе
}
