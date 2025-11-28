package repository

import (
	"context"
	"ridash/models"

	"github.com/jackc/pgx/v5"
)

// CreateTeam creates a new team
func CreateTeam(ctx context.Context, tx pgx.Tx, team models.Team) error {
	query := `INSERT INTO teams (id, owner_id, name, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5)`

	_, err := tx.Exec(ctx, query,
		team.ID,
		team.OwnerID,
		team.Name,
		team.CreatedAt,
		team.UpdatedAt,
	)

	return err
}

// CreateTeamMember creates a new team member
func CreateTeamMember(ctx context.Context, tx pgx.Tx, teamMember models.TeamMember) error {
	query := `INSERT INTO team_members (id, team_id, user_id, role, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := tx.Exec(ctx, query,
		teamMember.ID,
		teamMember.TeamID,
		teamMember.UserID,
		teamMember.Role,
		teamMember.CreatedAt,
		teamMember.UpdatedAt,
	)

	return err
}

// GetTeamByID retrieves a team by its ID
func GetTeamByID(ctx context.Context, tx pgx.Tx, teamID int64) (*models.Team, error) {
	query := `SELECT id, owner_id, name, created_at, updated_at
	          FROM teams
	          WHERE id = $1
	          LIMIT 1`

	var team models.Team
	err := tx.QueryRow(ctx, query, teamID).Scan(
		&team.ID,
		&team.OwnerID,
		&team.Name,
		&team.CreatedAt,
		&team.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // Not an error, just not found
	}

	if err != nil {
		return nil, err
	}

	return &team, nil
}

// UpdateTeam updates an existing team
func UpdateTeam(ctx context.Context, tx pgx.Tx, teamID int64, name string, updatedAt any) error {
	query := `UPDATE teams
	          SET name = $1, updated_at = $2
	          WHERE id = $3`

	_, err := tx.Exec(ctx, query, name, updatedAt, teamID)
	return err
}

// DeleteTeamMembersByTeamID removes all members linked to a team
func DeleteTeamMembersByTeamID(ctx context.Context, tx pgx.Tx, teamID int64) error {
	query := `DELETE FROM team_members WHERE team_id = $1`
	_, err := tx.Exec(ctx, query, teamID)
	return err
}

// DeleteTeam deletes a team by ID
func DeleteTeam(ctx context.Context, tx pgx.Tx, teamID int64) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := tx.Exec(ctx, query, teamID)
	return err
}
