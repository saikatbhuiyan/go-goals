package main

import (
	"log"

	serverapp "github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/app/server"
)

func main() {
	if err := serverapp.Run(); err != nil {
		log.Fatal(err)
	}
}
