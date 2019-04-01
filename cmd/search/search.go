package main

import (
	"fmt"

	"github.com/alexmorten/events-api/models"
	"github.com/alexmorten/events-api/search"
)

func main() {
	client, err := search.NewClient("http://0.0.0.0:9200")
	if err != nil {
		panic(err)
	}

	client.FuzzyNameSearch("Club", "something", func(props map[string]interface{}) {
		club := models.ClubFromProps(props)
		fmt.Println(club)
	})
}
