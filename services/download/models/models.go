package models

import "time"

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// File mirrors the Metadata Service file shape.
type File struct {
	ID         string    `json:"id"`
	OwnerID    string    `json:"ownerId"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mimeType"`
	Status     string    `json:"status"`
	StorageKey string    `json:"storageKey"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// MetadataFileResponse is the JSON envelope returned by GET /files/:id.
type MetadataFileResponse struct {
	Success bool `json:"success"`
	File    File `json:"file"`
}
