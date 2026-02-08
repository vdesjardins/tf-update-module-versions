package color

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Pager manages piping output through a pager command.
type Pager struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

// StartPager starts the pager defined in $PAGER when stdout is a TTY.
// It returns a pager instance and the writer to send output to.
// If no pager should be used, it returns (nil, os.Stdout, nil).
func StartPager() (*Pager, io.Writer, error) {
	if !IsTTY(os.Stdout) {
		return nil, os.Stdout, nil
	}

	pagerCommand := strings.TrimSpace(os.Getenv("PAGER"))
	if pagerCommand == "" {
		return nil, os.Stdout, nil
	}

	parts := strings.Fields(pagerCommand)
	if len(parts) == 0 {
		return nil, os.Stdout, fmt.Errorf("pager command is empty")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, os.Stdout, fmt.Errorf("failed to open pager stdin: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		return nil, os.Stdout, fmt.Errorf("failed to start pager: %w", err)
	}

	return &Pager{cmd: cmd, stdin: stdin}, stdin, nil
}

// Writer returns the pager stdin writer.
func (p *Pager) Writer() io.Writer {
	if p == nil {
		return os.Stdout
	}
	return p.stdin
}

// Close closes the pager stdin and waits for it to exit.
func (p *Pager) Close() error {
	if p == nil {
		return nil
	}
	if p.stdin != nil {
		_ = p.stdin.Close()
	}
	if p.cmd != nil {
		return p.cmd.Wait()
	}
	return nil
}
