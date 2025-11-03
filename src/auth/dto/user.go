package dto

type UserCreatedEvent struct {
	IDUser int64  `json:"id_user"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

type UpdateAuthEmailDTO struct {
	NewEmail string `json:"newEmail" binding:"required,email" validate:"required"`
}
