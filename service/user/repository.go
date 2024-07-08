package user

import (
	"database/sql"
	"fmt"

	"github.com/dawumnam/token-trader/types"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (s *Repository) GetUserByEmail(email string) (*types.User, error) {
	query := "SELECT id, firstName, lastName, email, password FROM users WHERE email=?"
	var user types.User
	err := s.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no user with email:%s found", email)
		}
		return nil, err
	}

	return &user, nil
}

func (s *Repository) GetUserById(id int) (*types.User, error) {
	query := "SELECT id, firstName, lastName, email, password FROM users WHERE id=?;"

	var user types.User
	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no user with email:%d found", id)
		}
		return nil, err
	}

	return &user, nil
}

func (s *Repository) CreateUser(u *types.User) error {
	result, err := s.db.Exec("INSERT INTO users (firstName, lastName, email, password) VALUES (?,?,?,?)", u.FirstName, u.LastName, u.Email, u.Password)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	u.ID = int(id)
	return nil
}
