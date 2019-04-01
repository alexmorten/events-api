package main

import (
	"flag"

	"github.com/alexmorten/events-api"

	//import .env file if present
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	config := api.ServerConfig{}

	flag.StringVar(&config.Neo4jAddress, "neo4j_address", "bolt://0.0.0.0:7687", "address to neo4j")
	flag.StringVar(&config.Neo4jAddress, "elastic_address", "http://0.0.0.0:9200", "address to elasticsearch")
	flag.IntVar(&config.Port, "port", 3000, "port the server should listen on for http requests")
	flag.BoolVar(&config.LazyInitializeElastic, "lazily_initialize_elastic", false, "if set to true, creating the connection to elastic_search will be defered until we make a call to it")
	flag.Parse()

	s := api.NewServer(config)
	s.Init()
	s.Run()
}
