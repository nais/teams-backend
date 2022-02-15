package main

import (
	"fmt"
	"time"

	"github.com/nais/console/pkg/version"
)

func main() {
	fmt.Println("hello, world!")
	fmt.Printf("version %s\n", version.Version())
	time.Sleep(24 * time.Hour)
}
