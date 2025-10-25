package models

type Identity struct {
	Name string `json:"name"`

	User  *User  `json:"user"`
	Group *Group `json:"group"`
}

func (i *Identity) GetIdentity() string {
	return i.Name
}

func (i *Identity) GetUser() *User {
	return i.User
}

func (i *Identity) GetGroup() *Group {
	return i.Group
}

func (i *Identity) IsUser() bool {
	return i.User != nil
}

func (i *Identity) IsGroup() bool {
	return i.Group != nil
}
