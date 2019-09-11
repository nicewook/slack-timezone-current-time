package main

import (
	"log"
	"net/http"

	api "github.com/nicewook/slack-timezone-current-time/api"
)

func main() {
	// Custom mux https://gist.github.com/reagent/043da4661d2984e9ecb1ccb5343bf438
	handler := http.NewServeMux()

	log.Println("Server started...")

	handler.HandleFunc("/tz", api.TimeZoneCurrentTime)
	handler.HandleFunc("/tzn", api.TimeZoneCurrentTimeNewYork)

	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
