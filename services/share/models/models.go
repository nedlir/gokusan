package models

// ShareRecord is stored in Redis under "share:<token>".
type ShareRecord struct {
	FileID    string `json:"fileId"`
	CreatedBy string `json:"createdBy"`
}

// CreateShareRequest is the POST /share request body.
type CreateShareRequest struct {
	FileID string `json:"fileId" binding:"required"`
	TTL    int    `json:"ttl"    binding:"required,min=1"`
}

// FileMetadata is the subset of the Metadata Service response we care about.
type FileMetadata struct {
	ID      string `json:"id"`
	OwnerID string `json:"ownerId"`
	Status  string `json:"status"`
}

// MetadataResponse wraps the Metadata Service GET /files/:id response.
type MetadataResponse struct {
	Success bool         `json:"success"`
	File    FileMetadata `json:"file"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
