package gitwrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// from: https://github.com/github/hub/blob/master/git/git.go

var GlobalFlags []string

func Version() (string, error) {
	versionCmd := gitCmd("version")
	output, err := versionCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running git version: %s", err)
	}
	return firstLine(output), nil
}

var cachedDir string

func Dir() (string, error) {
	if cachedDir != "" {
		return cachedDir, nil
	}

	dirCmd := gitCmd("rev-parse", "-q", "--git-dir")
	dirCmd.Stderr = nil
	output, err := dirCmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository (or any of the parent directories): .git")
	}

	var chdir string
	for i, flag := range GlobalFlags {
		if flag == "-C" {
			dir := GlobalFlags[i+1]
			if filepath.IsAbs(dir) {
				chdir = dir
			} else {
				chdir = filepath.Join(chdir, dir)
			}
		}
	}

	gitDir := firstLine(output)

	if !filepath.IsAbs(gitDir) {
		if chdir != "" {
			gitDir = filepath.Join(chdir, gitDir)
		}

		gitDir, err = filepath.Abs(gitDir)
		if err != nil {
			return "", err
		}

		gitDir = filepath.Clean(gitDir)
	}

	cachedDir = gitDir
	return gitDir, nil
}

func WorkdirName() (string, error) {
	toplevelCmd := gitCmd("rev-parse", "--show-toplevel")
	toplevelCmd.Stderr = nil
	output, err := toplevelCmd.Output()
	dir := firstLine(output)
	if dir == "" {
		return "", fmt.Errorf("unable to determine git working directory")
	}
	return dir, err
}

func HasFile(segments ...string) bool {
	// The blessed way to resolve paths within git dir since Git 2.5.0
	pathCmd := gitCmd("rev-parse", "-q", "--git-path", filepath.Join(segments...))
	pathCmd.Stderr = nil
	if output, err := pathCmd.Output(); err == nil {
		if lines := outputLines(output); len(lines) == 1 {
			if _, err := os.Stat(lines[0]); err == nil {
				return true
			}
		}
	}

	// Fallback for older git versions
	dir, err := Dir()
	if err != nil {
		return false
	}

	s := []string{dir}
	s = append(s, segments...)
	path := filepath.Join(s...)
	if _, err := os.Stat(path); err == nil {
		return true
	}

	return false
}

func Head() (string, error) {
	return SymbolicRef("HEAD")
}

// SymbolicRef reads a branch name from a ref such as "HEAD"
func SymbolicRef(ref string) (string, error) {
	refCmd := gitCmd("symbolic-ref", ref)
	refCmd.Stderr = nil
	output, err := refCmd.Output()
	return firstLine(output), err
}

// SymbolicFullName reads a branch name from a ref such as "@{upstream}"
func SymbolicFullName(name string) (string, error) {
	parseCmd := gitCmd("rev-parse", "--symbolic-full-name", name)
	parseCmd.Stderr = nil
	output, err := parseCmd.Output()
	if err != nil {
		return "", fmt.Errorf("unknown revision or path not in the working tree: %s", name)
	}

	return firstLine(output), nil
}

func Ref(ref string) (string, error) {
	parseCmd := gitCmd("rev-parse", "-q", ref)
	parseCmd.Stderr = nil
	output, err := parseCmd.Output()
	if err != nil {
		return "", fmt.Errorf("unknown revision or path not in the working tree: %s", ref)
	}

	return firstLine(output), nil
}

func RefList(a, b string) ([]string, error) {
	ref := fmt.Sprintf("%s...%s", a, b)
	listCmd := gitCmd("rev-list", "--cherry-pick", "--right-only", "--no-merges", ref)
	listCmd.Stderr = nil
	output, err := listCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Can't load rev-list for %s", ref)
	}

	return outputLines(output), nil
}

func NewRange(a, b string) (*Range, error) {
	parseCmd := gitCmd("rev-parse", "-q", a, b)
	parseCmd.Stderr = nil
	output, err := parseCmd.Output()
	if err != nil {
		return nil, err
	}

	lines := outputLines(output)
	if len(lines) != 2 {
		return nil, fmt.Errorf("can't parse range %s..%s", a, b)
	}
	return &Range{lines[0], lines[1]}, nil
}

type Range struct {
	A string
	B string
}

func (r *Range) IsIdentical() bool {
	return strings.EqualFold(r.A, r.B)
}

func (r *Range) IsAncestor() bool {
	cmd := gitCmd("merge-base", "--is-ancestor", r.A, r.B)
	return cmd.Success()
}

func CommentChar(text string) (string, error) {
	char, err := gitConfigGet("core.commentchar")
	if err != nil {
		return "#", nil
	}

	if char == "auto" {
		lines := strings.Split(text, "\n")
		commentCharCandidates := strings.Split("#;@!$%^&|:", "")
	candidateLoop:
		for _, candidate := range commentCharCandidates {
			for _, line := range lines {
				if strings.HasPrefix(line, candidate) {
					continue candidateLoop
				}
			}
			return candidate, nil
		}
		return "", fmt.Errorf("unable to select a comment character that is not used in the current message")
	}

	return char, nil
}

func Show(sha string) (string, error) {
	gw := New()
	gw.Stderr = nil
	gw.WithArg("-c", "log.showSignature=false")
	gw.WithArg("show").WithArg("-s").WithArg("--format=%s%n%+b").WithArg(sha)

	output, err := gw.Output()
	return strings.TrimSpace(output), err
}

func Log(sha1, sha2 string) (string, error) {
	execCmd := New()
	execCmd.WithArg("-c", "log.showSignature=false").WithArg("log").WithArg("--no-color")
	execCmd.WithArg("--format=%h (%aN, %ar)%n%w(78,3,3)%s%n%+b")
	execCmd.WithArg("--cherry")

	// shaRange := fmt.Sprintf("%s...%s", sha1, sha2)
	shaRange := strings.Join([]string{sha1, "...", sha2}, "")
	execCmd.WithArg(shaRange)

	outputs, err := execCmd.Output()
	if err != nil {
		return "", fmt.Errorf("can't load git log %s..%s", sha1, sha2)
	}

	return outputs, nil
}

func LocalBranches() ([]string, error) {
	branchesCmd := gitCmd("branch", "--list")
	output, err := branchesCmd.Output()
	if err != nil {
		return nil, err
	}

	branches := []string{}
	for _, branch := range outputLines(output) {
		branches = append(branches, branch[2:])
	}
	return branches, nil
}

func Remotes() ([]string, error) {
	remoteCmd := gitCmd("remote", "-v")
	remoteCmd.Stderr = nil
	output, err := remoteCmd.Output()
	return outputLines(output), err
}

// -------------------------------------------------
// git config
// -------------------------------------------------

// Var get by git var.
// all: git var -l
// one: git var GIT_EDITOR
func Var(name string) string {
	val, err := New("var", name).Output()
	if err != nil {
		return ""
	}
	return val
}

// -------------------------------------------------
// git config
// -------------------------------------------------

func Config(name string) string {
	val, err := gitConfigGet(name)
	if err != nil {
		return ""
	}
	return val
}

func ConfigAll(name string) ([]string, error) {
	mode := "--get-all"
	if strings.Contains(name, "*") {
		mode = "--get-regexp"
	}

	configCmd := gitCmd(gitConfigCommand([]string{mode, name})...)
	output, err := configCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("unknown config %s", name)
	}
	return outputLines(output), nil
}

func GlobalConfig(name string) (string, error) {
	return gitConfigGet("--global", name)
}

func SetGlobalConfig(name, value string) error {
	_, err := gitConfig("--global", name, value)
	return err
}

func gitConfigGet(args ...string) (string, error) {
	configCmd := gitCmd(gitConfigCommand(args)...)
	output, err := configCmd.Output()
	if err != nil {
		return "", fmt.Errorf("unknown config %s", args[len(args)-1])
	}

	return firstLine(output), nil
}

func gitConfig(args ...string) ([]string, error) {
	configCmd := gitCmd(gitConfigCommand(args)...)
	output, err := configCmd.Output()
	return outputLines(output), err
}

func gitConfigCommand(args []string) []string {
	cmd := []string{"config"}
	return append(cmd, args...)
}

func Alias(name string) string {
	return Config("alias." + name)
}

func Run(args ...string) error {
	cmd := gitCmd(args...)
	return cmd.Run()
}

func Spawn(args ...string) error {
	cmd := gitCmd(args...)
	return cmd.Spawn()
}

func Quiet(args ...string) bool {
	return New(args...).Success()
}

func IsGitDir(dir string) bool {
	cmd := New("--git-dir="+dir, "rev-parse", "--git-dir")
	return cmd.Success()
}

func gitCmd(args ...string) *GitWrap {
	cmd := New()

	// with global flags
	if len(GlobalFlags) > 0 {
		cmd.WithArgs(GlobalFlags)
	}

	return cmd.WithArgs(args)
}

func IsGitCommand(command string) bool {
	helpCmd := gitCmd("help", "--no-verbose", "-a")
	helpCmd.Stderr = nil
	// run
	cmdOutput, err := helpCmd.Output()
	if err != nil {
		// support git versions that don't recognize --no-verbose
		helpCommand := gitCmd("help", "-a")
		cmdOutput, err = helpCommand.Output()
	}
	if err != nil {
		return false
	}

	for _, helpCommandOutputLine := range outputLines(cmdOutput) {
		if strings.HasPrefix(helpCommandOutputLine, "  ") {
			for _, gitCommand := range strings.Split(helpCommandOutputLine, " ") {
				if gitCommand == command {
					return true
				}
			}
		}
	}
	return false
}

func outputLines(output string) []string {
	output = strings.TrimSuffix(output, "\n")
	if output == "" {
		return []string{}
	}
	return strings.Split(output, "\n")
}

func firstLine(output string) string {
	if i := strings.Index(output, "\n"); i >= 0 {
		return output[0:i]
	}
	return output
}
