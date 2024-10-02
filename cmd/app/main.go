package main

import (
	"effective-mobile/internal/app"

	_ "github.com/swaggo/http-swagger"
)

// @title Online song library
// @version beta 0.1
// @description API Server for  online song library
func main() {
	app.Run()
}
