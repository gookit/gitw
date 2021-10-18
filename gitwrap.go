// Package gitwrap is library warp git commands.
// some code is refer from github/hub
package gitwrap

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/gookit/color"
	"github.com/gookit/goutil/fsutil"
)

// from: https://github.com/github/hub/blob/master/cmd/cmd.go

var (
	DefaultBin = "git"
	GitDir = ".git"
)

// GitWrap is a project-wide struct that represents a command to be run in the console.
type GitWrap struct {
	// Bin git bin name. default is "git"
	Bin string
	// Cmd sub command name of git
	// Cmd  string
	Args []string
	// extra
	WorkDir string
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
	// BeforeExec command
	BeforeExec func(gw *GitWrap)
	// inner
	gitDir string
}

// New create instance with args
func New(args ...string) *GitWrap {
	return &GitWrap{
		Bin:    DefaultBin,
		// Cmd:    cmd,
		Args:   args,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// NewWithArgs create instance with cmd and args
func NewWithArgs(cmd string, args ...string) *GitWrap {
	return New(cmd).WithArgs(args)
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

// String to command line
func (gw *GitWrap) String() string {
	return gw.Cmdline()
}

// OnBeforeExec add hook
func (gw *GitWrap) OnBeforeExec(fn func(gw *GitWrap)) *GitWrap {
	gw.BeforeExec = fn
	return gw
}

// WithWorkDir returns the current object
func (gw *GitWrap) WithWorkDir(dir string) *GitWrap {
	gw.WorkDir = dir
	return gw
}

// WithStdin returns the current argument
func (gw *GitWrap) WithStdin(in *os.File) *GitWrap {
	gw.Stdin = in
	return gw
}

// WithOutput returns the current argument
func (gw *GitWrap) WithOutput(out *os.File, errOut *os.File) *GitWrap {
	gw.Stdout = out
	if errOut != nil {
		gw.Stderr = errOut
	}
	return gw
}

// SubCmd returns the current object
func (gw *GitWrap) SubCmd(cmd string) *GitWrap {
	gw.Args = append(gw.Args, cmd)
	return gw
}

// Add returns the current object
func (gw *GitWrap) Add(args ...string) *GitWrap {
	gw.Args = append(gw.Args, args...)
	return gw
}

// WithArg returns the current object. alias of the Add()
func (gw *GitWrap) WithArg(args ...string) *GitWrap {
	return gw.Add(args...)
}

// Addf add arg and returns the current object.
func (gw *GitWrap) Addf(format string, args ...interface{}) *GitWrap {
	gw.Args = append(gw.Args, fmt.Sprintf(format, args...))
	return gw
}

// WithArgf add arg and returns the current object. alias of the Addf()
func (gw *GitWrap) WithArgf(format string, args ...interface{}) *GitWrap {
	return gw.Addf(format, args...)
}

// AddIf add arg and returns the current object
func (gw *GitWrap) AddIf(arg string, exprOk bool) *GitWrap {
	if exprOk {
		gw.Args = append(gw.Args, arg)
	}
	return gw
}

// WithArgIf add arg and returns the current object
func (gw *GitWrap) WithArgIf(arg string, exprOk bool) *GitWrap {
	return gw.AddIf(arg, exprOk)
}

// WithArgs for the git
func (gw *GitWrap) WithArgs(args []string) *GitWrap {
	gw.Args = append(gw.Args, args...)
	return gw
}

// IsGitRepo return the work dir is an git repo.
func (gw *GitWrap) IsGitRepo() bool {
	return fsutil.IsDir(gw.WorkDir + "/" + GitDir)
}

// GitDir return git data dir
func (gw *GitWrap) GitDir() string {
	if gw.gitDir != "" {
		return gw.gitDir
	}

	if gw.WorkDir != "" {
		gw.gitDir = gw.WorkDir + "/.git"
	} else {
		gw.gitDir = GitDir
	}
	return gw.gitDir
}

// CurrentBranch return current branch name
func (gw *GitWrap) CurrentBranch() string {
	// cat .git/HEAD
	// ref: refs/heads/fea_4_12
	return ""
}

// NewExecCmd create exec.Cmd from current cmd
func (gw *GitWrap) NewExecCmd() *exec.Cmd {
	// create exec.Cmd
	return exec.Command(gw.Bin, gw.Args...)
}

// Success run and return whether success
func (gw *GitWrap) Success() bool {
	verboseLog(gw)
	c := exec.Command(gw.Bin, gw.Args...);

	if gw.BeforeExec != nil {
		gw.BeforeExec(gw)
	}
	return c.Run() == nil
}

// SafeOutput run and return output
func (gw *GitWrap) SafeOutput() string {
	out, err := gw.Output()
	if err != nil {
		return ""
	}

	return out
}

// Output run and return output
func (gw *GitWrap) Output() (string, error) {
	verboseLog(gw)
	c := exec.Command(gw.Bin, gw.Args...)
	c.Stderr = gw.Stderr
	if gw.BeforeExec != nil {
		gw.BeforeExec(gw)
	}

	output, err := c.Output()
	return string(output), err
}

// CombinedOutput run and return output, will combine stderr and stdout output
func (gw *GitWrap) CombinedOutput() (string, error) {
	verboseLog(gw)
	c := exec.Command(gw.Bin, gw.Args...)
	if gw.BeforeExec != nil {
		gw.BeforeExec(gw)
	}

	output, err := c.CombinedOutput()

	return string(output), err
}

// MustRun an command. will panic on error
func (gw *GitWrap) MustRun()  {
	if err := gw.Run(); err != nil {
		panic(err)
	}
}

// Run runs command with `Exec` on platforms except Windows
// which only supports `Spawn`
func (gw *GitWrap) Run() error {
	if isWindows() {
		return gw.Spawn()
	}
	return gw.Exec()
}

// Spawn runs command with spawn(3)
func (gw *GitWrap) Spawn() error {
	verboseLog(gw)
	c := exec.Command(gw.Bin, gw.Args...)
	c.Stdin = gw.Stdin
	c.Stdout = gw.Stdout
	c.Stderr = gw.Stderr

	if gw.BeforeExec != nil {
		gw.BeforeExec(gw)
	}
	return c.Run()
}

// Exec runs command with exec(3)
// Note that Windows doesn't support exec(3): http://golang.org/src/pkg/syscall/exec_windows.go#L339
func (gw *GitWrap) Exec() error {
	verboseLog(gw)

	binary, err := exec.LookPath(gw.Bin)
	if err != nil {
		return &exec.Error{
			Name: gw.Bin,
			Err:  fmt.Errorf("%s not found in the system", gw.Bin),
		}
	}

	args := []string{binary}
	args = append(args, gw.Args...)

	if gw.BeforeExec != nil {
		gw.BeforeExec(gw)
	}
	return syscall.Exec(binary, args, os.Environ())
}

func verboseLog(cmd *GitWrap) {
	if debug {
		PrintCmdline(cmd)
	}
}

func isWindows() bool {
	return runtime.GOOS == "windows" || detectWSL()
}

// PrintCmdline on exec
func PrintCmdline(gw *GitWrap) {
	color.Comment.Println("> ", gw.String())
}

var detectedWSL bool
var detectedWSLContents string

// https://github.com/Microsoft/WSL/issues/423#issuecomment-221627364
func detectWSL() bool {
	if !detectedWSL {
		b := make([]byte, 1024)
		f, err := os.Open("/proc/version")
		if err == nil {
			_, _ = f.Read(b) // ignore error
			f.Close()
			detectedWSLContents = string(b)
		}
		detectedWSL = true
	}
	return strings.Contains(detectedWSLContents, "Microsoft")
}
