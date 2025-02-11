package controllers

import (
	"ass3_part2/db/migrations" // импорт вашего пакета для работы с БД
	"ass3_part2/models"
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// Payment описывает входные данные платежа.
type Payment struct {
	UserID         uint        `json:"user_id"`
	SubscriptionID uint        `json:"subscription_id"`
	PaymentForm    PaymentForm `json:"payment_form"`
}

// PaymentForm содержит данные для оплаты.
type PaymentForm struct {
	CardNumber     string `json:"card_number"`
	ExpirationDate string `json:"expiration_date"`
	CVV            string `json:"cvv"`
}

// maskCard возвращает номер карты с замаскированными первыми цифрами (оставляет видимыми только последние 4 цифры).
func maskCard(cardNumber string) string {
	if len(cardNumber) < 4 {
		return "****"
	}
	return fmt.Sprintf("**** **** **** %s", cardNumber[len(cardNumber)-4:])
}

// generateFiscalReceiptPDF генерирует PDF-файл с фискальным чеком на английском языке.
func generateFiscalReceiptPDF(companyName string, transactionNumber uint, orderDate time.Time,
	itemName string, unitPrice float64, quantity int, clientName string, encryptedCard string) ([]byte, error) {

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header: Company/Project name
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, companyName)
	pdf.Ln(12)

	// Receipt details in English
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Transaction Number: %d", transactionNumber))
	pdf.Ln(10)

	pdf.Cell(40, 10, fmt.Sprintf("Order Date and Time: %s", orderDate.Format("2006-01-02 15:04:05")))
	pdf.Ln(10)

	pdf.Cell(40, 10, fmt.Sprintf("Item/Service: %s", itemName))
	pdf.Ln(10)

	pdf.Cell(40, 10, fmt.Sprintf("Unit Price: %.2f", unitPrice))
	pdf.Ln(10)

	pdf.Cell(40, 10, fmt.Sprintf("Quantity: %d", quantity))
	pdf.Ln(10)

	total := unitPrice * float64(quantity)
	pdf.Cell(40, 10, fmt.Sprintf("Total Amount: %.2f", total))
	pdf.Ln(10)

	pdf.Cell(40, 10, fmt.Sprintf("Client Name: %s", clientName))
	pdf.Ln(10)

	pdf.Cell(40, 10, fmt.Sprintf("Payment Method: %s", encryptedCard))
	pdf.Ln(10)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// PaySubscription обрабатывает запрос на оплату подписки.
// В рамках обработки:
//   - Проверяется корректность данных, в том числе срок действия карты.
//   - Создаются записи о подписке и транзакции.
//   - Генерируется PDF‑чек (на английском языке).
//   - Чек отправляется на электронную почту клиента через микросервис.
//   - Статус транзакции обновляется до "completed".
//   - Возвращается JSON с информацией о платеже.
func PaySubscription(w http.ResponseWriter, r *http.Request) {
	// Для формирования JSON-ответов устанавливаем Content-Type.
	w.Header().Set("Content-Type", "application/json")

	var payment Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Invalid JSON"})
		return
	}

	// Проверка обязательных платежных данных.
	if payment.PaymentForm.CardNumber == "" || payment.PaymentForm.ExpirationDate == "" || payment.PaymentForm.CVV == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Invalid payment details"})
		return
	}

	// Проверка срока действия карты.
	// Предполагается, что срок действия передаётся в формате "01/2006" (месяц/год).
	expirationTime, err := time.Parse("01/2006", payment.PaymentForm.ExpirationDate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Invalid expiration date format"})
		return
	}
	if expirationTime.Before(time.Now()) {
		// Если карта просрочена – имитируем отказ в оплате.
		w.WriteHeader(http.StatusPaymentRequired) // Код 402 Payment Required
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Payment rejected: Card expired"})
		return
	}

	// Рассчитываем период подписки.
	var subscription models.PremiumSubscription
	// Находим подписку по payment.SubscriptionID (предполагается, что модель содержит поля Period и Price).
	db.DB.First(&subscription, payment.SubscriptionID)

	startDate := time.Now()
	endDate := startDate.Add(time.Hour * 24 * time.Duration(subscription.Period)) // subscription.Period – количество дней

	// Создание записи о подписке пользователя.
	userSubscription := models.UserSubscription{
		UserID:         payment.UserID,
		SubscriptionID: payment.SubscriptionID,
		StartDate:      startDate.Format(time.RFC3339),
		EndDate:        endDate.Format(time.RFC3339),
		CreatedAt:      time.Now().Format(time.RFC3339),
		UpdatedAt:      time.Now().Format(time.RFC3339),
	}
	db.DB.Create(&userSubscription)

	// Создание записи транзакции с первоначальным статусом "paid".
	transaction := models.Transaction{
		SubscriptionID: payment.SubscriptionID,
		Status:         "paid",
		CreatedAt:      time.Now().Format(time.RFC3339),
		UpdatedAt:      time.Now().Format(time.RFC3339),
	}
	db.DB.Create(&transaction)

	// Получаем данные пользователя для отправки email (например, email и имя).
	var user models.User
	if err := db.DB.First(&user, payment.UserID).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "User not found"})
		return
	}
	// Используем имя пользователя из БД.
	clientName := user.Name

	// Генерация PDF‑чека (на английском языке).
	pdfBytes, err := generateFiscalReceiptPDF(
		"Example Corp",                           // Company/Project name
		transaction.ID,                           // Transaction Number
		time.Now(),                               // Order Date and Time
		"Premium Subscription",                   // Item/Service
		100,                                      // Unit Price (предполагается, что это поле есть в модели подписки)
		1,                                        // Quantity
		clientName,                               // Client Name
		maskCard(payment.PaymentForm.CardNumber), // Payment Method (masked card number)
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Error generating PDF receipt"})
		return
	}

	// Отправка PDF‑чека на электронную почту клиента через микросервис.
	// URL сервиса для отправки email можно задать через переменную окружения EMAIL_SERVICE_URL.
	emailServiceURL := os.Getenv("EMAIL_SERVICE_URL")
	if emailServiceURL == "" {
		emailServiceURL = "http://localhost:8080/send-email"
	}

	// Подготовка JSON-данных для email.
	emailData := map[string]string{
		"to":      user.Email,
		"subject": "Payment Receipt - Example Corp",
		"body":    "Dear " + clientName + ",\n\nPlease find attached your payment receipt.\n\nThank you for your purchase.",
	}
	emailDataJSON, err := json.Marshal(emailData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Error preparing email data"})
		return
	}

	// Формирование multipart/form-data запроса для отправки email.
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Добавляем поле "json" с данными письма.
	fw, err := writer.CreateFormField("json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Error creating form field"})
		return
	}
	_, err = fw.Write(emailDataJSON)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Error writing email JSON data"})
		return
	}

	// Добавляем PDF‑чек как вложение.
	fw, err = writer.CreateFormFile("file", "receipt.pdf")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Error creating form file"})
		return
	}
	_, err = fw.Write(pdfBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Error attaching PDF file"})
		return
	}
	writer.Close()

	req, err := http.NewRequest("POST", emailServiceURL, &b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Error creating email request"})
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Error sending email receipt"})
		return
	}
	defer resp.Body.Close()

	// Обновляем статус транзакции до "completed".
	transaction.Status = "completed"
	transaction.UpdatedAt = time.Now().Format(time.RFC3339)
	db.DB.Save(&transaction)

	// Подготовка данных для ответа.
	responseData := map[string]interface{}{
		"payment":           payment,
		"user_subscription": userSubscription,
		"transaction":       transaction,
		"subscription":      subscription,
		"message":           "Payment successful. Receipt has been sent to " + user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Status: "success", Message: "Payment successful", Data: responseData})
}
