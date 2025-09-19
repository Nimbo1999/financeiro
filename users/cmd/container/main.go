package main

import (
	"os"

	_ "github.com/jackc/pgx/v5"
	"github.com/nimbo1999/financeiro/users/internal/app"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(postgres.Open("postgres://matheus:dev1234@localhost:5432/financeiro_user"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	port := os.Getenv("PORT")
	app := app.New(db)
	if err := app.Run(port); err != nil {
		panic(err)
	}
}
