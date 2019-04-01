package main

import (
	"github.com/alexmorten/events-api"

	//import .env file if present
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	s := api.NewServer(":3000")
	s.Init()
	s.Run()
}
