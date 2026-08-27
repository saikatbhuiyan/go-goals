package main

import (
	"log"

	serverapp "github.com/saikatbhuiyan/go-goals/apps/api/internal/app/server"
)

func main() {
	if err := serverapp.Run(); err != nil {
		log.Fatal(err)
	}
}
