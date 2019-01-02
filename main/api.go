package main

import (
	"github.com/alexmorten/events-api"
)

func main() {
	s := api.NewServer(":3000")
	s.Init()
	s.Run()
}
