package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/jefferson-calmon/clean-my-cache/internal/engine"
	"github.com/jefferson-calmon/clean-my-cache/internal/system"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type options struct {
	only    []string
	exclude []string
	all     bool
	yes     bool
	dryRun  bool
	trash   bool
	json    bool
	verbose bool
}

func (o options) selection() engine.Selection {
	return engine.Selection{Only: o.only, Exclude: o.exclude, IncludeCaution: o.all}
}

type app struct {
	out         io.Writer
	errOut      io.Writer
	interactive bool
	home        string
	runner      system.Runner
}

var errExitCode = errors.New("exit with failure")

func newApp() (*app, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return &app{
		out:         os.Stdout,
		errOut:      os.Stderr,
		interactive: term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())),
		home:        home,
		runner:      system.ExecRunner{},
	}, nil
}

func (a *app) engine(ctx context.Context) (*engine.Engine, error) {
	set, err := engine.LoadRecipes(a.home)
	if err != nil {
		return nil, fmt.Errorf("invalid recipe: %w", err)
	}
	return engine.New(ctx, a.home, a.runner, set), nil
}

func (a *app) println(args ...any) {
	fmt.Fprintln(a.out, args...)
}

func requireMacOS() error {
	if runtime.GOOS == "darwin" {
		return nil
	}
	return errors.New("clean-my-cache supports macOS only for now. Linux support is on the roadmap")
}

func newRootCommand(a *app, version string) *cobra.Command {
	opts := options{}

	root := &cobra.Command{
		Use:           "clean-my-cache",
		Short:         "Find and safely clean developer and app caches",
		Long:          "Find and safely clean developer and app caches.\nSee what each cache is before you delete it.",
		Version:       version,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.runClean(cmd.Context(), opts)
		},
	}

	flags := root.PersistentFlags()
	flags.StringSliceVar(&opts.only, "only", nil, "clean only these items, e.g. --only npm,slack,'docker-*'")
	flags.StringSliceVar(&opts.exclude, "exclude", nil, "never touch these items")
	flags.BoolVar(&opts.all, "all", false, `also include items marked "caution"`)
	flags.BoolVarP(&opts.yes, "yes", "y", false, "skip the questions and clean the default selection")
	flags.BoolVarP(&opts.dryRun, "dry-run", "n", false, "show what would be cleaned, delete nothing")
	flags.BoolVar(&opts.trash, "trash", false, "move folders to the Trash instead of deleting them")
	flags.BoolVar(&opts.json, "json", false, "machine-readable output")
	flags.BoolVarP(&opts.verbose, "verbose", "v", false, "also show empty and unavailable items")

	root.AddCommand(
		&cobra.Command{
			Use:   "scan",
			Short: "Only show what can be cleaned",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				scanOnly := opts
				scanOnly.dryRun = true
				return a.runClean(cmd.Context(), scanOnly)
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List everything it knows how to clean",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return a.runList(opts.json)
			},
		},
		&cobra.Command{
			Use:   "explain <id>",
			Short: "Explain what an item is before you delete it",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return a.runExplain(args[0])
			},
		},
	)

	return root
}

func Execute(version string) int {
	a, err := newApp()
	if err == nil {
		err = newRootCommand(a, version).ExecuteContext(context.Background())
	}

	switch {
	case err == nil:
		return 0
	case errors.Is(err, errExitCode):
		return 1
	default:
		fmt.Fprintf(os.Stderr, "clean-my-cache: %v\n", err)
		return 1
	}
}
