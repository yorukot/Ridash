package models

import "time"

// Folder represents a folder in a team
type Folder struct {
	ID           int64     `json:"id,string" example:"175928847299117063"`                      // Unique identifier for the folder
	TeamID       int64     `json:"team_id,string" example:"175928847299117063"`                 // Team ID this folder belongs to
	Name         string    `json:"name" example:"My Folder"`                                    // Folder name
	ParentFolder *int64    `json:"parent_folder,string,omitempty" example:"175928847299117063"` // Parent folder ID (null for root folders)
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T12:00:00Z"`                   // Timestamp when the folder was created
	UpdatedAt    time.Time `json:"updated_at" example:"2023-01-01T12:00:00Z"`                   // Timestamp when the folder was last updated
}
