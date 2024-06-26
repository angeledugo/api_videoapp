package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/angeledugo/video-app-api/controllers"
	"github.com/angeledugo/video-app-api/database"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	err := loadEnv()
	if err != nil {
		fmt.Printf("Error loading .env file: %s\n", err)
		return
	}
	router := mux.NewRouter()
	database.Connect()

	router.HandleFunc("/api/videos", controllers.GetVideos).Methods("GET")
	router.HandleFunc("/api/videos/upload", controllers.UploadVideo).Methods("POST")
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	handler := c.Handler(router)

	log.Fatal(http.ListenAndServe(":8000", handler))
}

func loadEnv() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	return nil
}
