package main

import (
	"fmt"
	"log"
	"time"

	"github.com/xlab/closer"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	closer.Bind(cleanupFunc)

	go func() {
		// do some pseudo background work
		fmt.Println("10 seconds to go...")
		time.Sleep(10 * time.Second)
		closer.Close()
	}()

	closer.Hold()
}

func cleanupFunc() {
	fmt.Print("Hang on! I'm closing some DBs, wiping some trails..")
	time.Sleep(3 * time.Second)
	fmt.Println("  Done.")
}
