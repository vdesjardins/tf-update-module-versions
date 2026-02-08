package report

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const toolTimeout = 10 * time.Second

type writer interface {
	Write(input string) (string, error)
}

type defaultWriter struct{}

func (w *defaultWriter) Write(input string) (string, error) {
	return input, nil
}

type toolWriter struct {
	command string
}

func newToolWriter(command string) (*toolWriter, error) {
	if strings.TrimSpace(command) == "" {
		return nil, fmt.Errorf("diff tool command is empty")
	}
	return &toolWriter{command: command}, nil
}

func (w *toolWriter) Write(input string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), toolTimeout)
	defer cancel()

	parts := strings.Fields(w.command)
	if len(parts) == 0 {
		return "", fmt.Errorf("diff tool command is empty")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Stdin = strings.NewReader(input)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("diff tool failed: %s", strings.TrimSpace(stderr.String()))
		}
		return "", fmt.Errorf("diff tool failed: %w", err)
	}

	return stdout.String(), nil
}
