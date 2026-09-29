package user

// Admin is a user as the user-management API returns it. It never contains
// password material.
type Admin struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone,omitempty"`
	Role     string `json:"role"`     // display name, e.g. "Club Secretary"
	RoleCode string `json:"roleCode"` // input code, e.g. "club_secretary"
	TeamID   int    `json:"teamId,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
}

// CreateInput is a new user. Role is a role code (e.g. "manager"); TeamID
// is kept only for managers.
type CreateInput struct {
	Name   string
	Email  string
	Phone  string
	Role   string
	TeamID int
}

// UpdateInput changes a user; nil fields are left unchanged.
type UpdateInput struct {
	Name        *string
	Email       *string
	Phone       *string
	Role        *string
	TeamID      *int
	RemoveImage bool
}

// Created is the result of creating a user. TempPassword is set only when
// the signup email could not be sent, so the admin can pass it on.
type Created struct {
	User         Admin  `json:"user"`
	EmailSent    bool   `json:"emailSent"`
	TempPassword string `json:"tempPassword,omitempty"`
}

// ResetResult is the result of an admin-triggered password reset. ResetURL
// is set only when the email could not be sent.
type ResetResult struct {
	EmailSent bool   `json:"emailSent"`
	ResetURL  string `json:"resetUrl,omitempty"`
}
