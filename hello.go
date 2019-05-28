package main

import "fmt"
import "os"

func main() {
  target_url := os.Getenv("TARGET_URL")
  fmt.Printf(target_url)
}