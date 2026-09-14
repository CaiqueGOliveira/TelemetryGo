package repositories

import (
	"testing"

	domain "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newPostgresUserRepo(t *testing.T) (*postgresUserFixture, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to init sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	repo := &PostgresUserRepository{db: gormDB}
	return &postgresUserFixture{repo: repo, sqlDB: sqlDB}, mock
}

type postgresUserFixture struct {
	repo  *PostgresUserRepository
	sqlDB interface {
		Close() error
	}
}

func TestPostgresUserRepositorySaveSuccess(t *testing.T) {
	fixture, mock := newPostgresUserRepo(t)
	defer fixture.sqlDB.Close()

	user, err := domain.CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "user_records"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := fixture.repo.Save(user); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresUserRepositorySaveDuplicate(t *testing.T) {
	fixture, mock := newPostgresUserRepo(t)
	defer fixture.sqlDB.Close()

	user, err := domain.CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "user_records"`).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectRollback()

	if err := fixture.repo.Save(user); err != repository.ErrDuplicateEmail {
		t.Fatalf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestPostgresUserRepositoryUpdateSuccess(t *testing.T) {
	fixture, mock := newPostgresUserRepo(t)
	defer fixture.sqlDB.Close()

	user, err := domain.CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err != nil {
		t.Fatal(err)
	}
	if err := user.UpdateProfile("Novo Nome", "novo@example.com"); err != nil {
		t.Fatal(err)
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_records" SET`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := fixture.repo.Update(user); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresUserRepositoryFindByEmail(t *testing.T) {
	fixture, mock := newPostgresUserRepo(t)
	defer fixture.sqlDB.Close()

	userID := uuid.NewString()
	hash, err := domain.CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err != nil {
		t.Fatal(err)
	}

	rows := sqlmock.NewRows([]string{"id", "email", "name", "hash", "api_key"}).
		AddRow(userID, "user@example.com", "Caique", hash.Hash.GetPasswordHash(), "tg_abcdef")
	mock.ExpectQuery(`SELECT \* FROM "user_records" WHERE email = \$1`).
		WillReturnRows(rows)

	found, err := fixture.repo.FindByEmail("user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Id.String() != userID {
		t.Errorf("expected id %s, got %s", userID, found.Id)
	}
	if found.Email.Text() != "user@example.com" {
		t.Errorf("unexpected email %s", found.Email.Text())
	}
}

func TestPostgresUserRepositoryFindByEmailNotFound(t *testing.T) {
	fixture, mock := newPostgresUserRepo(t)
	defer fixture.sqlDB.Close()

	rows := sqlmock.NewRows([]string{"id", "email", "name", "hash", "api_key"})
	mock.ExpectQuery(`SELECT \* FROM "user_records" WHERE email = \$1`).
		WillReturnRows(rows)

	if _, err := fixture.repo.FindByEmail("missing@example.com"); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgresUserRepositoryFindByApiKey(t *testing.T) {
	fixture, mock := newPostgresUserRepo(t)
	defer fixture.sqlDB.Close()

	rows := sqlmock.NewRows([]string{"id", "email", "name", "hash", "api_key"}).
		AddRow(uuid.NewString(), "user@example.com", "Caique", "hash", "tg_abcdef")
	mock.ExpectQuery(`SELECT \* FROM "user_records" WHERE api_key = \$1`).
		WillReturnRows(rows)

	found, err := fixture.repo.FindByApiKey("tg_abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.ApiKey != "tg_abcdef" {
		t.Errorf("unexpected api key %s", found.ApiKey)
	}
}

func TestPostgresUserRepositoryDeleteSuccess(t *testing.T) {
	fixture, mock := newPostgresUserRepo(t)
	defer fixture.sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "user_records" WHERE id = \$1`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := fixture.repo.Delete(uuid.New()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresUserRepositoryDeleteNotFound(t *testing.T) {
	fixture, mock := newPostgresUserRepo(t)
	defer fixture.sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "user_records" WHERE id = \$1`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	if err := fixture.repo.Delete(uuid.New()); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
