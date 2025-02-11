package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var (
	Logger *zap.Logger
	file   *os.File
)

// NewLogger инициализирует логгер, записывающий логи в консоль и файл app.log в формате JSON
func NewLogger() error {
	var err error

	// Открываем или создаем файл для логов
	file, err = os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	// Настройка JSON-энкодера
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",                        // Ключ для времени
		LevelKey:       "level",                       // Ключ для уровня логов
		MessageKey:     "msg",                         // Ключ для сообщения
		CallerKey:      "caller",                      // Ключ для вызова
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // Время в формате ISO8601
		EncodeLevel:    zapcore.CapitalLevelEncoder,   // Уровень логов (INFO, ERROR и т.д.)
		EncodeCaller:   zapcore.ShortCallerEncoder,    // Короткий путь к файлу вызова
		EncodeDuration: zapcore.StringDurationEncoder, // Продолжительность в строковом формате
	}

	// Создаем JSON-энкодер
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Создаем вывод в консоль
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Создаем вывод в файл
	fileWriter := zapcore.AddSync(file)

	// Настраиваем комбинированный вывод в консоль и файл
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, consoleWriter, zapcore.DebugLevel), // JSON-логи в консоль
		zapcore.NewCore(jsonEncoder, fileWriter, zapcore.DebugLevel),    // JSON-логи в файл
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return nil
}

func CloseLogger() {
	if file != nil {
		_ = file.Close()
	}
}
