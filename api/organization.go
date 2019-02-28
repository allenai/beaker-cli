package api

import "time"

// OrganizationSpec defines creation properties for an organization.
type OrganizationSpec struct {
	// (required) The organization's name. This name may not collide with an
	// existing organization or user.
	Name string `json:"name"`

	// (required) Existing user who will administer this organization.
	Owner string `json:"owner"`

	// (optional) A human-friendly name for display.
	DisplayName string `json:"displayName,omitempty"`

	// (optional) A brief description of the organization.
	Description string `json:"description,omitempty"`
}

// Organization describes an organization, or group of users.
type Organization struct {
	Identity
	Created     time.Time `json:"created"`
	Description string    `json:"description,omitempty"`
}

// OrganizationPage is a page of results from a batch organization API.
type OrganizationPage struct {
	// Results of a batch query.
	Data []Organization `json:"data"`

	// Opaque token to the element after Data, provided only if more data is available.
	NextCursor string `json:"nextCursor,omitempty"`
}

// OrgMembership describe's a user's membership within an organization.
type OrgMembership struct {
	// Role of the user within the org. Values may be "admin" or "member".
	Role string `json:"role"`

	// Organization in which the user is a member.
	Organization Organization `json:"organization"`

	// User holding a membership in the org.
	User UserDetail `json:"user"`
}
