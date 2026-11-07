package system

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"
	"time"
)

const commandTimeout = 10 * time.Minute

type Result struct {
	OK     bool
	Stdout string
}

type Runner interface {
	Run(ctx context.Context, argv []string) Result
	HasCommand(bin string) bool
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, argv []string) Result {
	if len(argv) == 0 {
		return Result{}
	}

	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Stdout = &stdout
	err := cmd.Run()
	return Result{OK: err == nil, Stdout: stdout.String()}
}

func (ExecRunner) HasCommand(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

var appBundlePattern = regexp.MustCompile(`/([^/\n]+)\.app/Contents/`)

func RunningApps(ctx context.Context, runner Runner) map[string]bool {
	apps := map[string]bool{}
	result := runner.Run(ctx, []string{"ps", "-axo", "command="})
	if !result.OK {
		return apps
	}
	for _, match := range appBundlePattern.FindAllStringSubmatch(result.Stdout, -1) {
		apps[match[1]] = true
	}
	return apps
}
