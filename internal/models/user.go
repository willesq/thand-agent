package models

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/thand-io/agent/internal/common"
)

type User struct {
	ID       string   `json:"id,omitempty"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Verified *bool    `json:"verified,omitempty"`
	Source   string   `json:"source,omitempty"`
	Groups   []string `json:"groups,omitempty"`
}

func (u *User) GetName() string {
	if len(u.Name) > 0 {
		return u.Name
	} else if len(u.Username) > 0 {
		return u.Username
	} else if len(u.Email) > 0 {
		return u.Email
	}
	return "Unknown"
}

func (u *User) GetUsername() string {
	if len(u.Username) > 0 {
		return u.Username
	} else if len(u.Name) > 0 {
		return common.ConvertToSnakeCase(u.Name)
	} else if len(u.Email) > 0 {
		// Use the part before @ in email
		return common.ConvertToSnakeCase(u.Email[:strings.Index(u.Email, "@")])
	} else if len(u.ID) > 0 {
		return common.ConvertToSnakeCase(u.ID)
	}

	return ""
}

func (u *User) GetIdentity() string {
	if len(u.Email) > 0 {
		return u.Email
	} else if len(u.Username) > 0 {
		return u.Username
	} else if len(u.ID) > 0 {
		return u.ID
	}
	return common.ConvertToSnakeCase(u.Name)
}

func (u *User) GetFirstName() string {
	if len(u.Name) > 0 {
		parts := strings.Split(u.Name, " ")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return ""
}

func (u *User) GetLastName() string {
	if len(u.Name) > 0 {
		parts := strings.Split(u.Name, " ")
		if len(parts) > 1 {
			return parts[len(parts)-1]
		}
	}
	return ""
}

func (u *User) AsMap() map[string]any {
	// Convert User struct to a map[string]any
	mapUser, err := common.ConvertInterfaceToMap(u)
	if err != nil {

		logrus.WithError(err).Error("Failed to convert User struct to map")
		return nil

	}
	return mapUser
}

type AuthorizeUser struct {
	Scopes      []string `json:"scopes"`
	State       string   `json:"state"`
	RedirectUri string   `json:"redirect_uri"`
	Code        string   `json:"code"`
}
