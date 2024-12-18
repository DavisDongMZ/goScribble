package main

import (
    "log"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

// Pixel model without board_id
type Pixel struct {
    ID          int       `gorm:"primaryKey;autoIncrement"`
    X           int       `gorm:"not null;uniqueIndex:x_y_idx"`
    Y           int       `gorm:"not null;uniqueIndex:x_y_idx"`
    Color       string    `gorm:"size:7;default:#FFFFFF"`
    LastUpdated time.Time `gorm:"autoUpdateTime"`
}

var db *gorm.DB

func initDB() {
    // Please replace the credentials as needed
    dsn := "host=localhost user=postgres password=123456 dbname=rplace port=5432 sslmode=disable"
    var err error
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    // Auto-migrate the Pixel table
    err = db.AutoMigrate(&Pixel{})
    if err != nil {
        log.Fatalf("Failed to migrate database: %v", err)
    }

    // Clear the database on startup
    clearDatabase()

    log.Println("Database connected, migrated, and cleared!")
}

// clearDatabase removes all pixel data
func clearDatabase() {
    err := db.Exec("TRUNCATE pixels RESTART IDENTITY CASCADE").Error
    if err != nil {
        log.Fatalf("Failed to clear database: %v", err)
    }
    log.Println("Database cleared!")
}

// savePixel saves or updates a pixel with the given coordinates and color
func savePixel(x, y int, color string) error {
    pixel := Pixel{
        X:     x,
        Y:     y,
        Color: color,
    }
    return db.Save(&pixel).Error
}

