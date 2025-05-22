package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

func main() {
	envErr := godotenv.Load(".env")
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

	http.HandleFunc("/", getNotesHandler)
	http.HandleFunc("/note", postNotesHandler)

	listenAddr := net.JoinHostPort(addr, port)

	log.Println("---------")
	log.Printf("Server running at %s", listenAddr)
	connectGoogleDocs()

	err := http.ListenAndServe(listenAddr, nil)
	log.Fatalf("Error running on: %s, err: %s", listenAddr, err)

}

func getNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		fmt.Fprintln(w, "Method Not Allowed")
		log.Printf("Not Allowed Method Requested: %s\n", r.Method)
		return
	}
	fmt.Fprintf(w, "Get Notes Handler\n")
}

func postNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintln(w, "Method Not Allowed")
		log.Printf("Not Allowed Method Requested: %s\n", r.Method)
		return
	}
	fmt.Fprintf(w, "Post Notes Handler\n")
}

func connectGoogleDocs() {
	ctx := context.Background()
	docService, err := docs.NewService(ctx,
		option.WithScopes(docs.DocumentsScope),
	)
	if err != nil {
		log.Fatalf("Error connecting to Google Docs: %s", err)
	} else {
		fmt.Println("-- --")
		fmt.Println("Succesfully connected to Google Docs")
	}

	docId := os.Getenv("GOOGLE_DOCS_ID")
	if docId == "" {
		log.Fatal("Error getting DOC ID ")
	}
	docRef := docService.Documents.Get(docId)
	doc, err := docRef.Do()

	if err != nil {
		fmt.Printf("An error occured: %s ", err)
	}

	fmt.Println("Document Found with ID:", docId)
	fmt.Println("Doc Title:", doc.Title)

	for _, element := range doc.Body.Content {
		if element.Paragraph != nil {
			for _, elem := range element.Paragraph.Elements {
				if elem.TextRun != nil {
					if elem.TextRun.Content != "\n" {
					} else {
					}
				}
			}
		}
	}
}
