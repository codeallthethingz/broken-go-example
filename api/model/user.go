package model

// User a user object
type User struct {
	Email       string `json:"email"`
	AccessToken string `json:"-"`
}
