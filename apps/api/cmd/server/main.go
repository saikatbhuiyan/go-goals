package main

import (
	"log"

	serverapp "github.com/saikatbhuiyan/go-goals/internal/app/server"
)

func main() {
	if err := serverapp.Run(); err != nil {
		log.Fatal(err)
	}
}
