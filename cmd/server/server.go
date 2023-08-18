package main

import (
	"github.com/baely/weightloss-tracker/internal/server"
)

func main() {
	s, _ := server.NewServer()
	s.Run()
}
