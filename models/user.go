package models

type User struct {
	ID        int     `json:"id"`
	Username  string  `json:"username"`
	Password  string  `json:"password,omitempty"`
	FirstName string  `json:"firstName,omitempty"`
	LastName  string  `json:"lastName,omitempty"`
	Age       int     `json:"age,omitempty"`
	Height    float64 `json:"height"`
	Gender    string  `json:"gender,omitempty"`
	Email     string  `json:"email,omitempty"`
	Target    float64 `json:"target,omitempty"`
	IsVisible bool    `json:"isVisible,omitempty"`
}

type UserProfile struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Height   float64 `json:"height"`
	Progress float64 `json:"progress"`
}
