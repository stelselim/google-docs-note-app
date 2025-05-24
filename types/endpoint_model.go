package types

import (
	"mime/multipart"
)

type NoteCreatePostRequest struct {
	Title  string                  `json:"title"`
	Note   string                  `json:"note"`
	Tags   []string                `json:"tags,omitempty"`
	Images []*multipart.FileHeader `json:"images,omitempty"`
}

type NoteGetResponse struct {
	Title string `json:"title"`
	Note  string `json:"note"`
	Tags  string `json:"tags,omitempty"`
	Date  string `json:"date,omitempty"`
}
