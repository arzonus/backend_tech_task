package main

import (
	"context"
	"log"

	"github.com/PxyUp/backend_tech_task/internal/api/app"
)

func main() {
	a, err := app.NewApp()
	if err != nil {
		log.Println(err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := a.Run(ctx); err != nil {
		log.Println(err)
	}
}
