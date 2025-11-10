// infrastructure/persistence/sqlite/session_repository.go
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"osbb-accounting/domain/entity"
	domainErrors "osbb-accounting/domain/errors"
)

// SessionRepository реалізує repository.SessionRepository для SQLite.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository створює новий SessionRepository.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create створює нову сесію.
func (r *SessionRepository) Create(ctx context.Context, session *entity.Session) error {
	query := `
		INSERT INTO sessions (
			user_id, token, expires_at, created_at, ip_address, user_agent
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		session.UserID,
		session.Token,
		session.ExpiresAt.Unix(),
		session.CreatedAt.Unix(),
		session.IPAddress,
		session.UserAgent,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	session.ID = id

	return nil
}

// GetByToken отримує сесію за токеном.
func (r *SessionRepository) GetByToken(ctx context.Context, token string) (*entity.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, ip_address, user_agent
		FROM sessions
		WHERE token = ?
	`

	session := &entity.Session{}
	var expiresAt, createdAt int64
	var ipAddress, userAgent sql.NullString

	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&expiresAt,
		&createdAt,
		&ipAddress,
		&userAgent,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	session.ExpiresAt = time.Unix(expiresAt, 0)
	session.CreatedAt = time.Unix(createdAt, 0)
	session.IPAddress = nullStringToPtr(ipAddress)
	session.UserAgent = nullStringToPtr(userAgent)

	return session, nil
}

// GetByID отримує сесію за ID.
func (r *SessionRepository) GetByID(ctx context.Context, id int64) (*entity.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, ip_address, user_agent
		FROM sessions
		WHERE id = ?
	`

	session := &entity.Session{}
	var expiresAt, createdAt int64
	var ipAddress, userAgent sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&expiresAt,
		&createdAt,
		&ipAddress,
		&userAgent,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by id: %w", err)
	}

	session.ExpiresAt = time.Unix(expiresAt, 0)
	session.CreatedAt = time.Unix(createdAt, 0)
	session.IPAddress = nullStringToPtr(ipAddress)
	session.UserAgent = nullStringToPtr(userAgent)

	return session, nil
}

// GetActiveByUserID отримує всі активні сесії користувача.
func (r *SessionRepository) GetActiveByUserID(ctx context.Context, userID int64) ([]*entity.Session, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, ip_address, user_agent
		FROM sessions
		WHERE user_id = ? AND expires_at > ?
		ORDER BY created_at DESC
	`

	now := time.Now().Unix()
	rows, err := r.db.QueryContext(ctx, query, userID, now)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]*entity.Session, 0)
	for rows.Next() {
		session := &entity.Session{}
		var expiresAt, createdAt int64
		var ipAddress, userAgent sql.NullString

		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.Token,
			&expiresAt,
			&createdAt,
			&ipAddress,
			&userAgent,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		session.ExpiresAt = time.Unix(expiresAt, 0)
		session.CreatedAt = time.Unix(createdAt, 0)
		session.IPAddress = nullStringToPtr(ipAddress)
		session.UserAgent = nullStringToPtr(userAgent)

		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sessions: %w", err)
	}

	return sessions, nil
}

// Update оновлює сесію (наприклад, продовжує час життя).
func (r *SessionRepository) Update(ctx context.Context, session *entity.Session) error {
	query := `
		UPDATE sessions
		SET expires_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		session.ExpiresAt.Unix(),
		session.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrSessionNotFound
	}

	return nil
}

// Delete видаляє сесію (logout).
func (r *SessionRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM sessions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrSessionNotFound
	}

	return nil
}

// DeleteByToken видаляє сесію за токеном.
func (r *SessionRepository) DeleteByToken(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = ?`

	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session by token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrSessionNotFound
	}

	return nil
}

// DeleteAllByUserID видаляє всі сесії користувача.
func (r *SessionRepository) DeleteAllByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM sessions WHERE user_id = ?`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all user sessions: %w", err)
	}

	return nil
}

// DeleteExpired видаляє всі закінчені сесії.
func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at <= ?`

	now := time.Now().Unix()
	result, err := r.db.ExecContext(ctx, query, now)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// CleanupOldSessions видаляє старі закінчені сесії (для cleanup job).
func (r *SessionRepository) CleanupOldSessions(ctx context.Context, olderThanDays int) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at <= ?`

	cutoffTime := time.Now().AddDate(0, 0, -olderThanDays).Unix()
	result, err := r.db.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
