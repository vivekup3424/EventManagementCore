package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/BEOpenSourceCollabs/EventManagementCore/pkg/models"
)

// UserRepository represents the interface for user-related database operations.
type UserRepository interface {
	CreateUser(user *models.UserModel) error
	GetUserByID(id string) (*models.UserModel, error)
	UpdateUser(user *models.UserModel) error
	DeleteUser(id string) error
	GetUserByEmail(email string) (*models.UserModel, error)
	InsertUser(user *models.UserModel) error
}

type sqlUserRepository struct {
	database *sql.DB
}

// NewSQLUserRepository creates and returns a new sql flavoured UserRepository instance.
func NewSQLUserRepository(database *sql.DB) UserRepository {
	return &sqlUserRepository{database: database}
}

// CreateUser inserts a new user into the database.
func (r *sqlUserRepository) CreateUser(user *models.UserModel) error {
	query := `INSERT INTO public.users (username, email, password, first_name, last_name, birth_date, role, verified, about)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	err := r.database.QueryRow(query, user.Username, user.Email, user.Password, user.FirstName, user.LastName, user.BirthDate, user.Role, user.Verified, user.About).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user from the database by its unique ID.
func (r *sqlUserRepository) GetUserByID(id string) (*models.UserModel, error) {
	query := `SELECT * FROM public.users WHERE id = $1`

	user := &models.UserModel{}
	err := r.database.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.BirthDate,
		&user.Role,
		&user.Verified,
		&user.About,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// UpdateUser update a user in the database.
func (r *sqlUserRepository) UpdateUser(user *models.UserModel) error {
	query := `UPDATE public.users SET username = $1, email = $2, password = $3, first_name = $4, last_name = $5, birth_date = $6, role = $7, verified = $8, about = $9, updated_at = $10 WHERE id = $11`

	// This is a guard to prevent any partial user from being submitted.
	// Otherwise it would be possible to accidently empty out columns by passing empty/uninitialized values.
	if user.CreatedAt.Unix() == 0 {
		return fmt.Errorf("unable to update a user that was not loaded from the database")
	}

	rs, err := r.database.Exec(
		query,
		user.Username,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.BirthDate,
		user.Role,
		user.Verified,
		user.About,
		time.Now(),
		user.ID,
	)
	if err != nil {
		return err
	}

	if affected, err := rs.RowsAffected(); affected < 1 {
		if err != nil {
			return err
		}
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser delete a user from the database.
func (r *sqlUserRepository) DeleteUser(id string) error {
	query := `DELETE FROM public.users WHERE id = $1`

	rs, err := r.database.Exec(query, id)
	if err != nil {
		return err
	}

	if affected, err := rs.RowsAffected(); affected < 1 {
		if err != nil {
			return err
		}
		return ErrUserNotFound
	}

	return nil
}

func (r *sqlUserRepository) GetUserByEmail(email string) (*models.UserModel, error) {
	query := `SELECT * FROM public.users WHERE email = $1`

	user := &models.UserModel{}
	err := r.database.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.BirthDate,
		&user.Role,
		&user.Verified,
		&user.About,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

func (r *sqlUserRepository) InsertUser(user *models.UserModel) error {

	//include required fields in columns first
	insertQ := "INSERT INTO public.users (email, password, username"
	valuesQ := "VALUES($1, $2, $3"
	args := []interface{}{user.Email, user.Password, user.Username}
	argsCounter := 3

	if user.FirstName.String != "" {
		argsCounter++
		insertQ += ", first_name"
		valuesQ += fmt.Sprintf(", $%d", argsCounter)
		args = append(args, user.FirstName)
	}

	if user.LastName.String != "" {
		argsCounter++
		insertQ += ", last_name"
		valuesQ += fmt.Sprintf(", $%d", argsCounter)
		args = append(args, user.LastName)
	}

	insertQ += ") "
	valuesQ += ") RETURNING id"

	query := insertQ + valuesQ

	err := r.database.QueryRow(query, args...).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

var (
	ErrUserNotFound = errors.New("user not found") // ErrUserNotFound is returned when a user is not found in the database.
)
