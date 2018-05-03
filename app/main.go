package main

import (
	"fmt"

	"github.com/jetuuuu/youtube2audio/app/config"
	"github.com/jetuuuu/youtube2audio/app/rest"
)

func main() {
	c := config.New("")
	c, _ = c.Reload()
	fmt.Println("Start server with config ", c)
	s := rest.New(c)
	s.Run()
}
