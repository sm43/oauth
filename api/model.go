package oauth

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Name     string
	GithubID string
}
