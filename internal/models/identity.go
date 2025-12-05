package models

// Identity represents either a user or a group in the system.
// It serves as a unified abstraction for access control subjects,
// allowing policies to reference both users and groups consistently.
type Identity struct {
	// ID is the unique identifier for this identity. This will most likely be an email for users
	// or a group name for groups. This will be used to tie identities across providers.
	ID string `json:"id"`
	// Label is a human-readable name or description for this identity.
	Label string `json:"label"`

	// User contains the user details if this identity represents a user.
	// Will be nil if this identity represents a group.
	User *User `json:"user"`
	// Group contains the group details if this identity represents a group.
	// Will be nil if this identity represents a user.
	Group *Group `json:"group"`

	// The providers this identity is associated with
	// Format is map[provider_name]provider_type
	Providers map[string]string `json:"providers,omitempty"`
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

func (i *Identity) GetProviders() map[string]string {
	return i.Providers
}

func (i *Identity) AddProvider(provider *Provider) {

	if provider == nil {
		return
	}

	// Check if provider already exists
	if i.Providers == nil {
		i.Providers = make(map[string]string)
	}
	if _, exists := i.Providers[provider.Name]; !exists {
		i.Providers[provider.Name] = provider.Provider
	}
}
