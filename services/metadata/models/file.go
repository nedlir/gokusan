package models

import "time"

type File struct {
	ID         string    `json:"id" db:"id"`
	OwnerID    string    `json:"ownerId" db:"owner_id"`
	Name       string    `json:"name" db:"name"`
	Size       int64     `json:"size" db:"size"`
	MimeType   string    `json:"mimeType" db:"mime_type"`
	Status     string    `json:"status" db:"status"`
	StorageKey string    `json:"storageKey" db:"storage_key"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

// Kafka event types

type FileUploadedEvent struct {
	FileID     string `json:"fileId"`
	OwnerID    string `json:"ownerId"`
	FileName   string `json:"fileName"`
	FileSize   int64  `json:"fileSize"`
	MimeType   string `json:"mimeType"`
	StorageKey string `json:"storageKey"`
}

type FileSanitizedEvent struct {
	FileID     string `json:"fileId"`
	StorageKey string `json:"storageKey"`
}

type FileQuarantinedEvent struct {
	FileID string `json:"fileId"`
}

type FileDeletedEvent struct {
	FileID string `json:"fileId"`
}
