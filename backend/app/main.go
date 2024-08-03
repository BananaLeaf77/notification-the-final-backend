package main

import (
	"fmt"
	"notification/config"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	db, err := config.BootDB()
	if err != nil {
		fmt.Println("error : ", err)
	}

	defer db.Close()

}
