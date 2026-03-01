package main

import (
"pharmacy-pos/api/app"
)

func main() {
	server := app.Routes{}
	server.StartGin()
}
