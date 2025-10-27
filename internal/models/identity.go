package models

type Identity struct {
	ID    string `json:"id"`
	Label string `json:"label"`

	User  *User  `json:"user"`
	Group *Group `json:"group"`
}

func (i *Identity) GetId() string {
	return i.ID
}

func (i *Identity) GetLabel() string {
	return i.Label
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
