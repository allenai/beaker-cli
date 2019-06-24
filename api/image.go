package api

import (
	"path"
	"time"
)

// CreateImageResponse is a service response returned when a new image is
// created. For now it's just the image ID, but may be expanded in the future.
type CreateImageResponse struct {
	ID string `json:"id"`
}

// Image describes the Docker image ran by a Task while executing an Experiment.
type Image struct {
	// Identity
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`

	// Ownership
	Owner  Identity `json:"owner"`
	Author Identity `json:"author"`
	User   Identity `json:"user"` // TODO: Deprecated.

	// Status
	Created   time.Time `json:"created"`
	Committed time.Time `json:"committed,omitempty"`

	// Original image tag, if supplied on creation. See ImageSpec.
	OriginalTagDeprecated string `json:"original_tag,omitempty"`
	OriginalTag           string `json:"originalTag,omitempty"`

	// A plain-text description of this image.
	Description string `json:"description,omitempty"`
}

// DisplayID returns the most human-friendly name available for an image
// while guaranteeing that it's unique and non-empty.
func (b *Image) DisplayID() string {
	if b.Name != "" {
		return path.Join(b.User.Name, b.Name)
	}
	return b.ID
}

// ImageSpec is a specification for creating a new Image.
type ImageSpec struct {
	// (optional) Organization on behalf of whom this resource is created. The
	// user issuing the request must be a member of the organization. If omitted,
	// the resource will be owned by the requestor.
	Organization string `json:"org,omitempty"`

	// (required) Unique identifier for the image's image. In Docker images,
	// this is a SHA256 hash.
	ImageID string `json:"ImageID"` // TODO: convert to loweCase name and update reflection tag

	// (optional) Text description for the image.
	Description string `json:"Description,omitempty"` // TODO: convert to loweCase name and update reflection tag

	// (optional) Original image tag from which the image was created.
	ImageTag string `json:"ImageTag,omitempty"` // TODO: convert to loweCase name and update reflection tag

	// (optional) A token representing the user to which the object should be attributed.
	// If omitted attribution will be given to the user issuing the request.
	AuthorTokenDeprecated string `json:"author_token,omitempty"`
	AuthorToken           string `json:"authorToken,omitempty"`
}

// ImagePatchSpec describes a patch to apply to an image's editable fields.
// Only one field may be set in a single request.
type ImagePatchSpec struct {
	// (optional) Unqualified name to assign to the image. It is considered
	// a collision error if another image has the same creator and name.
	Name *string `json:"name,omitempty"`

	// (optional) Description to assign to the image or empty string to
	// delete an existing description.
	Description *string `json:"description,omitempty"`

	// (optional) Whether the image should be committed. Ignored if false.
	// When committed, an image is placed in Beaker's internal registry and
	// further attempts to push its image will be ignored.
	Commit bool `json:"commit,omitempty"`
}

// ImageRepository contains a repository/tag and credentials required to
// upload an image's Docker image via "docker push".
type ImageRepository struct {
	// Full tag, including registry, expected by Beaker. Clients must push this
	// tag exactly for Beaker to recognize the image.
	ImageTagDeprecated string `json:"image_tag"`
	ImageTag           string `json:"imageTag"`

	// Credentials for the image's registry.
	Auth RegistryAuth `json:"auth"`
}

// RegistryAuth supplies authorization for a Docker registry.
type RegistryAuth struct {
	ServerAddress string `json:"server_address"`
	User          string `json:"user"`
	Password      string `json:"password"`
}
