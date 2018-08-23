package client

import (
	"context"
	"net/http"
	"path"

	"github.com/pkg/errors"

	"github.com/allenai/beaker/api"
)

// TaskHandle provides operations on a task.
type TaskHandle struct {
	client *Client
	id     string
}

// Task gets a handle for a task by name or ID. The returned handle is
// guaranteed throughout its lifetime to refer to the same object, even if that
// object is later renamed.
func (c *Client) Task(ctx context.Context, reference string) (*TaskHandle, error) {
	id, err := c.resolveRef(ctx, "/api/v3/tasks", reference)
	if err != nil {
		return nil, errors.WithMessage(err, "could not resolve task reference "+reference)
	}

	return &TaskHandle{client: c, id: id}, nil
}

// ID returns a task's stable, unique ID.
func (h *TaskHandle) ID() string {
	return h.id
}

// Get retrieves a task's details.
func (h *TaskHandle) Get(ctx context.Context) (*api.Task, error) {
	path := path.Join("/api/v3/tasks", h.id)
	resp, err := h.client.sendRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var task api.Task
	if err := parseResponse(resp, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// GetResults retrieves a task's results.
func (h *TaskHandle) GetResults(ctx context.Context) (*api.TaskResults, error) {
	path := path.Join("/api/v3/tasks", h.id, "results")
	resp, err := h.client.sendRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp.Body)

	var results api.TaskResults
	if err := parseResponse(resp, &results); err != nil {
		return nil, err
	}

	return &results, nil
}

// SetDescription sets a task's description.
func (h *TaskHandle) SetDescription(ctx context.Context, description string) error {
	path := path.Join("/api/v3/tasks", h.id)
	body := api.TaskPatchSpec{Description: &description}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// Stop cancels a task. If the task has already completed, this succeeds with no effect.
func (h *TaskHandle) Stop(ctx context.Context) error {
	path := path.Join("/api/v3/tasks", h.id)
	body := api.TaskPatchSpec{Cancel: true}
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, body)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// SetStatus overrides a task's status.
// TODO: Move this to an internal-only API. External clients should call Stop to cancel tasks.
func (h *TaskHandle) SetStatus(ctx context.Context, status api.TaskStatus) error {
	path := path.Join("/api/tasks", h.id, "status")
	query := map[string]string{"status": string(status)}
	resp, err := h.client.sendRequest(ctx, http.MethodPut, path, query, nil)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}

// PatchPermissions ammends a task's permissions.
func (h *TaskHandle) PatchPermissions(
	ctx context.Context,
	permissionPatch api.PermissionPatch,
) error {
	path := path.Join("/api/v3/tasks", h.id, "auth")
	resp, err := h.client.sendRequest(ctx, http.MethodPatch, path, nil, permissionPatch)
	if err != nil {
		return err
	}
	defer safeClose(resp.Body)
	return errorFromResponse(resp)
}
