package models

type ProviderRolesResponse struct {
	Version  string         `json:"version"`
	Provider string         `json:"provider"`
	Roles    []ProviderRole `json:"roles"`
}

type ProviderRole struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	// Store the underlying provider-specific role object if needed
	Role any `json:"-"`
}
