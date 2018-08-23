package api

import (
	"path"
	"time"
)

// CreateBlueprintResponse is a service response returned when a new blueprint is
// created. For now it's just the blueprint ID, but may be expanded in the future.
type CreateBlueprintResponse struct {
	ID string `json:"id"`
}

// Blueprint is a file or collection of files. It may be the result of a task or
// uploaded directly by a user.
type Blueprint struct {
	// The unique ID of the blueprint.
	ID   string `json:"id"`
	User User   `json:"user"`
	Name string `json:"name,omitempty"`

	// Status
	Created   time.Time `json:"created"`
	Committed time.Time `json:"committed,omitempty"`

	// Original image tag, if supplied on creation. See BlueprintSpec.
	OriginalTag string `json:"original_tag,omitempty"`

	// A plain-text description of this blueprint.
	Description string `json:"description,omitempty"`
}

// DisplayID returns the most human-friendly name available for a blueprint
// while guaranteeing that it's unique and non-empty.
func (b *Blueprint) DisplayID() string {
	if b.Name != "" {
		return path.Join(b.User.Name, b.Name)
	}
	return b.ID
}

// BlueprintSpec is a specification for creating a new Blueprint.
type BlueprintSpec struct {
	// (required) Unique identifier for the blueprint's image. In Docker images,
	// this is a SHA256 hash.
	ImageID string `json:"ImageID"` // TODO: convert to loweCase name and update reflection tag

	// (optional) Text description for the blueprint.
	Description string `json:"Description,omitempty"` // TODO: convert to loweCase name and update reflection tag

	// (optional) Original image tag from which the blueprint was created.
	ImageTag string `json:"ImageTag,omitempty"` // TODO: convert to loweCase name and update reflection tag

	// (optional) A token representing the user to which the object should be attributed.
	// If omitted attribution will be given to the user issuing request.
	AuthorToken string `json:"author_token,omitempty"`
}

// BlueprintPatchSpec describes a patch to apply to a blueprint's editable fields.
// Only one field may be set in a single request.
type BlueprintPatchSpec struct {
	// (optional) Unqualified name to assign to the blueprint. It is considered
	// a collision error if another blueprint has the same creator and name.
	Name *string `json:"name,omitempty"`

	// (optional) Description to assign to the blueprint or empty string to
	// delete an existing description.
	Description *string `json:"description,omitempty"`

	// (optional) Whether the blueprint should be committed. Ignored if false.
	// When committed, a blueprint is placed in Beaker's internal registry and
	// further attempts to push its image will be ignored.
	Commit bool `json:"commit,omitempty"`
}

// BlueprintRepository contains a repository/tag and credentials required to
// upload a blueprint's image via "docker push".
type BlueprintRepository struct {
	// Full tag, including registry, expected by Beaker. Clients must push this
	// tag exactly for Beaker to recognize the image.
	ImageTag string `json:"image_tag"`

	// Credentials for the image's registry.
	Auth RegistryAuth `json:"auth"`
}

// RegistryAuth supplies authorization for a Docker registry.
type RegistryAuth struct {
	ServerAddress string `json:"server_address"`
	User          string `json:"user"`
	Password      string `json:"password"`
}
