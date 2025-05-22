package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"my-notes-in-google-docs/types"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	NOTE_SEPERATOR = "\n\n**********END**********\n\n"
)

var (
	driveService   *drive.Service
	noteDocService *docs.Service
	noteDocumentId string
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

	http.HandleFunc("/", getNotesHandler)
	http.HandleFunc("/note", postNotesHandler)

	listenAddr := net.JoinHostPort(addr, port)

	log.Println("---------")
	log.Printf("Server running at %s", listenAddr)
	_, docErr := connectGoogleDocs()
	driveService = connectGoogleDrive()
	if docErr != nil {
		log.Fatalf("Error at getting and connecting to Google Docs: %s", docErr)
	}

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

	getDocContent()
}

func postNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintln(w, "Method Not Allowed")
		log.Printf("Not Allowed Method Requested: %s\n", r.Method)
		return
	}
	log.Println("Post Notes Handler")

	// Parse the multipart form (10 MB max memory, larger files go to temp files)
	err := r.ParseMultipartForm(10 << 22)
	if err != nil {
		fmt.Fprint(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	note := r.FormValue("note")
	tags := r.FormValue("tags")

	files := r.MultipartForm.File["images"]

	req := types.NoteCreatePostRequest{
		Title:  title,
		Note:   note,
		Tags:   strings.Split(tags, ","),
		Images: files,
	}

	for _, fileHeader := range req.Images {
		fmt.Printf("Received file: %s (%d bytes)", fileHeader.Filename, fileHeader.Size)
		fmt.Printf("tags: %s, note: %s", req.Tags, req.Note)
	}

	saveErr := saveNoteToGoogleDocs(req)
	if saveErr != nil {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(types.BaseResponse[string]{
			Success: false,
			Message: "Note Could not Added: " + saveErr.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(types.BaseResponse[string]{
		Success: true,
		Message: "A new note added.",
	})
}

func connectGoogleDocs() (*docs.Document, error) {
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

	noteDocId := os.Getenv("GOOGLE_DOCS_ID")
	if noteDocId == "" {
		log.Fatal("Error getting DOC ID ")
	}
	docRef := docService.Documents.Get(noteDocId)
	doc, err := docRef.Do()
	if err != nil {
		fmt.Printf("An error occured: %s ", err)
		return nil, err
	}
	fmt.Println("Document Found with ID:", noteDocId)
	fmt.Println("Doc Title:", doc.Title)

	noteDocService = docService
	noteDocumentId = noteDocId

	return doc, nil
}

func getDocContent() {
	docRef := noteDocService.Documents.Get(noteDocumentId)
	noteDocument, err := docRef.Do()

	if err != nil {
		fmt.Printf("An error occured: %s ", err)
	}

	for _, element := range noteDocument.Body.Content {
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

func saveNoteToGoogleDocs(note types.NoteCreatePostRequest) error {
	// Create request to insert text at the start (index 1)
	requests := []*docs.Request{}
	requests = append(requests, GetImageRequest(note.Images)...)
	requests = append(requests, GetNoteRequest(note.Note, 16)...)
	requests = append(requests, GetNoteRequest(strings.Join(note.Tags, ","), 14)...)
	requests = append(requests, GetDateTextRequest()...)
	requests = append(requests, GetTitleRequest(note.Title)...)
	requests = append(requests, GetSeperatorRequest()...)

	_, err := noteDocService.Documents.BatchUpdate(
		noteDocumentId,
		&docs.BatchUpdateDocumentRequest{
			Requests: requests,
		}).Do()

	if err != nil {
		return err
	}

	fmt.Println("Text inserted successfully!")
	return nil
}

func GetSeperatorRequest() []*docs.Request {
	return []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Text: NOTE_SEPERATOR,
				Location: &docs.Location{
					Index: 1,
				},
			},
		},
		{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				TextStyle: &docs.TextStyle{
					FontSize: &docs.Dimension{
						Magnitude: 14,
						Unit:      "PT",
					},
				},
				Range: &docs.Range{
					StartIndex: 1,
					EndIndex:   int64(len(NOTE_SEPERATOR)) + 1,
				},
				Fields: "FontSize",
			},
		},
	}
}
func GetDateTextRequest() []*docs.Request {
	currentTime := time.Now()
	dateText := currentTime.Format("03:04:05 PM, 2006-01-02")
	return []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Text: dateText + "\n",
				Location: &docs.Location{
					Index: 1,
				},
			},
		},

		{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				TextStyle: &docs.TextStyle{
					Bold: false,
					FontSize: &docs.Dimension{
						Magnitude: 12,
						Unit:      "PT",
					},
				},
				Range: &docs.Range{
					StartIndex: 1,
					EndIndex:   int64(len(dateText)) + 1,
				},
				Fields: "Bold,FontSize",
			},
		},
	}
}

func GetTitleRequest(title string) []*docs.Request {
	return []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Text: title + "\n",
				Location: &docs.Location{
					Index: 1,
				},
			},
		},

		{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				TextStyle: &docs.TextStyle{
					Bold: true,
					FontSize: &docs.Dimension{
						Magnitude: 24,
						Unit:      "PT",
					},
				},
				Range: &docs.Range{
					StartIndex: 1,
					EndIndex:   int64(len(title)) + 1,
				},
				Fields: "Bold,FontSize",
			},
		},
	}
}
func GetNoteRequest(note string, fontSize float64) []*docs.Request {
	if fontSize == 0 {
		fontSize = 16
	}

	return []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Text: note + "\n",
				Location: &docs.Location{
					Index: 1,
				},
			},
		},

		{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				TextStyle: &docs.TextStyle{
					Bold: false,
					FontSize: &docs.Dimension{
						Magnitude: fontSize,
						Unit:      "PT",
					},
				},
				Range: &docs.Range{
					StartIndex: 1,
					EndIndex:   int64(len(note)) + 1,
				},
				Fields: "Bold,FontSize",
			},
		},
	}
}

func GetImageRequest(files []*multipart.FileHeader) []*docs.Request {
	var requests []*docs.Request

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error oppening file: %s", err)
			continue
		}
		defer file.Close()

		fileId, err := uploadToDrive(driveService, file, fileHeader)
		if err != nil {
			log.Printf("Error uploading file to drive: %s", err)
		}

		requests = append(requests, &docs.Request{
			InsertInlineImage: &docs.InsertInlineImageRequest{
				Location: &docs.Location{
					Index: 1,
				},
				Uri: "https://drive.google.com/uc?id=" + fileId + "&sz=w1000",
				ObjectSize: &docs.Size{
					Height: &docs.Dimension{Magnitude: 160, Unit: "PT"},
					Width:  &docs.Dimension{Magnitude: 240, Unit: "PT"},
				},
			},
		})
	}

	return requests
}

func connectGoogleDrive() *drive.Service {
	ctx := context.Background()
	// Create Drive service
	driveService, err := drive.NewService(ctx, option.WithScopes(drive.DriveScope))
	if err != nil {
		log.Fatalf("Unable to create Drive client: %v", err)
	}
	fmt.Println("Successfully connected to Google Drive")
	return driveService
}

// file: your *multipart.FileHeader from request
func uploadToDrive(driveService *drive.Service, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	defer file.Close()

	f := &drive.File{
		Name:     fileHeader.Filename,
		MimeType: fileHeader.Header.Get("Content-Type"),
	}

	uploadedFile, err := driveService.Files.Create(f).Media(file).Do()
	if err != nil {
		return "", err
	}
	// Set permission
	_, err = driveService.Permissions.Create(uploadedFile.Id, &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}).Do()
	if err != nil {
		return "", err
	}

	return uploadedFile.Id, nil
}
