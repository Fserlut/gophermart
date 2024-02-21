package main

import (
	"fmt"

	"github.com/Fserlut/gophermart/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	fmt.Println(cfg)
}
