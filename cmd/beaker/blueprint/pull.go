package blueprint

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
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	beaker "github.com/allenai/beaker/client"
	"github.com/allenai/beaker/config"
)

type pullOptions struct {
	blueprint string
	tag       string
	quiet     bool
}

func newPullCmd(
	parent *kingpin.CmdClause,
	parentOpts *blueprintOptions,
	config *config.Config,
) {
	o := &pullOptions{}
	cmd := parent.Command("pull", "Pull the blueprint's Docker image")
	cmd.Flag("quiet", "Only display the pulled image's tag").Short('q').BoolVar(&o.quiet)
	cmd.Arg("blueprint", "Blueprint name or ID").Required().StringVar(&o.blueprint)
	cmd.Arg("tag", "Name and optional tag in the 'name:tag' format").StringVar(&o.tag)

	cmd.Action(func(c *kingpin.ParseContext) error {
		beaker, err := beaker.NewClient(parentOpts.addr, config.UserToken)
		if err != nil {
			return err
		}
		return o.run(beaker)
	})
}

func (o *pullOptions) run(beaker *beaker.Client) error {
	ctx := context.TODO()
	docker, err := docker.NewEnvClient()
	if err != nil {
		return errors.Wrap(err, "failed to create Docker client")
	}

	blueprint, err := beaker.Blueprint(ctx, o.blueprint)
	if err != nil {
		return err
	}

	repo, err := blueprint.Repository(ctx, false)
	if err != nil {
		return errors.WithMessage(err, "failed to retrieve credentials for remote repository")
	}

	if !o.quiet {
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
	if o.quiet {
		stream = ioutil.Discard
	}
	if err := jsonmessage.DisplayJSONMessagesStream(r, stream, 0, false, nil); err != nil {
		return errors.WithStack(err)
	}

	tag := repo.ImageTag
	if o.tag != "" {
		if !o.quiet {
			fmt.Printf("Renaming %s to %s ...\n", repo.ImageTag, o.tag)
		}

		if err := docker.ImageTag(ctx, repo.ImageTag, o.tag); err != nil {
			return errors.Wrap(err, "failed to tag image")
		}

		// We ignore the error here intentionally. Cleaning up is best-effort
		// and we can't do anything to recover if this fails.
		_, _ = docker.ImageRemove(ctx, repo.ImageTag, types.ImageRemoveOptions{})
		tag = o.tag
	}

	if o.quiet {
		fmt.Println(tag)
	} else {
		fmt.Println("Done.")
	}
	return nil
}
