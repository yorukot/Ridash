package models

import "time"

// TeamMember represents a member of a team
type TeamMember struct {
	ID        int64     `json:"id,string" example:"175928847299117063"`      // Unique identifier for the team member
	TeamID    int64     `json:"team_id,string" example:"175928847299117063"` // Team ID
	UserID    int64     `json:"user_id,string" example:"175928847299117063"` // User ID
	Role      Role      `json:"role" example:"member"`                       // Role of the member in the team
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T12:00:00Z"`   // Timestamp when the member was added
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T12:00:00Z"`   // Timestamp when the member was last updated
}
