package models

import "time"

// Role represents the role of a team member
type Role string

// Role constants
const (
	RoleOwner  Role = "owner"  // Team owner with full permissions
	RoleAdmin  Role = "admin"  // Team administrator with elevated permissions
	RoleMember Role = "member" // Regular team member
)

// Team represents a team in the system
type Team struct {
	ID        int64     `json:"id,string" example:"175928847299117063"`       // Unique identifier for the team
	OwnerID   int64     `json:"owner_id,string" example:"175928847299117063"` // Owner user ID
	Name      string    `json:"name" example:"My Team"`                       // Team name (max 50 characters)
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T12:00:00Z"`    // Timestamp when the team was created
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T12:00:00Z"`    // Timestamp when the team was last updated
}
