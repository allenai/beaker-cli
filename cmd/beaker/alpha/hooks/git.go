package hooks

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetTopDir returns the absolute path to the base directory for the current repo or an
// error if it fails to determine the top directory
func GetTopDir() (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("resolving root directory for git repo: %w", err)
	}
	topDir := strings.TrimSpace(buf.String())
	return topDir, nil
}
