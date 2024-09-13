package user

import (
	"database/sql"

	"avi/internal/database"
	"avi/internal/model"

	"github.com/google/uuid"
)

type UserRepo struct {
	db *sql.DB
}

func (repo *UserRepo) GetUserByName(name string) (user *model.User, err error) {
	selectQuery := `
		SELECT id, username, first_name, last_name, created_at, updated_at
		FROM employee WHERE username = $1 
	`
	user = &model.User{}
	row := repo.db.QueryRow(selectQuery, name)
	err = row.Scan(
		&user.Id,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return
}

func (repo *UserRepo) GetUserById(id uuid.UUID) (user *model.User, err error) {
	selectQuery := `
		SELECT id, username, first_name, last_name, created_at, updated_at
		FROM employee WHERE id = $1 
	`
	user = &model.User{}
	row := repo.db.QueryRow(selectQuery, id)
	err = row.Scan(
		&user.Id,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return
	}

	return
}

func NewRepo() (repo *UserRepo, err error) {
	db, err := database.Connect()
	if err != nil {
		return
	}
	repo = &UserRepo{db: db}
	return
}
