// Package gitw git command wrapper, git changelog, repo information and some git tools.
package gitw

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/sysutil/cmdr"
)

// some from: https://github.com/github/hub/blob/master/cmd/cmd.go

const (
	// GitDir name
	GitDir = ".git"
	// HeadFile in .git/
	HeadFile = "HEAD"
	// ConfFile in .git/
	ConfFile = "config"
	// GitHubHost name
	GitHubHost = "github.com"
	// GitHubURL string
	GitHubURL = "https://github.com"
	// GitHubGit string
	GitHubGit = "git@github.com"
)

// git host type
const (
	TypeGitHub  = "github"
	TypeGitlab  = "gitlab"
	TypeDefault = "git"
)

const (
	// DefaultBin name
	DefaultBin = "git"

	// DefaultBranchName value
	DefaultBranchName = "master"
	// DefaultRemoteName value
	DefaultRemoteName = "origin"
)

// GitWrap is a project-wide struct that represents a command to be run in the console.
type GitWrap struct {
	// Workdir for run git
	Workdir string
	// Bin git bin name. default is "git"
	Bin string
	// Args for run git. contains git command name.
	Args []string
	// Stdin more settings
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// DryRun if True, not real execute command
	DryRun bool
	// BeforeExec command hook.
	//
	// Usage: gw.BeforeExec = gitw.PrintCmdline
	BeforeExec func(gw *GitWrap)
}

// New create instance with args
func New(args ...string) *GitWrap {
	return &GitWrap{
		Bin:  DefaultBin,
		Args: args,
		// Stdin:  os.Stdin, // not init stdin
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// Cmd create instance with git cmd and args
func Cmd(cmd string, args ...string) *GitWrap {
	return New(cmd).WithArgs(args)
}

// NewWithArgs create instance with git cmd and args
func NewWithArgs(cmd string, args ...string) *GitWrap {
	return New(cmd).WithArgs(args)
}

// NewWithWorkdir create instance with workdir and args
func NewWithWorkdir(workdir string, args ...string) *GitWrap {
	return New(args...).WithWorkDir(workdir)
}

// New git wrap from current instance, can with args
func (gw *GitWrap) New(args ...string) *GitWrap {
	nw := *gw
	nw.Args = args
	return &nw
}

// Sub new sub git cmd from current instance, can with args
func (gw *GitWrap) Sub(cmd string, args ...string) *GitWrap {
	return gw.Cmd(cmd, args...)
}

// Cmd new git wrap from current instance, can with args
func (gw *GitWrap) Cmd(cmd string, args ...string) *GitWrap {
	nw := *gw
	nw.Args = []string{cmd}

	if len(args) > 0 {
		nw.WithArgs(args)
	}
	return &nw
}

// WithFn for setting gw
func (gw *GitWrap) WithFn(fn func(gw *GitWrap)) *GitWrap {
	fn(gw)
	return gw
}

// -------------------------------------------------
// config the git command
// -------------------------------------------------

// String to command line
func (gw *GitWrap) String() string {
	return gw.Cmdline()
}

// Cmdline to command line
func (gw *GitWrap) Cmdline() string {
	b := new(strings.Builder)
	b.WriteString(gw.Bin)

	for _, a := range gw.Args {
		b.WriteByte(' ')
		if strings.ContainsRune(a, '"') {
			b.WriteString(fmt.Sprintf(`'%s'`, a))
		} else if a == "" || strings.ContainsRune(a, '\'') || strings.ContainsRune(a, ' ') {
			b.WriteString(fmt.Sprintf(`"%s"`, a))
		} else {
			b.WriteString(a)
		}
	}
	return b.String()
}

// IsGitRepo return the work dir is a git repo.
func (gw *GitWrap) IsGitRepo() bool {
	return fsutil.IsDir(gw.GitDir())
}

// GitDir return .git data dir
func (gw *GitWrap) GitDir() string {
	gitDir := GitDir
	if gw.Workdir != "" {
		gitDir = gw.Workdir + "/" + GitDir
	}

	return gitDir
}

// -------------------------------------------------
// config the git command
// -------------------------------------------------

// PrintCmdline on exec command
func (gw *GitWrap) PrintCmdline() *GitWrap {
	gw.BeforeExec = PrintCmdline
	return gw
}

// WithDryRun on exec command
func (gw *GitWrap) WithDryRun(dryRun bool) *GitWrap {
	gw.DryRun = dryRun
	return gw
}

// OnBeforeExec add hook
func (gw *GitWrap) OnBeforeExec(fn func(gw *GitWrap)) *GitWrap {
	gw.BeforeExec = fn
	return gw
}

// WithWorkDir returns the current object
func (gw *GitWrap) WithWorkDir(dir string) *GitWrap {
	gw.Workdir = dir
	return gw
}

// WithStdin returns the current argument
func (gw *GitWrap) WithStdin(in io.Reader) *GitWrap {
	gw.Stdin = in
	return gw
}

// WithOutput returns the current argument
func (gw *GitWrap) WithOutput(out, errOut io.Writer) *GitWrap {
	gw.Stdout = out
	if errOut != nil {
		gw.Stderr = errOut
	}
	return gw
}

// WithArg add args and returns the current object. alias of the WithArg()
func (gw *GitWrap) WithArg(args ...string) *GitWrap {
	gw.Args = append(gw.Args, args...)
	return gw
}

// AddArg add args and returns the current object
func (gw *GitWrap) AddArg(args ...string) *GitWrap {
	return gw.WithArg(args...)
}

// Argf add arg and returns the current object.
func (gw *GitWrap) Argf(format string, args ...interface{}) *GitWrap {
	gw.Args = append(gw.Args, fmt.Sprintf(format, args...))
	return gw
}

// WithArgf add arg and returns the current object. alias of the Argf()
func (gw *GitWrap) WithArgf(format string, args ...interface{}) *GitWrap {
	return gw.Argf(format, args...)
}

// ArgIf add arg and returns the current object
func (gw *GitWrap) ArgIf(arg string, exprOk bool) *GitWrap {
	if exprOk {
		gw.Args = append(gw.Args, arg)
	}
	return gw
}

// WithArgIf add arg and returns the current object
func (gw *GitWrap) WithArgIf(arg string, exprOk bool) *GitWrap {
	return gw.ArgIf(arg, exprOk)
}

// AddArgs for the git. alias of WithArgs()
func (gw *GitWrap) AddArgs(args []string) *GitWrap {
	return gw.WithArgs(args)
}

// WithArgs for the git
func (gw *GitWrap) WithArgs(args []string) *GitWrap {
	if len(args) > 0 {
		gw.Args = append(gw.Args, args...)
	}
	return gw
}

// WithArgsIf add arg and returns the current object
func (gw *GitWrap) WithArgsIf(args []string, exprOk bool) *GitWrap {
	if exprOk && len(args) > 0 {
		gw.Args = append(gw.Args, args...)
	}
	return gw
}

// ResetArgs for git
func (gw *GitWrap) ResetArgs() {
	gw.Args = make([]string, 0)
}

// -------------------------------------------------
// run git command
// -------------------------------------------------

// NewExecCmd create exec.Cmd from current cmd
func (gw *GitWrap) NewExecCmd() *exec.Cmd {
	c := exec.Command(gw.Bin, gw.Args...)
	c.Dir = gw.Workdir
	c.Stdin = gw.Stdin
	c.Stdout = gw.Stdout
	c.Stderr = gw.Stderr

	return c
}

// Success run and return whether success
func (gw *GitWrap) Success() bool {
	if gw.BeforeExec != nil {
		gw.BeforeExec(gw)
	}
	if gw.DryRun {
		return true
	}

	return gw.NewExecCmd().Run() == nil
}

// SafeLines run and return output as lines
func (gw *GitWrap) SafeLines() []string {
	ss, _ := gw.OutputLines()
	return ss
}

// OutputLines run and return output as lines
func (gw *GitWrap) OutputLines() ([]string, error) {
	out, err := gw.Output()
	if err != nil {
		return nil, err
	}
	return cmdr.OutputLines(out), err
}

// SafeOutput run and return output
func (gw *GitWrap) SafeOutput() string {
	gw.Stderr = nil // ignore stderr
	out, err := gw.Output()

	if err != nil {
		return ""
	}
	return out
}

// Output run and return output
func (gw *GitWrap) Output() (string, error) {
	c := exec.Command(gw.Bin, gw.Args...)
	c.Dir = gw.Workdir
	c.Stderr = gw.Stderr

	if gw.BeforeExec != nil {
		gw.BeforeExec(gw)
	}
	if gw.DryRun {
		return "DIY-RUN: OK", nil
	}

	output, err := c.Output()
	return string(output), err
}

// CombinedOutput run and return output, will combine stderr and stdout output
func (gw *GitWrap) CombinedOutput() (string, error) {
	c := exec.Command(gw.Bin, gw.Args...)
	c.Dir = gw.Workdir

	if gw.BeforeExec != nil {
		gw.BeforeExec(gw)
	}
	if gw.DryRun {
		return "DIY-RUN: OK", nil
	}

	output, err := c.CombinedOutput()
	return string(output), err
}

// MustRun a command. will panic on error
func (gw *GitWrap) MustRun() {
	if err := gw.Run(); err != nil {
		panic(err)
	}
}

// Run runs command with `Exec` on platforms except Windows
// which only supports `Spawn`
func (gw *GitWrap) Run() error {
	if gw.BeforeExec != nil {
		gw.BeforeExec(gw)
	}
	if gw.DryRun {
		fmt.Println("DIY-RUN: OK")
		return nil
	}

	return gw.NewExecCmd().Run()

	// if envutil.IsWindows() {
	// 	return gw.Spawn()
	// }
	// return gw.Exec()
}

// Spawn runs command with spawn(3)
func (gw *GitWrap) Spawn() error {
	if gw.DryRun {
		return nil
	}
	return gw.NewExecCmd().Run()
}

// Exec runs command with exec(3)
// Note that Windows doesn't support exec(3): http://golang.org/src/pkg/syscall/exec_windows.go#L339
func (gw *GitWrap) Exec() error {
	binary, err := exec.LookPath(gw.Bin)
	if err != nil {
		return &exec.Error{
			Name: gw.Bin,
			Err:  errorx.Newf("%s not found in the system", gw.Bin),
		}
	}

	args := []string{binary}
	args = append(args, gw.Args...)

	if gw.BeforeExec != nil {
		gw.BeforeExec(gw)
	}
	if gw.DryRun {
		fmt.Println("DIY-RUN: OK")
		return nil
	}

	return syscall.Exec(binary, args, os.Environ())
}

// -------------------------------------------------
// commands of git
// -------------------------------------------------

// Add command for git
func (gw *GitWrap) Add(args ...string) *GitWrap {
	return gw.Cmd("add", args...)
}

// Annotate command for git
func (gw *GitWrap) Annotate(args ...string) *GitWrap {
	return gw.Cmd("annotate", args...)
}

// Apply command for git
func (gw *GitWrap) Apply(args ...string) *GitWrap {
	return gw.Cmd("apply", args...)
}

// Bisect command for git
func (gw *GitWrap) Bisect(args ...string) *GitWrap {
	return gw.Cmd("bisect", args...)
}

// Blame command for git
func (gw *GitWrap) Blame(args ...string) *GitWrap {
	return gw.Cmd("blame", args...)
}

// Branch command for git
func (gw *GitWrap) Branch(args ...string) *GitWrap {
	return gw.Cmd("branch", args...)
}

// Checkout command for git
func (gw *GitWrap) Checkout(args ...string) *GitWrap {
	return gw.Cmd("checkout", args...)
}

// CherryPick command for git
func (gw *GitWrap) CherryPick(args ...string) *GitWrap {
	return gw.Cmd("cherry-pick", args...)
}

// Clean command for git
func (gw *GitWrap) Clean(args ...string) *GitWrap {
	return gw.Cmd("clean", args...)
}

// Clone command for git
func (gw *GitWrap) Clone(args ...string) *GitWrap {
	return gw.Cmd("clone", args...)
}

// Commit command for git
func (gw *GitWrap) Commit(args ...string) *GitWrap {
	return gw.Cmd("commit", args...)
}

// Config command for git
func (gw *GitWrap) Config(args ...string) *GitWrap {
	return gw.Cmd("config", args...)
}

// Describe command for git
func (gw *GitWrap) Describe(args ...string) *GitWrap {
	return gw.Cmd("describe", args...)
}

// Diff command for git
func (gw *GitWrap) Diff(args ...string) *GitWrap {
	return gw.Cmd("diff", args...)
}

// Fetch command for git
func (gw *GitWrap) Fetch(args ...string) *GitWrap {
	return gw.Cmd("fetch", args...)
}

// Grep command for git
func (gw *GitWrap) Grep(args ...string) *GitWrap {
	return gw.Cmd("grep", args...)
}

// Init command for git
func (gw *GitWrap) Init(args ...string) *GitWrap {
	return gw.Cmd("init", args...)
}

// Log command for git
func (gw *GitWrap) Log(args ...string) *GitWrap {
	return gw.Cmd("log", args...)
}

// Merge command for git
func (gw *GitWrap) Merge(args ...string) *GitWrap {
	return gw.Cmd("merge", args...)
}

// Mv command for git
func (gw *GitWrap) Mv(args ...string) *GitWrap {
	return gw.Cmd("mv", args...)
}

// Pull command for git
func (gw *GitWrap) Pull(args ...string) *GitWrap {
	return gw.Cmd("pull", args...)
}

// Push command for git
func (gw *GitWrap) Push(args ...string) *GitWrap {
	return gw.Cmd("push", args...)
}

// Rebase command for git
func (gw *GitWrap) Rebase(args ...string) *GitWrap {
	return gw.Cmd("rebase", args...)
}

// Reflog command for git
func (gw *GitWrap) Reflog(args ...string) *GitWrap {
	return gw.Cmd("reflog", args...)
}

// Remote command for git
func (gw *GitWrap) Remote(args ...string) *GitWrap {
	return gw.Cmd("remote", args...)
}

// Reset command for git
func (gw *GitWrap) Reset(args ...string) *GitWrap {
	return gw.Cmd("reset", args...)
}

// Restore command for git
func (gw *GitWrap) Restore(args ...string) *GitWrap {
	return gw.Cmd("restore", args...)
}

// Revert command for git
func (gw *GitWrap) Revert(args ...string) *GitWrap {
	return gw.Cmd("revert", args...)
}

// RevList command for git
func (gw *GitWrap) RevList(args ...string) *GitWrap {
	return gw.Cmd("rev-list", args...)
}

// RevParse command for git
//
// rev-parse usage:
//
//	git rev-parse --show-toplevel // get git workdir, repo dir.
//	git rev-parse -q --git-dir // get git data dir name. eg: .git
func (gw *GitWrap) RevParse(args ...string) *GitWrap {
	return gw.Cmd("rev-parse", args...)
}

// Rm command for git
func (gw *GitWrap) Rm(args ...string) *GitWrap {
	return gw.Cmd("rm", args...)
}

// ShortLog command for git
func (gw *GitWrap) ShortLog(args ...string) *GitWrap {
	return gw.Cmd("shortlog", args...)
}

// Show command for git
func (gw *GitWrap) Show(args ...string) *GitWrap {
	return gw.Cmd("show", args...)
}

// Stash command for git
func (gw *GitWrap) Stash(args ...string) *GitWrap {
	return gw.Cmd("stash", args...)
}

// Status command for git
func (gw *GitWrap) Status(args ...string) *GitWrap {
	return gw.Cmd("status", args...)
}

// Switch command for git
func (gw *GitWrap) Switch(args ...string) *GitWrap {
	return gw.Cmd("switch", args...)
}

// Tag command for git
func (gw *GitWrap) Tag(args ...string) *GitWrap {
	return gw.Cmd("tag", args...)
}

// Var command for git
func (gw *GitWrap) Var(args ...string) *GitWrap {
	return gw.Cmd("var", args...)
}

// Worktree command for git
func (gw *GitWrap) Worktree(args ...string) *GitWrap {
	return gw.Cmd("worktree", args...)
}
