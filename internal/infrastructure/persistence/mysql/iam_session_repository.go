package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	domainiam "github.com/rin721/keiyaku-go/internal/domain/iam"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IAMRefreshSessionModel struct {
	ID                  int64      `gorm:"column:id;primaryKey"`
	UserID              int64      `gorm:"column:user_id"`
	RefreshTokenID      string     `gorm:"column:refresh_token_id"`
	Status              string     `gorm:"column:status"`
	ReplacedBySessionID *int64     `gorm:"column:replaced_by_session_id"`
	ExpiresAt           time.Time  `gorm:"column:expires_at"`
	RevokedAt           *time.Time `gorm:"column:revoked_at"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
}

func (IAMRefreshSessionModel) TableName() string {
	return "iam_refresh_sessions"
}

type IAMRefreshSessionRepository struct {
	db *gorm.DB
}

func NewIAMRefreshSessionRepository(db *gorm.DB) *IAMRefreshSessionRepository {
	return &IAMRefreshSessionRepository{db: db}
}

func (r *IAMRefreshSessionRepository) CreateRefreshSession(ctx context.Context, session *domainiam.RefreshSession) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("iam refresh session repository is not ready")
	}
	if session == nil {
		return derrors.ErrInvalidArgument
	}
	if err := r.db.WithContext(ctx).Create(iamRefreshSessionToModel(session)).Error; err != nil {
		if isDuplicate(err) {
			return derrors.ErrConflict
		}
		return fmt.Errorf("create iam refresh session: %w", err)
	}
	return nil
}

func (r *IAMRefreshSessionRepository) RotateRefreshSession(ctx context.Context, currentTokenID string, expectedUserID int64, next *domainiam.RefreshSession, now time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("iam refresh session repository is not ready")
	}
	if next == nil || currentTokenID == "" || expectedUserID <= 0 {
		return derrors.ErrInvalidArgument
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := findRefreshSessionForUpdate(tx, currentTokenID)
		if err != nil {
			return err
		}
		entity, err := iamRefreshSessionFromModel(current)
		if err != nil {
			return err
		}
		if entity.UserID != expectedUserID || !entity.Usable(now) {
			return derrors.ErrUnauthorized
		}
		if err := tx.Create(iamRefreshSessionToModel(next)).Error; err != nil {
			if isDuplicate(err) {
				return derrors.ErrConflict
			}
			return fmt.Errorf("create rotated iam refresh session: %w", err)
		}
		updates := map[string]any{
			"status":                 string(domainiam.RefreshSessionRotated),
			"replaced_by_session_id": next.ID,
			"revoked_at":             now,
			"updated_at":             now,
		}
		result := tx.Model(&IAMRefreshSessionModel{}).
			Where("id = ? AND status = ?", current.ID, string(domainiam.RefreshSessionActive)).
			Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("rotate iam refresh session: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return derrors.ErrUnauthorized
		}
		return nil
	})
}

func (r *IAMRefreshSessionRepository) RevokeRefreshSession(ctx context.Context, refreshTokenID string, now time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("iam refresh session repository is not ready")
	}
	if refreshTokenID == "" {
		return derrors.ErrInvalidArgument
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	result := r.db.WithContext(ctx).Model(&IAMRefreshSessionModel{}).
		Where("refresh_token_id = ? AND status = ?", refreshTokenID, string(domainiam.RefreshSessionActive)).
		Updates(map[string]any{
			"status":     string(domainiam.RefreshSessionRevoked),
			"revoked_at": now,
			"updated_at": now,
		})
	if result.Error != nil {
		return fmt.Errorf("revoke iam refresh session: %w", result.Error)
	}
	return nil
}

func (r *IAMRefreshSessionRepository) RevokeActiveRefreshSessionsByUser(ctx context.Context, userID int64, now time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("iam refresh session repository is not ready")
	}
	if userID <= 0 {
		return derrors.ErrInvalidArgument
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if err := r.db.WithContext(ctx).Model(&IAMRefreshSessionModel{}).
		Where("user_id = ? AND status = ?", userID, string(domainiam.RefreshSessionActive)).
		Updates(map[string]any{
			"status":     string(domainiam.RefreshSessionRevoked),
			"revoked_at": now,
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("revoke user iam refresh sessions: %w", err)
	}
	return nil
}

func findRefreshSessionForUpdate(tx *gorm.DB, refreshTokenID string) (*IAMRefreshSessionModel, error) {
	var model IAMRefreshSessionModel
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("refresh_token_id = ?", refreshTokenID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, derrors.ErrUnauthorized
		}
		return nil, fmt.Errorf("find iam refresh session: %w", err)
	}
	return &model, nil
}

func iamRefreshSessionToModel(entity *domainiam.RefreshSession) *IAMRefreshSessionModel {
	var replacedBy *int64
	if entity.ReplacedBySessionID > 0 {
		replacedBy = &entity.ReplacedBySessionID
	}
	return &IAMRefreshSessionModel{
		ID:                  entity.ID,
		UserID:              entity.UserID,
		RefreshTokenID:      entity.RefreshTokenID,
		Status:              string(entity.Status),
		ReplacedBySessionID: replacedBy,
		ExpiresAt:           entity.ExpiresAt,
		RevokedAt:           entity.RevokedAt,
		CreatedAt:           entity.CreatedAt,
		UpdatedAt:           entity.UpdatedAt,
	}
}

func iamRefreshSessionFromModel(model *IAMRefreshSessionModel) (*domainiam.RefreshSession, error) {
	if model == nil {
		return nil, derrors.ErrNotFound
	}
	replacedBy := int64(0)
	if model.ReplacedBySessionID != nil {
		replacedBy = *model.ReplacedBySessionID
	}
	return &domainiam.RefreshSession{
		ID:                  model.ID,
		UserID:              model.UserID,
		RefreshTokenID:      model.RefreshTokenID,
		Status:              domainiam.RefreshSessionStatus(model.Status),
		ReplacedBySessionID: replacedBy,
		ExpiresAt:           model.ExpiresAt,
		RevokedAt:           model.RevokedAt,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}, nil
}
