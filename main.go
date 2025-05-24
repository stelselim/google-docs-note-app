package main

import (
	"context"
	"fmt"
	"log"
	"my-notes-in-google-docs/controller"
	"my-notes-in-google-docs/service"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	envErr := godotenv.Load(".local.env")
	if envErr != nil {
		log.Fatalf("Error finding environment variables: %s", envErr)
	}
	port, addr := os.Getenv("PORT"), os.Getenv("LISTEN_ADDR")
	if port == "" {
		port = "8080"
	}
	if addr == "" {
		addr = "localhost"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "note app is running")
	})
	http.HandleFunc("/notes", controller.GetNotesHandler)
	http.HandleFunc("/note", controller.PostNotesHandler)

	listenAddr := net.JoinHostPort(addr, port)

	log.Println("---------")
	log.Printf("Server running at %s", listenAddr)

	ctx := context.Background()
	_, docErr := service.ConnectGoogleDocs(ctx)
	_, driveErr := service.ConnectGoogleDrive(ctx)

	if docErr != nil || driveErr != nil {
		log.Fatalf("Error at getting and connecting to Google Services: %s -  %s", docErr, driveErr)
	}

	err := http.ListenAndServe(listenAddr, nil)
	log.Fatalf("Error running on: %s, err: %s", listenAddr, err)

}
