package models

// DocsSharePermission represents the permission level for a shared document
type DocsSharePermission string

// DocsSharePermission constants
const (
	DocsSharePermissionRead  DocsSharePermission = "read"  // User can read the document
	DocsSharePermissionWrite DocsSharePermission = "write" // User can read and write the document
)

// DocsShare represents a document share with a user
type DocsShare struct {
	ID         int64               `json:"id,string" example:"175928847299117063"`          // Unique identifier for the share
	DocumentID int64               `json:"document_id,string" example:"175928847299117063"` // Document ID being shared
	UserID     int64               `json:"user_id,string" example:"175928847299117063"`     // User ID with whom the document is shared
	Roles      DocsSharePermission `json:"roles" example:"read"`                            // Permission level for this share
}
