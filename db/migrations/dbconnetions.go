package db

import (
	"ass3_part2/models"
	"fmt"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

var DB *gorm.DB

type DbConfig struct {
	Host     string `env:"host"`
	User     string `env:"user"`
	Password string `env:"password"`
	Dbname   string `env:"dbname"`
	Port     string `env:"port"`
	Sslmode  string `env:"sslmode"`
}

func init() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("aaaa")
	}
	dbConfig := LoadDbConfigFromEnv()
	NewDb(dbConfig)

	if err := DB.AutoMigrate(
		&models.User{},
		&models.Movie{},
		&models.Role{},
		&models.PremiumSubscription{},
		&models.UserSubscription{},
		&models.Transaction{},
	); err != nil {
		log.Fatal("Error migrating models: ", err)
	}
}

func NewDb(dbConfig DbConfig) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.Dbname, dbConfig.Port, dbConfig.Sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}

	DB = db
	fmt.Println("Database connected successfully!")

	// Выполняем миграции после успешного подключения
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Movie{},
		&models.Role{},
		&models.PremiumSubscription{},
		&models.UserSubscription{},
		&models.Transaction{},
	); err != nil {
		log.Fatal("Error migrating models: ", err)
	}
}

func CloseDb() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Println("Error retrieving sql.DB from Gorm:", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Println("Error closing the database connection:", err)
	} else {
		fmt.Println("Database connection closed successfully.")
	}
}

func LoadDbConfigFromEnv() DbConfig {
	return DbConfig{
		Host:     os.Getenv("host"),
		User:     os.Getenv("user"),
		Password: os.Getenv("password"),
		Dbname:   os.Getenv("dbname"),
		Port:     os.Getenv("port"),
		Sslmode:  os.Getenv("sslmode"),
	}
}
