package main

import (
	"context"
	"log"
)

func main() {
	ctx := context.Background()
	if err := RunStream(ctx); err != nil {
		log.Fatal(err)
	}

}
