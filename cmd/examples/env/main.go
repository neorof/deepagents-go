package main

import (
	"fmt"
	"os"
)

func main() {
	key := os.Getenv("ANTHROPIC_API_KEY")
	fmt.Println("ANTHROPIC_API_KEY:", key)
}
