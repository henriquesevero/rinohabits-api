package dto

type RegisterRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Timezone   string `json:"timezone"`
	InviteCode string `json:"invite_code"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type UserResponse struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	Email                 string   `json:"email"`
	AvatarURL             *string  `json:"avatar_url"`
	BookCollectionOrder   []string `json:"book_collection_order"`
	CourseCollectionOrder []string `json:"course_collection_order"`
}

type UpdateCollectionOrderRequest struct {
	Order []string `json:"order"`
}

type ChangeEmailRequest struct {
	CurrentPassword string `json:"current_password"`
	NewEmail        string `json:"new_email"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type DeleteAccountRequest struct {
	CurrentPassword string `json:"current_password"`
}
