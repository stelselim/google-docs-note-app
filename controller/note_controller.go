package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"my-notes-in-google-docs/service"
	"my-notes-in-google-docs/types"
	"net/http"
	"strings"
)

func PostNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(types.BaseResponse[string]{
			Success: false,
			Message: "method not allowed.",
		})
		return
	}

	multiParseError := r.ParseMultipartForm(10 << 21)
	if multiParseError != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.BaseResponse[string]{
			Success: false,
			Message: "error at parsing the note.",
		})
		return
	}

	// Parse Multipart Form Data
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
		log.Printf("Received file: %s (%d bytes)", fileHeader.Filename, fileHeader.Size)
	}

	saveErr := service.SaveNoteToGoogleDocs(req)
	if saveErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(types.BaseResponse[string]{
			Success: false,
			Message: "note could not be added: " + saveErr.Error(),
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

func GetNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(types.BaseResponse[string]{
			Success: false,
			Message: "method not allowed.",
		})
		return
	}

	notes, noteErr := service.GetDocContent()

	if noteErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.BaseResponse[string]{
			Success: false,
			Message: "error at getting the notes.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(types.BaseResponse[[]types.NoteGetResponse]{
		Success: true,
		Data:    notes,
		Message: fmt.Sprintf("%d note(s) found.", len(notes)),
	})
}
