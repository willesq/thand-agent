package models

type Group struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (g *Group) GetID() string {
	return g.ID
}

func (g *Group) GetName() string {
	return g.Name
}

func (g *Group) GetEmail() string {
	return g.Email
}
