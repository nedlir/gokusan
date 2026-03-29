package models

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// FileUploadedEvent is published to the file.uploaded Kafka topic.
type FileUploadedEvent struct {
	FileID     string `json:"fileId"`
	OwnerID    string `json:"ownerId"`
	FileName   string `json:"fileName"`
	FileSize   int64  `json:"fileSize"`
	MimeType   string `json:"mimeType"`
	StorageKey string `json:"storageKey"`
}
