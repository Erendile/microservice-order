package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(user CreateUser) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	id := uuid.New().String()
	query := `INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4)`
	_, err = r.db.Exec(query, id, user.Name, user.Email, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) FindAll() ([]User, error) {
	query := `SELECT id, name, email, password FROM users`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to find all users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, nil
}

func (r *PostgresRepository) FindById(id string) (User, error) {
	query := `SELECT id, name, email, password FROM users WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var user User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("user not found: %w", err)
		}
		return User{}, fmt.Errorf("failed to find user by id: %w", err)
	}

	return user, nil
}

func (r *PostgresRepository) FindByEmail(email string) (User, error) {
	query := `SELECT id, name, email, password FROM users WHERE email = $1`
	row := r.db.QueryRow(query, email)

	var user User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("user not found: %w", err)
		}
		return User{}, fmt.Errorf("failed to find user by email: %w", err)
	}

	return user, nil
}
