package api

import (
	"time"
)

// User contains details about a user in Beaker.
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Institution string `json:"institution,omitempty"`
	Role        string `json:"role,omitempty"`
	Email       string `json:"email,omitempty"`

	// Deprecated: snake_case version(s) of above
	OldDN string `json:"display_name,omitempty"`
}

// UserPage is a page of results from a batch user API.
type UserPage struct {
	// Results of a batch query.
	Data []User `json:"data"`

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

// UserStats describes usage metrics attached to a particular user.
type UserStats struct {
	User          User                 `json:"user"`
	TotalStarted  int64                `json:"total_started"`
	WeeklyStarted []ExperimentsStarted `json:"weekly_started"`
}

// ExperimentsStarted describes how many experiments were started on a given day.
// Intended to be used in an aggregate statistic reporting.
type ExperimentsStarted struct {
	Date    time.Time `json:"date"`
	Started int64     `json:"started"`
}

// UserComputeTime summarizes a user's computational usage over time.
type UserComputeTime struct {
	User              User          `json:"user"`
	WeeklyComputeTime []ComputeTime `json:"weekly_compute_time"`
}

// ComputeTime describes a unit of computation time; intended for usage summaries.
type ComputeTime struct {
	Date         time.Time `json:"date"`
	Milliseconds int64     `json:"milliseconds"`
	GPU          bool      `json:"gpu"`
}
