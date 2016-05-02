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
	closer.Bind(cleanup)
	closer.Checked(run, true)
}

func run() error {
	fmt.Println("Will fatal in 10 seconds...")
	time.Sleep(10 * time.Second)
	closer.Fatalln("KAWABANGA!")
	return nil
}

func cleanup() {
	fmt.Print("Hang on! I'm closing some DBs, wiping some trails...")
	time.Sleep(3 * time.Second)
	fmt.Println("  Done.")
}
