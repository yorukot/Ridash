package models

import "time"

// DocsPermission represents the permission level for a document
type DocsPermission string

// DocsPermission constants
const (
	DocsPermissionPrivate     DocsPermission = "private"      // Only owner can access
	DocsPermissionPublic      DocsPermission = "public"       // Anyone can read
	DocsPermissionPublicWrite DocsPermission = "public_write" // Anyone can read and write
)

// Document represents a document in the system
type Document struct {
	ID         int64          `json:"id,string" example:"175928847299117063"`       // Unique identifier for the document
	OwnerID    int64          `json:"owner_id,string" example:"175928847299117063"` // Owner user ID or folder ID
	Name       string         `json:"name" example:"My Document"`                   // Document name
	Permission DocsPermission `json:"permission" example:"private"`                 // Document permission level
	CreatedAt  time.Time      `json:"created_at" example:"2023-01-01T12:00:00Z"`    // Timestamp when the document was created
	UpdatedAt  time.Time      `json:"updated_at" example:"2023-01-01T12:00:00Z"`    // Timestamp when the document was last updated
}
