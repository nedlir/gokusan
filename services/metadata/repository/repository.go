package repository

import (
	"context"
	"fmt"

	"github.com/gokusan/metadata/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// InsertFile inserts a new file row with status 'pending'.
func (r *Repository) InsertFile(ctx context.Context, f *models.File) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO files (id, owner_id, name, size, mime_type, status, storage_key, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, 'pending', $6, NOW(), NOW())`,
		f.ID, f.OwnerID, f.Name, f.Size, f.MimeType, f.StorageKey,
	)
	return err
}

// UpdateFileStatus updates the status (and optionally storage_key) of a file by ID.
func (r *Repository) UpdateFileStatus(ctx context.Context, fileID, status, storageKey string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE files
		 SET status = $1, storage_key = COALESCE(NULLIF($2, ''), storage_key), updated_at = NOW()
		 WHERE id = $3`,
		status, storageKey, fileID,
	)
	return err
}

// GetFilesByOwner returns all non-deleted files for a given owner.
func (r *Repository) GetFilesByOwner(ctx context.Context, ownerID string) ([]models.File, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, owner_id, name, size, mime_type, status, COALESCE(storage_key, '') AS storage_key, created_at, updated_at
		 FROM files
		 WHERE owner_id = $1 AND status != 'deleted'
		 ORDER BY created_at DESC`,
		ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.File])
	if err != nil {
		return nil, err
	}
	return files, nil
}

// GetFileByID returns a single file by ID regardless of owner (caller checks ownership).
func (r *Repository) GetFileByID(ctx context.Context, fileID string) (*models.File, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, owner_id, name, size, mime_type, status, COALESCE(storage_key, '') AS storage_key, created_at, updated_at
		 FROM files
		 WHERE id = $1`,
		fileID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	file, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.File])
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", pgx.ErrNoRows)
	}
	return &file, nil
}

// SoftDeleteFile sets status='deleted' for a file owned by ownerID.
// Returns an error if the file is not found or not owned by ownerID.
func (r *Repository) SoftDeleteFile(ctx context.Context, fileID, ownerID string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE files
		 SET status = 'deleted', updated_at = NOW()
		 WHERE id = $1 AND owner_id = $2`,
		fileID, ownerID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("file not found or not owned by user: %w", pgx.ErrNoRows)
	}
	return nil
}
