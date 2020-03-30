package image

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	beaker "github.com/beaker/client/client"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/allenai/beaker/config"
)

// PullOptions defines settings for image pull command
// TODO: make PullOptions and fields unexported once not needed by blueprint command
type PullOptions struct {
	Image string
	Tag   string
	Quiet bool
}

func newPullCmd(
	parent *kingpin.CmdClause,
	parentOpts *CmdOptions,
	config *config.Config,
) {
	o := &PullOptions{}
	cmd := parent.Command("pull", "Pull the image's cooresponding Docker image")
	cmd.Flag("quiet", "Only display the pulled image's tag").Short('q').BoolVar(&o.Quiet)
	cmd.Arg("image", "Image name or ID").Required().StringVar(&o.Image)
	cmd.Arg("tag", "Name and optional tag in the 'name:tag' format").StringVar(&o.Tag)

	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.Addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.Run(beaker)
	})
}

// Run executes image pull command
// TODO: make Run unexported once not needed by blueprint command
func (o *PullOptions) Run(beaker *beaker.Client) error {
	ctx := context.TODO()
	docker, err := docker.NewEnvClient()
	if err != nil {
		return errors.Wrap(err, "failed to create Docker client")
	}

	image, err := beaker.Image(ctx, o.Image)
	if err != nil {
		return err
	}

	repo, err := image.Repository(ctx, false)
	if err != nil {
		return errors.WithMessage(err, "failed to retrieve credentials for remote repository")
	}

	if !o.Quiet {
		fmt.Printf("Pulling %s ...\n", repo.ImageTag)
	}

	authConfig := types.AuthConfig{
		ServerAddress: repo.Auth.ServerAddress,
		Username:      repo.Auth.User,
		Password:      repo.Auth.Password,
	}
	authJSON, err := json.Marshal(authConfig)
	if err != nil {
		return errors.Wrap(err, "failed to encode remote repository auth")
	}
	authStr := base64.URLEncoding.EncodeToString(authJSON)

	r, err := docker.ImagePull(ctx, repo.ImageTag, types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		return errors.WithStack(err)
	}
	defer r.Close()

	// Display push responses as the Docker CLI would. This also translates remote errors.
	var stream io.Writer = os.Stdout
	if o.Quiet {
		stream = ioutil.Discard
	}
	if err := jsonmessage.DisplayJSONMessagesStream(r, stream, 0, false, nil); err != nil {
		return errors.WithStack(err)
	}

	tag := repo.ImageTag
	if o.Tag != "" {
		if !o.Quiet {
			// We intentionally print the un-mangled tag.
			fmt.Printf("Renaming %s to %s ...\n", repo.ImageTag, o.Tag)
		}

		// We must normalize or ImageTag will return an error on otherwise valid references.
		normalized, err := reference.ParseNormalizedNamed(o.Tag)
		if err != nil {
			return errors.Wrap(err, "invalid target name")
		}
		if err := docker.ImageTag(ctx, repo.ImageTag, normalized.String()); err != nil {
			return errors.Wrap(err, "failed to tag image")
		}

		// We ignore the error here intentionally. Cleaning up is best-effort
		// and we can't do anything to recover if this fails.
		_, _ = docker.ImageRemove(ctx, repo.ImageTag, types.ImageRemoveOptions{})
		tag = normalized.String()
	}

	if o.Quiet {
		fmt.Println(tag)
	} else {
		fmt.Println("Done.")
	}
	return nil
}
