package service

import (
	"github.com/brook/server/web/api"
	"github.com/brook/server/web/db"
)

// GetToken retrieves an authentication token from the database
// It returns the token as a string or an empty string if an error occurs
func GetToken() string {
	get, err := db.Get[api.AuthInfo](api.AuthKey)
	if err != nil {
		return ""
	}
	return get.Token
}
