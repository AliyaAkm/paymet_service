package main

import (
	db "ass3_part2/db/migrations"
	"ass3_part2/logging"
	router2 "ass3_part2/router"
	"context"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	router := router2.NewRouter()
	dbConfig := db.LoadDbConfigFromEnv()
	db.NewDb(dbConfig)
	// Указываем серверу использовать папку "static" для HTML, CSS и JS
	http.Handle("/", http.FileServer(http.Dir("./static")))

	err := logging.NewLogger()
	if err != nil {
		log.Fatal(err)
	}

	// Запускаем сервер на порту 8081
	server := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}
	go func() {
		logging.Logger.Info("Сервер запущен: http://localhost:8081/index")
		if err = server.ListenAndServe(); err != nil {
			logging.Logger.Error("Ошибка запуска сервера:", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	logging.Logger.Info("Получен сигнал завершения, остановка сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logging.Logger.Error("Ошибка при завершении сервера:", zap.Error(err))
	} else {
		logging.Logger.Info("Сервер успешно остановлен.")
	}

	db.CloseDb()
	logging.Logger.Sync()
}
