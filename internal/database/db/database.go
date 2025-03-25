package db

import (
	"be/internal/database/models"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Dbinstance struct {
	AionDB *gorm.DB
}

var DB Dbinstance

func ConnectDb() {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: false,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Default().Print("Failed to connect to database \n", err)
		os.Exit(2)
	}

	// /*
	// if err := dropAllTables(db); err != nil {
	// 	log.Default().Print("Failed to drop all tables \n", err)
	// }

	// log.Println("running migrations")
	// if err := db.AutoMigrate(&models.User{}, &models.CustomModel{}, &models.TrainingPlan{}, &models.Workout{}, &models.Exercise{}); err != nil {
	// 	log.Default().Print("Failed to migrate database \n", err)
	// }

	// if err := populateData(db); err != nil {
	// log.Default().Print("Failed to populate data \n", err)
	// }
	// */

	log.Printf("Connected to %s - %s\n", dbHost, dbName)

	DB = Dbinstance{
		AionDB: db,
	}
}

func dropAllTables(db *gorm.DB) error {
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	var tables []string
	db.Raw("SHOW TABLES").Scan(&tables)
	for _, table := range tables {
		fmt.Printf("Dropping table %s\n", table)
		db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
	}
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	return nil
}

func truncateAllTables(db *gorm.DB) error {
	db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	var tables []string
	db.Raw("SHOW TABLES").Scan(&tables)
	for _, table := range tables {
		fmt.Printf("Truncating table %s\n", table)
		db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table))
	}
	db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	return nil
}

func populateData(db *gorm.DB) error {
	for i := 0; i < 50; i++ {
		uID, _ := uuid.Parse(gofakeit.UUID())
		ukcID, _ := uuid.Parse(gofakeit.UUID())
		user := models.User{
			KCID:        ukcID,
			FirstName:   gofakeit.FirstName(),
			LastName:    gofakeit.LastName(),
			Nickname:    gofakeit.Username(),
			CustomModel: models.CustomModel{ID: uID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		db.Create(&user)
	}
	fmt.Println("Populated!")
	return nil
}
