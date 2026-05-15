package mysql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	driver "github.com/go-sql-driver/mysql"
	derrors "github.com/rin721/keiyaku-go/internal/domain/errors"
	domainuser "github.com/rin721/keiyaku-go/internal/domain/user"
	"gorm.io/gorm"
)

type UserModel struct {
	ID           int64     `gorm:"column:id;primaryKey"`
	Username     string    `gorm:"column:username"`
	Email        string    `gorm:"column:email"`
	PasswordHash string    `gorm:"column:password_hash"`
	DisplayName  string    `gorm:"column:display_name"`
	Status       string    `gorm:"column:status"`
	RolesJSON    string    `gorm:"column:roles_json"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (UserModel) TableName() string {
	return "users"
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, entity *domainuser.User) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("user repository is not ready")
	}
	model, err := userToModel(entity)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if isDuplicate(err) {
			return derrors.ErrConflict
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domainuser.User, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("user repository is not ready")
	}
	var model UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if IsNotFound(err) {
			return nil, derrors.ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return userFromModel(&model)
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domainuser.User, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("user repository is not ready")
	}
	var model UserModel
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&model).Error; err != nil {
		if IsNotFound(err) {
			return nil, derrors.ErrNotFound
		}
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	return userFromModel(&model)
}

func userToModel(entity *domainuser.User) (*UserModel, error) {
	if entity == nil {
		return nil, derrors.ErrInvalidArgument
	}
	roles, err := json.Marshal(entity.Roles)
	if err != nil {
		return nil, fmt.Errorf("marshal user roles: %w", err)
	}
	return &UserModel{
		ID:           entity.ID,
		Username:     entity.Username,
		Email:        entity.Email,
		PasswordHash: entity.PasswordHash,
		DisplayName:  entity.DisplayName,
		Status:       string(entity.Status),
		RolesJSON:    string(roles),
		CreatedAt:    entity.CreatedAt,
		UpdatedAt:    entity.UpdatedAt,
	}, nil
}

func userFromModel(model *UserModel) (*domainuser.User, error) {
	if model == nil {
		return nil, derrors.ErrNotFound
	}
	var roles []string
	if model.RolesJSON != "" {
		if err := json.Unmarshal([]byte(model.RolesJSON), &roles); err != nil {
			return nil, fmt.Errorf("unmarshal user roles: %w", err)
		}
	}
	return &domainuser.User{
		ID:           model.ID,
		Username:     model.Username,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		DisplayName:  model.DisplayName,
		Status:       domainuser.Status(model.Status),
		Roles:        roles,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}, nil
}

func isDuplicate(err error) bool {
	var mysqlErr *driver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
