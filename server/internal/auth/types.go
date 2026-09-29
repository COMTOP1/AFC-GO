package auth

import (
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

// CurrentUser is the logged-in user as the API exposes it.
type CurrentUser struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Phone       string      `json:"phone,omitempty"`
	Role        string      `json:"role"`
	TeamID      int         `json:"teamId,omitempty"`
	ImageURL    string      `json:"imageUrl,omitempty"`
	Permissions Permissions `json:"permissions"`
}

// Permissions tells the client which admin controls to show. The server
// still enforces every rule with guards.
type Permissions struct {
	CanEdit          bool `json:"canEdit"`
	CanManageGallery bool `json:"canManageGallery"`
	CanManageUsers   bool `json:"canManageUsers"`
}

// LoginInput is the body of POST /auth/login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

// LoginResponse is either the logged-in user or a reset instruction.
type LoginResponse struct {
	User          *CurrentUser `json:"user,omitempty"`
	ResetRequired bool         `json:"resetRequired"`
	ResetURL      string       `json:"resetUrl,omitempty"`
}

// PasswordInput is the body of POST /auth/password.
type PasswordInput struct {
	OldPassword          string `json:"oldPassword"`
	NewPassword          string `json:"newPassword"`
	ConfirmationPassword string `json:"confirmationPassword"`
}

// ResetInput is the body of POST /auth/reset/{token}.
type ResetInput struct {
	NewPassword          string `json:"newPassword"`
	ConfirmationPassword string `json:"confirmationPassword"`
}

// NewCurrentUser projects u for the API; it never includes password data.
func NewCurrentUser(u user.User, files *upload.Files) CurrentUser {
	return CurrentUser{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Phone:    u.Phone.String,
		Role:     u.Role.String(),
		TeamID:   u.TeamID,
		ImageURL: files.URL(u.FileName.String),
		Permissions: Permissions{
			CanEdit:          u.Role.CanEdit(),
			CanManageGallery: u.Role.CanManageGallery(),
			CanManageUsers:   u.Role.IsClubSecretaryHigher(),
		},
	}
}
