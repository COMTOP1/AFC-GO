package user_test

import (
	"context"
	"testing"

	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/user"
)

// TestEditUserRefusesToStealAnotherUsersEmail reproduces the admin "edit
// user" overwrite bug: EditUser looked up the row to update with
// "WHERE email = $1 OR id = $2", so setting user B's email to an email
// already belonging to user A caused the lookup to return user A instead of
// B, and A's row was silently overwritten with B's edited profile fields.
func TestEditUserRefusesToStealAnotherUsersEmail(t *testing.T) {
	db := openTestDB(t)
	store := user.NewUserRepo(db)
	ctx := context.Background()

	_, err := store.AddUser(ctx, user.User{
		Name:  "Alice",
		Email: "alice@example.test",
		Role:  role.Webmaster,
		Phone: null.StringFrom("111"),
	})
	if err != nil {
		t.Fatalf("failed to add user A: %v", err)
	}
	alice, err := store.GetUser(ctx, user.User{Email: "alice@example.test"})
	if err != nil {
		t.Fatalf("failed to get user A: %v", err)
	}

	_, err = store.AddUser(ctx, user.User{
		Name:  "Bob",
		Email: "bob@example.test",
		Role:  role.Treasurer,
		Phone: null.StringFrom("222"),
	})
	if err != nil {
		t.Fatalf("failed to add user B: %v", err)
	}
	bob, err := store.GetUser(ctx, user.User{Email: "bob@example.test"})
	if err != nil {
		t.Fatalf("failed to get user B: %v", err)
	}

	// Admin attempts to edit Bob, but sets Bob's email to Alice's email
	// (e.g. a typo or a stale form). This must fail rather than silently
	// overwriting Alice's account.
	bobEdit := bob
	bobEdit.Name = "Bob Renamed"
	bobEdit.Email = alice.Email
	bobEdit.Role = role.Chairperson
	bobEdit.Phone = null.StringFrom("333")

	_, err = store.EditUser(ctx, bobEdit)
	if err == nil {
		t.Fatal("expected EditUser to fail when the new email belongs to another user")
	}

	aliceAfter, err := store.GetUser(ctx, user.User{Email: "alice@example.test"})
	if err != nil {
		t.Fatalf("failed to get user A after failed edit: %v", err)
	}
	if aliceAfter.ID != alice.ID {
		t.Errorf("expected user A's id to remain %d, got %d", alice.ID, aliceAfter.ID)
	}
	if aliceAfter.Name != alice.Name {
		t.Errorf("user A's name was overwritten: expected %q, got %q", alice.Name, aliceAfter.Name)
	}
	if aliceAfter.Role != alice.Role {
		t.Errorf("user A's role was overwritten: expected %v, got %v", alice.Role, aliceAfter.Role)
	}
	if aliceAfter.Phone.String != alice.Phone.String {
		t.Errorf("user A's phone was overwritten: expected %q, got %q", alice.Phone.String, aliceAfter.Phone.String)
	}
	if aliceAfter.FileName.String != alice.FileName.String {
		t.Errorf("user A's file_name was overwritten: expected %q, got %q", alice.FileName.String, aliceAfter.FileName.String)
	}
	if aliceAfter.Email != alice.Email {
		t.Errorf("user A's email changed: expected %q, got %q", alice.Email, aliceAfter.Email)
	}
}

// TestEditUserNormalEditSucceeds is the positive counterpart: a normal edit
// with a new name and a new, unique email must still work and must only
// affect the edited user.
func TestEditUserNormalEditSucceeds(t *testing.T) {
	db := openTestDB(t)
	store := user.NewUserRepo(db)
	ctx := context.Background()

	_, err := store.AddUser(ctx, user.User{
		Name:  "Alice",
		Email: "alice2@example.test",
		Role:  role.Webmaster,
	})
	if err != nil {
		t.Fatalf("failed to add user A: %v", err)
	}
	alice, err := store.GetUser(ctx, user.User{Email: "alice2@example.test"})
	if err != nil {
		t.Fatalf("failed to get user A: %v", err)
	}

	_, err = store.AddUser(ctx, user.User{
		Name:  "Bob",
		Email: "bob2@example.test",
		Role:  role.Treasurer,
	})
	if err != nil {
		t.Fatalf("failed to add user B: %v", err)
	}
	bob, err := store.GetUser(ctx, user.User{Email: "bob2@example.test"})
	if err != nil {
		t.Fatalf("failed to get user B: %v", err)
	}

	bobEdit := bob
	bobEdit.Name = "Bob Renamed"
	bobEdit.Email = "bob2-new@example.test"

	_, err = store.EditUser(ctx, bobEdit)
	if err != nil {
		t.Fatalf("expected a normal edit with a unique email to succeed, got: %v", err)
	}

	bobAfter, err := store.GetUser(ctx, user.User{Email: "bob2-new@example.test"})
	if err != nil {
		t.Fatalf("failed to get user B after edit: %v", err)
	}
	if bobAfter.ID != bob.ID {
		t.Errorf("expected user B's id to remain %d, got %d", bob.ID, bobAfter.ID)
	}
	if bobAfter.Name != "Bob Renamed" {
		t.Errorf("expected user B's name to be updated, got %q", bobAfter.Name)
	}

	aliceAfter, err := store.GetUser(ctx, user.User{Email: "alice2@example.test"})
	if err != nil {
		t.Fatalf("failed to get user A after B's edit: %v", err)
	}
	if aliceAfter.ID != alice.ID || aliceAfter.Name != alice.Name {
		t.Errorf("user A must be unaffected by editing user B, got %+v", aliceAfter)
	}
}
