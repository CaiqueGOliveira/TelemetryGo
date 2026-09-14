package repositories

import (
	"testing"

	domain "github.com/CaiqueGOliveira/TelemetryGo/src/domain"
	"github.com/CaiqueGOliveira/TelemetryGo/src/domain/repository"
	"github.com/google/uuid"
)

func TestUserRepositorySaveAndFind(t *testing.T) {
	repo := NewUserRepository()

	user, err := domain.CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(user); err != nil {
		t.Fatal(err)
	}

	found, err := repo.FindByEmail("user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Id != user.Id {
		t.Errorf("expected id to match, got %s != %s", found.Id, user.Id)
	}

	byKey, err := repo.FindByApiKey(user.ApiKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if byKey.Id != user.Id {
		t.Errorf("expected api key find to match, got %s != %s", byKey.Id, user.Id)
	}

	byID, err := repo.FindById(user.Id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if byID.Id != user.Id {
		t.Errorf("expected id find to match")
	}
}

func TestUserRepositorySaveDuplicateEmail(t *testing.T) {
	repo := NewUserRepository()

	first, _ := domain.CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err := repo.Save(first); err != nil {
		t.Fatal(err)
	}

	second, _ := domain.CreateUser("user@example.com", "Outro", "Senha#Segura1")
	if err := repo.Save(second); err != repository.ErrDuplicateEmail {
		t.Fatalf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestUserRepositoryUpdate(t *testing.T) {
	repo := NewUserRepository()

	user, _ := domain.CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err := repo.Save(user); err != nil {
		t.Fatal(err)
	}

	if err := user.UpdateProfile("Novo Nome", "novo@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(user); err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	found, err := repo.FindByEmail("novo@example.com")
	if err != nil {
		t.Fatalf("expected user with updated email, got: %v", err)
	}
	if found.Name != "Novo Nome" {
		t.Errorf("expected updated name, got %s", found.Name)
	}
}

func TestUserRepositoryUpdateDuplicateEmail(t *testing.T) {
	repo := NewUserRepository()

	alice, _ := domain.CreateUser("alice@example.com", "Alice", "Senha#Segura1")
	if err := repo.Save(alice); err != nil {
		t.Fatal(err)
	}

	bob, _ := domain.CreateUser("bob@example.com", "Bob", "Senha#Segura1")
	if err := repo.Save(bob); err != nil {
		t.Fatal(err)
	}

	// fetch alice through the repo (same pointer is returned and mutated)
	alice, err := repo.FindByEmail("alice@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if err := alice.UpdateProfile("Alice", "bob@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(alice); err != repository.ErrDuplicateEmail {
		t.Fatalf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestUserRepositoryUpdateNotFound(t *testing.T) {
	repo := NewUserRepository()

	user, _ := domain.CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err := repo.Update(user); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserRepositoryDelete(t *testing.T) {
	repo := NewUserRepository()

	user, _ := domain.CreateUser("user@example.com", "Caique", "Senha#Segura1")
	if err := repo.Save(user); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete(user.Id); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	if _, err := repo.FindById(user.Id); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}

	if err := repo.Delete(user.Id); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound on second delete, got %v", err)
	}
}

func TestUserRepositoryFindMissing(t *testing.T) {
	repo := NewUserRepository()

	if _, err := repo.FindByEmail("missing@example.com"); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound on email, got %v", err)
	}
	if _, err := repo.FindByApiKey("tg_missing"); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound on api key, got %v", err)
	}
	if _, err := repo.FindById(uuid.New()); err != repository.ErrNotFound {
		t.Fatalf("expected ErrNotFound on id, got %v", err)
	}
}
