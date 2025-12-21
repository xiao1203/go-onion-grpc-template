package domain

type User struct {
	ID          int64
	Email       string
	DisplayName string
	PictureURL  string
	Roles       []string
}
