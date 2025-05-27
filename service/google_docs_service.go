package service

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"my-notes-in-google-docs/types"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

const (
	SEPERATOR_TEXT = "**********END**********"
	NOTE_SEPERATOR = "\n\n" + SEPERATOR_TEXT + "\n\n"
)

var (
	GoogleDocService *docs.Service
	GoogleDocNoteId  string
	initDocOnce      sync.Once
	initDocsErr      error
)

func ConnectGoogleDocs(ctx context.Context) (*docs.Service, error) {
	initDocOnce.Do(func() {
		docService, serviceErr := docs.NewService(ctx,
			option.WithScopes(docs.DocumentsScope),
		)
		if serviceErr != nil {
			initDocsErr = fmt.Errorf("an error occured connecting to the Google Docs: %s", serviceErr)
			return
		} else {
			log.Println("-- --")
			log.Println("Succesfully connected to Google Docs")
		}

		noteDocId := os.Getenv("GOOGLE_DOCS_ID")
		if noteDocId == "" {
			initDocsErr = fmt.Errorf("error reading GOOGLE_DOCS_ID from .env")
			return
		}
		docRef := docService.Documents.Get(noteDocId)
		doc, DocumentErr := docRef.Do()

		if DocumentErr != nil {
			initDocsErr = fmt.Errorf("error getting the Google document by id")
			return
		}
		log.Println("Document Found with ID:", noteDocId)
		log.Println("Doc Title:", doc.Title)

		GoogleDocService = docService
		GoogleDocNoteId = noteDocId
	})

	return GoogleDocService, initDocsErr
}

func GetGoogleDocsDocument() (*docs.Document, error) {
	if GoogleDocService == nil {
		return nil, fmt.Errorf("error initialiazing the Google Doc Service")
	}
	doc, docErr := GoogleDocService.Documents.Get(GoogleDocNoteId).Do()
	if docErr != nil {
		return nil, fmt.Errorf("error getting the document by id: %s", docErr)
	}

	return doc, nil
}

func GetDocContent() ([]types.NoteGetResponse, error) {
	docRef := GoogleDocService.Documents.Get(GoogleDocNoteId)
	noteDocument, documentErr := docRef.Do()

	notes := []types.NoteGetResponse{}

	if documentErr != nil {
		return nil, fmt.Errorf("an error occured: %s ", documentErr)
	}

	fullNoteText := ""

	for _, element := range noteDocument.Body.Content {
		if element.Paragraph != nil {
			for _, elem := range element.Paragraph.Elements {
				if elem.TextRun != nil {
					if elem.TextRun.Content != "\n" {
						fullNoteText += elem.TextRun.Content
					}
				}
			}
		}
	}

	noteList := strings.Split(fullNoteText, SEPERATOR_TEXT)

	for _, note := range noteList {
		noteLines := strings.Split(note, "\n")

		if noteLines[0] == "" {
			noteLines = noteLines[1:]
		}
		if len(noteLines) < 4 {
			continue
		}

		newNote := types.NoteGetResponse{
			Title: noteLines[0],
			Date:  noteLines[1],
			Tags:  noteLines[2],
			Note:  strings.Join(noteLines[3:], "\n"),
		}
		notes = append(notes, newNote)
	}

	return notes, nil
}

func SaveNoteToGoogleDocs(note types.NoteCreatePostRequest) error {
	// Create request to insert text at the start (index 1)
	requests := []*docs.Request{}
	requests = append(requests, getImageRequest(note.Images)...)
	requests = append(requests, getNoteRequest(note.Note, 16)...)
	requests = append(requests, getNoteRequest(strings.Join(note.Tags, ","), 14)...)
	requests = append(requests, getDateTextRequest()...)
	requests = append(requests, getTitleRequest(note.Title)...)
	requests = append(requests, getSeperatorRequest()...)

	_, err := GoogleDocService.Documents.BatchUpdate(
		GoogleDocNoteId,
		&docs.BatchUpdateDocumentRequest{
			Requests: requests,
		}).Do()

	if err != nil {
		return err
	}

	log.Println("Note added successfully!")
	return nil
}

func getSeperatorRequest() []*docs.Request {
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

func getDateTextRequest() []*docs.Request {
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

func getTitleRequest(title string) []*docs.Request {
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

func getNoteRequest(note string, fontSize float64) []*docs.Request {
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

func getImageRequest(files []*multipart.FileHeader) []*docs.Request {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var requests []*docs.Request

	// Upload All Images Concurrently
	for _, fileHeader := range files {
		wg.Add(1)

		go func(fh *multipart.FileHeader) {
			defer wg.Done()

			file, err := fh.Open()
			if err != nil {
				log.Printf("Error oppening file: %s", err)
				return
			}

			defer file.Close()

			fileId, uploadErr := UploadImageToDrive(file, fh)
			if uploadErr != nil {
				log.Printf("Error uploading file to drive: %s", uploadErr)
				return
			}

			newRequest := &docs.Request{
				InsertInlineImage: &docs.InsertInlineImageRequest{
					Location: &docs.Location{
						Index: 1,
					},
					Uri: GetFileUrlById(fileId),
					ObjectSize: &docs.Size{
						Height: &docs.Dimension{Magnitude: 160, Unit: "PT"},
						Width:  &docs.Dimension{Magnitude: 240, Unit: "PT"},
					},
				},
			}

			mu.Lock()
			requests = append(requests, newRequest)
			mu.Unlock()

		}(fileHeader)
	}

	wg.Wait()
	return requests
}
