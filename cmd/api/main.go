package main

import (
	"github.com/azevedoguigo/thermora_core/internal/config"
	"github.com/azevedoguigo/thermora_core/internal/server"
)

func main() {
	config := config.LoadConfig()
	s := server.NewServer(config)
	s.Run()
}
