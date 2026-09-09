package repositories

import (
	"errors"
	"fmt"

	"github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	vo "github.com/CaiqueGOliveira/TelemetryGo/src/domain/valueobjects"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ repository.UserRepository = (*PostgresUserRepository)(nil)

type PostgresUserRepository struct {
	db *gorm.DB
}

type userRecord struct {
	ID     string `gorm:"primaryKey"`
	Email  string `gorm:"uniqueIndex"`
	Name   string
	Hash   string
	ApiKey string `gorm:"uniqueIndex"`
}

func NewPostgresUserRepository(db *gorm.DB) *PostgresUserRepository {
	db.AutoMigrate(&userRecord{})

	return &PostgresUserRepository{db: db}
}

func (repo *PostgresUserRepository) Save(user *domain.User) error {
	if user == nil {
		return errors.New("user is nil")
	}

	record := &userRecord{
		ID:     user.Id.String(),
		Email:  user.Email.Text(),
		Name:   user.Name,
		Hash:   user.Hash.GetPasswordHash(),
		ApiKey: user.ApiKey,
	}

	return repo.db.Create(record).Error
}

func (repo *PostgresUserRepository) FindByEmail(email string) (*domain.User, error) {
	var record userRecord
	if err := repo.db.Where("email = ?", email).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user doesn't exist")
		}
		return nil, err
	}

	return recordToUser(&record)
}

func (repo *PostgresUserRepository) FindByApiKey(apiKey string) (*domain.User, error) {
	var record userRecord
	if err := repo.db.Where("api_key = ?", apiKey).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with this api key doesn't exist")
		}
		return nil, err
	}

	return recordToUser(&record)
}

func recordToUser(record *userRecord) (*domain.User, error) {
	id, err := uuid.Parse(record.ID)
	if err != nil {
		return nil, err
	}

	email, err := vo.CreateEmail(record.Email)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		Id:     id,
		Email:  email,
		Name:   record.Name,
		Hash:   vo.NewPasswordHashedFromHash(record.Hash),
		ApiKey: record.ApiKey,
	}, nil
}