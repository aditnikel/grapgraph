package main

import (
	"fmt"
	"log"

	"github.com/aditnikel/grapgraph/graph"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env (no-op in prod if not present)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment")
	}

	client, err := graph.NewFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Seeding graph...")
	if err := graph.SeedMoneyGraph(client); err != nil {
		log.Fatal(err)
	}

	log.Println("Tracing money from account a1")
	res, _ := graph.TraceMoney(client, "a1")
	fmt.Printf("%#v\n", res)
}
