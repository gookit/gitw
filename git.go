package gitwrap

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/gookit/color"
)

// from: https://github.com/github/hub/blob/master/cmd/cmd.go

// GitCmd is a project-wide struct that represents a command to be run in the console.
type GitCmd struct {
	Name   string
	Args   []string
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}

// New instance
func New(name string) *GitCmd {
	return &GitCmd{
		Name:   name,
		Args:   []string{},
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// NewGit instance
func NewGit(args ...string) *GitCmd {
	return NewWithArgs("git", args)
}

// NewWithArgs instance
func NewWithArgs(name string, args []string) *GitCmd {
	return &GitCmd{
		Name:   name,
		Args:   args,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

func (cmd GitCmd) String() string {
	args := make([]string, len(cmd.Args))
	for i, a := range cmd.Args {
		if strings.ContainsRune(a, '"') {
			args[i] = fmt.Sprintf(`'%s'`, a)
		} else if a == "" || strings.ContainsRune(a, '\'') || strings.ContainsRune(a, ' ') {
			args[i] = fmt.Sprintf(`"%s"`, a)
		} else {
			args[i] = a
		}
	}
	return fmt.Sprintf("%s %s", cmd.Name, strings.Join(args, " "))
}

// WithArg returns the current argument
func (cmd *GitCmd) WithArg(arg string) *GitCmd {
	cmd.Args = append(cmd.Args, arg)

	return cmd
}

func (cmd *GitCmd) WithArgs(args ...string) *GitCmd {
	for _, arg := range args {
		cmd.WithArg(arg)
	}

	return cmd
}

func (cmd *GitCmd) Output() (string, error) {
	verboseLog(cmd)
	c := exec.Command(cmd.Name, cmd.Args...)
	c.Stderr = cmd.Stderr
	output, err := c.Output()

	return string(output), err
}

func (cmd *GitCmd) CombinedOutput() (string, error) {
	verboseLog(cmd)
	output, err := exec.Command(cmd.Name, cmd.Args...).CombinedOutput()

	return string(output), err
}

// Success exec
func (cmd *GitCmd) Success() bool {
	verboseLog(cmd)
	err := exec.Command(cmd.Name, cmd.Args...).Run()
	return err == nil
}

// Run runs command with `Exec` on platforms except Windows
// which only supports `Spawn`
func (cmd *GitCmd) Run() error {
	if isWindows() {
		return cmd.Spawn()
	}
	return cmd.Exec()
}

func isWindows() bool {
	return runtime.GOOS == "windows" || detectWSL()
}

var detectedWSL bool
var detectedWSLContents string

// https://github.com/Microsoft/WSL/issues/423#issuecomment-221627364
func detectWSL() bool {
	if !detectedWSL {
		b := make([]byte, 1024)
		f, err := os.Open("/proc/version")
		if err == nil {
			_,_ = f.Read(b) // ignore error
			f.Close()
			detectedWSLContents = string(b)
		}
		detectedWSL = true
	}
	return strings.Contains(detectedWSLContents, "Microsoft")
}

// Spawn runs command with spawn(3)
func (cmd *GitCmd) Spawn() error {
	verboseLog(cmd)
	c := exec.Command(cmd.Name, cmd.Args...)
	c.Stdin = cmd.Stdin
	c.Stdout = cmd.Stdout
	c.Stderr = cmd.Stderr

	return c.Run()
}

// Exec runs command with exec(3)
// Note that Windows doesn't support exec(3): http://golang.org/src/pkg/syscall/exec_windows.go#L339
func (cmd *GitCmd) Exec() error {
	verboseLog(cmd)

	binary, err := exec.LookPath(cmd.Name)
	if err != nil {
		return &exec.Error{
			Name: cmd.Name,
			Err:  fmt.Errorf("command not found"),
		}
	}

	args := []string{binary}
	args = append(args, cmd.Args...)

	return syscall.Exec(binary, args, os.Environ())
}

func verboseLog(cmd *GitCmd) {
	if debug {
		msg := fmt.Sprintf("> %s", cmd.String())
		color.Infoln(msg)
	}
}
