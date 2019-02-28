package api

// Identity is summary of a user's identity.
type Identity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// UserDetail is a user's full informationYou'l.
type UserDetail struct {
	Identity
	Institution string `json:"institution,omitempty"`
	Role        string `json:"role,omitempty"`
	Email       string `json:"email,omitempty"`
}

// UserPage is a page of results from a batch user API.
type UserPage struct {
	// Results of a batch query.
	Data []UserDetail `json:"data"`

	// Opaque token to the element after Data, provided only if more data is available.
	NextCursor string `json:"nextCursor,omitempty"`
}

// UserPatchSpec describes a patch to apply to a user's editable fields.
type UserPatchSpec struct {
	// (optional) User account name, used as identifier and for scoping object references.
	// Resource name validation rules apply.
	Name *string `json:"name,omitempty"`

	// (optional) Name to display when showing the user. Unlike a user account
	// name, display names have no restrictions character set or uniqueness.
	DisplayName *string `json:"display_name,omitempty"`

	// (optional) User-submitted professional affiliation.
	Institution *string `json:"institution,omitempty"`

	// (optional) Assign an authorization level to the user.
	Role *string `json:"role,omitempty"`
}
