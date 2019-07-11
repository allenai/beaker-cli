package image

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/beaker/client/api"
	beaker "github.com/beaker/client/client"

	"github.com/allenai/beaker/config"
)

// CreateOptions wraps options used to create an image.
type CreateOptions struct {
	Description string
	Name        string
	Quiet       bool
}

func newCreateCmd(
	parent *kingpin.CmdClause,
	parentOpts *CmdOptions,
	config *config.Config,
) {
	opts := &CreateOptions{}
	image := new(string)

	cmd := parent.Command("create", "Create a new image")
	cmd.Flag("desc", "Assign a description to the image").StringVar(&opts.Description)
	cmd.Flag("name", "Assign a name to the image").Short('n').StringVar(&opts.Name)
	cmd.Flag("quiet", "Only display created image's ID").Short('q').BoolVar(&opts.Quiet)
	cmd.Arg("image", "Docker image ID").Required().StringVar(image)

	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.Addr, config.UserToken)
		if err != nil {
			return err
		}
		_, err = Create(context.TODO(), os.Stdout, beaker, *image, opts)
		return err
	})
}

// Create creates a new image and returns its ID.
func Create(
	ctx context.Context,
	w io.Writer,
	beaker *beaker.Client,
	imageTag string,
	opts *CreateOptions,
) (string, error) {
	if w == nil {
		w = ioutil.Discard
	}
	if opts == nil {
		opts = &CreateOptions{}
	}

	docker, err := docker.NewEnvClient()
	if err != nil {
		return "", errors.Wrap(err, "failed to create Docker client")
	}

	dockerImage, _, err := docker.ImageInspectWithRaw(ctx, imageTag)
	if err != nil {
		return "", errors.WithStack(err)
	}

	spec := api.ImageSpec{
		Description: opts.Description,
		ImageID:     dockerImage.ID,
		ImageTag:    imageTag,
	}
	image, err := beaker.CreateImage(ctx, spec, opts.Name)
	if err != nil {
		return "", err
	}

	if !opts.Quiet {
		if opts.Name == "" {
			fmt.Fprintf(w, "Pushing %s as %s ...\n", imageTag, color.BlueString(image.ID()))
		} else {
			fmt.Fprintf(w, "Pushing %s as %s (%s)...\n", imageTag, color.BlueString(opts.Name), image.ID())
		}
	}

	repo, err := image.Repository(ctx, true)
	if err != nil {
		return "", errors.WithMessage(err, "failed to retrieve credentials for remote repository")
	}

	// Tag the image to the remote repository.
	if err := docker.ImageTag(ctx, imageTag, repo.ImageTag); err != nil {
		return "", errors.Wrap(err, "failed to set remote image tag")
	}
	defer func() {
		// We ignore the error here intentionally. Cleaning up is best-effort
		// and we can't do anything to recover if this fails.
		_, _ = docker.ImageRemove(ctx, repo.ImageTag, types.ImageRemoveOptions{})
	}()

	authConfig := types.AuthConfig{
		ServerAddress: repo.Auth.ServerAddress,
		Username:      repo.Auth.User,
		Password:      repo.Auth.Password,
	}
	authJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode remote repository auth")
	}
	authStr := base64.URLEncoding.EncodeToString(authJSON)

	r, err := docker.ImagePush(ctx, repo.ImageTag, types.ImagePushOptions{RegistryAuth: authStr})
	if err != nil {
		return "", errors.WithStack(err)
	}
	// Display push responses as the Docker CLI would. This also translates remote errors.
	var stream io.Writer = os.Stdout
	if opts.Quiet {
		stream = ioutil.Discard
	}
	if err := jsonmessage.DisplayJSONMessagesStream(r, stream, 0, false, nil); err != nil {
		_ = r.Close()
		return "", errors.WithStack(err)
	}
	if err := r.Close(); err != nil {
		return "", errors.WithStack(err)
	}

	if err := image.Commit(ctx); err != nil {
		return "", errors.WithMessage(err, "failed to commit image")
	}

	if opts.Quiet {
		fmt.Fprintln(w, image.ID())
	} else {
		fmt.Fprintln(w, "Done.")
	}
	return image.ID(), nil
}
