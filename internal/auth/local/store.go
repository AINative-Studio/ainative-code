package local

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth/jwt"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite" // SQLite driver
)

const (
	// BcryptCost is the cost factor for bcrypt hashing (12 rounds)
	BcryptCost = 12

	// LocalTokenDuration is the lifetime of local access tokens
	LocalTokenDuration = 24 * time.Hour

	// LocalRefreshDuration is the lifetime of local refresh tokens
	LocalRefreshDuration = 7 * 24 * time.Hour
)

// Store manages local authentication and credentials.
type Store struct {
	db *sql.DB
}

// User represents a local user account.
type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Session represents an authentication session.
type Session struct {
	ID           int64
	UserID       int64
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

// NewStore creates a new local auth store with SQLite backend.
func NewStore(dbPath string) (*Store, error) {
	// Open database with connection parameters for better concurrency
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for SQLite
	// SQLite works best with a single writer, so limit connections
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	store := &Store{db: db}

	// Initialize database schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// initSchema creates the database tables.
func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		access_token TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_access_token ON sessions(access_token);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// Register creates a new local user with hashed password.
func (s *Store) Register(email, password string) error {
	// Validate inputs
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Hash password with bcrypt (12 rounds)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user
	now := time.Now()
	_, err = s.db.Exec(
		"INSERT INTO users (email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?)",
		email, string(passwordHash), now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Authenticate validates credentials and creates a session.
func (s *Store) Authenticate(email, password string) (*Session, error) {
	// Get user
	var user User
	err := s.db.QueryRow(
		"SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate tokens
	accessToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	now := time.Now()
	expiresAt := now.Add(LocalTokenDuration)

	result, err := s.db.Exec(
		"INSERT INTO sessions (user_id, access_token, refresh_token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)",
		user.ID, accessToken, refreshToken, expiresAt, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	sessionID, _ := result.LastInsertId()

	session := &Session{
		ID:           sessionID,
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
	}

	return session, nil
}

// ValidateToken validates an access token and returns the user ID.
func (s *Store) ValidateToken(accessToken string) (int64, error) {
	var userID int64
	var expiresAt time.Time

	err := s.db.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE access_token = ?",
		accessToken,
	).Scan(&userID, &expiresAt)

	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid token")
	}
	if err != nil {
		return 0, fmt.Errorf("failed to query session: %w", err)
	}

	// Check expiration
	if time.Now().After(expiresAt) {
		return 0, fmt.Errorf("token expired")
	}

	return userID, nil
}

// RefreshSession creates a new session using a refresh token.
func (s *Store) RefreshSession(refreshToken string) (*Session, error) {
	// Get existing session
	var userID int64
	var sessionID int64

	err := s.db.QueryRow(
		"SELECT id, user_id FROM sessions WHERE refresh_token = ?",
		refreshToken,
	).Scan(&sessionID, &userID)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid refresh token")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	// Delete old session
	if _, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", sessionID); err != nil {
		return nil, fmt.Errorf("failed to delete old session: %w", err)
	}

	// Generate new tokens
	accessToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create new session
	now := time.Now()
	expiresAt := now.Add(LocalTokenDuration)

	result, err := s.db.Exec(
		"INSERT INTO sessions (user_id, access_token, refresh_token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)",
		userID, accessToken, newRefreshToken, expiresAt, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	newSessionID, _ := result.LastInsertId()

	session := &Session{
		ID:           newSessionID,
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
	}

	return session, nil
}

// GetUser returns a user by ID.
func (s *Store) GetUser(userID int64) (*User, error) {
	var user User

	err := s.db.QueryRow(
		"SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// DeleteSession deletes a session by access token.
func (s *Store) DeleteSession(accessToken string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE access_token = ?", accessToken)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteAllSessions deletes all sessions for a user.
func (s *Store) DeleteAllSessions(userID int64) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}
	return nil
}

// ToTokenPair converts a session to a JWT token pair format.
func (s *Session) ToTokenPair() *jwt.TokenPair {
	return &jwt.TokenPair{
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
		ExpiresIn:    int64(time.Until(s.ExpiresAt).Seconds()),
		TokenType:    "Bearer",
	}
}

// generateToken generates a random token string.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
