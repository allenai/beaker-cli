package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/beaker/client/api"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newImageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image <command>",
		Short: "Manage images",
	}
	cmd.AddCommand(newImageCommitCommand())
	cmd.AddCommand(newImageCreateCommand())
	cmd.AddCommand(newImageDeleteCommand())
	cmd.AddCommand(newImageGetCommand())
	cmd.AddCommand(newImagePullCommand())
	cmd.AddCommand(newImageRenameCommand())
	return cmd
}

func newImageCommitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "commit <image>",
		Short: "Commit an image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := beaker.Image(args[0]).Commit(ctx); err != nil {
				return err
			}

			if !quiet {
				fmt.Printf("Committed %s\n", color.BlueString(args[0]))
			}
			return nil
		},
	}
}

func newImageCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <docker image ID>",
		Short: "Create a new image",
		Args:  cobra.ExactArgs(1),
	}

	var description string
	var name string
	var workspace string
	cmd.Flags().StringVar(&description, "description", "", "Image description")
	cmd.Flags().StringVarP(&name, "name", "n", "", "Image name")
	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Image workspace")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var err error
		if workspace, err = ensureWorkspace(workspace); err != nil {
			return err
		}

		docker, err := docker.NewClientWithOpts(docker.FromEnv)
		if err != nil {
			return fmt.Errorf("failed to create Docker client: %w", err)
		}

		imageTag := args[0]
		dockerImage, _, err := docker.ImageInspectWithRaw(ctx, imageTag)
		if err != nil {
			return err
		}

		spec := api.ImageSpec{
			Description: description,
			ImageID:     dockerImage.ID,
			ImageTag:    imageTag,
			Workspace:   workspace,
		}
		image, err := beaker.CreateImage(ctx, spec, name)
		if err != nil {
			return err
		}

		if !quiet {
			if name == "" {
				fmt.Printf("Pushing %s as %s ...\n", imageTag, color.BlueString(image.Ref()))
			} else {
				fmt.Printf("Pushing %s as %s (%s)...\n", imageTag, color.BlueString(name), image.Ref())
			}
		}

		repo, err := image.Repository(ctx, true)
		if err != nil {
			return fmt.Errorf("failed to retrieve credentials for remote repository: %w", err)
		}

		// Tag the image to the remote repository.
		if err := docker.ImageTag(ctx, imageTag, repo.ImageTag); err != nil {
			return fmt.Errorf("failed to set remote image tag: %w", err)
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
			return fmt.Errorf("failed to encode remote repository auth: %w", err)
		}
		authStr := base64.URLEncoding.EncodeToString(authJSON)

		r, err := docker.ImagePush(ctx, repo.ImageTag, types.ImagePushOptions{RegistryAuth: authStr})
		if err != nil {
			return err
		}
		// Display push responses as the Docker CLI would. This also translates remote errors.
		var stream io.Writer = os.Stdout
		if quiet {
			stream = ioutil.Discard
		}
		if err := jsonmessage.DisplayJSONMessagesStream(r, stream, 0, false, nil); err != nil {
			_ = r.Close()
			return err
		}
		if err := r.Close(); err != nil {
			return err
		}

		if err := image.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit image: %w", err)
		}

		if quiet {
			fmt.Println(image.Ref())
		} else {
			fmt.Println("Done.")
		}
		return nil
	}
	return cmd
}

func newImageDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <image>",
		Short: "Permanently delete an image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := beaker.Image(args[0]).Delete(ctx); err != nil {
				return err
			}

			if !quiet {
				fmt.Printf("Deleted %s\n", color.BlueString(args[0]))
			}
			return nil
		},
	}
}

func newImageGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "get <image...>",
		Aliases: []string{"inspect"},
		Short:   "Display detailed information about one or more images",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var images []api.Image
			for _, name := range args {
				image, err := beaker.Image(name).Get(ctx)
				if err != nil {
					return err
				}
				images = append(images, *image)
			}
			return printImages(images)
		},
	}
}

func newImagePullCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pull <image> [tag]",
		Short: "Pull an image",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			imageRef := args[0]
			var tag string
			if len(args) > 1 {
				tag = args[1]
			}

			docker, err := docker.NewClientWithOpts(docker.FromEnv)
			if err != nil {
				return errors.Wrap(err, "failed to create Docker client")
			}

			repo, err := beaker.Image(imageRef).Repository(ctx, false)
			if err != nil {
				return errors.WithMessage(err, "failed to retrieve credentials for remote repository")
			}

			if !quiet {
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
			if quiet {
				stream = ioutil.Discard
			}
			if err := jsonmessage.DisplayJSONMessagesStream(r, stream, 0, false, nil); err != nil {
				return errors.WithStack(err)
			}

			if tag != "" {
				if !quiet {
					// We intentionally print the un-mangled tag.
					fmt.Printf("Renaming %s to %s ...\n", repo.ImageTag, tag)
				}

				// We must normalize or ImageTag will return an error on otherwise valid references.
				normalized, err := reference.ParseNormalizedNamed(tag)
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
			} else {
				tag = repo.ImageTag
			}

			if quiet {
				fmt.Println(tag)
			} else {
				fmt.Println("Done.")
			}
			return nil
		},
	}
}

func newImageRenameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <image> <name>",
		Short: "Rename an image",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName := args[0]
			newName := args[1]

			image, err := beaker.Image(oldName).Patch(ctx, api.ImagePatch{
				Name: &newName,
			})
			if err != nil {
				return err
			}

			if quiet {
				fmt.Println(image.ID)
			} else {
				fmt.Printf("Renamed %s to %s\n", color.BlueString(oldName), image.FullName)
			}
			return nil
		},
	}
}
