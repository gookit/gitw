package gitwrap

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/gookit/color"
	"github.com/gookit/goutil"
	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/strutil"
	"github.com/gookit/goutil/sysutil"
	"github.com/gookit/slog"
)

// MustString must return string, will panic on error
func MustString(s string, err error) string {
	goutil.PanicIfErr(err)
	return s
}

// MustStrings must return strings, will panic on error
func MustStrings(ss []string, err error) []string {
	goutil.PanicIfErr(err)
	return ss
}

// PrintCmdline on exec
func PrintCmdline(gw *GitWrap) {
	color.Comment.Println(">", gw.String())
}

var editorCmd string

// Editor returns program name of the editor.
// from https://github.com/alibaba/git-repo-go/blob/master/editor/editor.go
func Editor() string {
	if editorCmd != "" {
		return editorCmd
	}

	var env, str string
	if env = os.Getenv("GIT_EDITOR"); env != "" {
		str = env
	} else if env = Var("GIT_EDITOR"); env != "" { // git var GIT_EDITOR
		str = env
	} else if env = Config("core.editor"); env != "" { // git config --get core.editer OR git config core.editer
		str = env
	} else if env = os.Getenv("VISUAL"); env != "" {
		str = env
	} else if env = os.Getenv("EDITOR"); env != "" {
		str = env
	} else if os.Getenv("TERM") == "dumb" {
		slog.Fatal(
			"No editor specified in GIT_EDITOR, core.editor, VISUAL or EDITOR.\n" +
				"Tried to fall back to vi but terminal is dumb.  Please configure at\n" +
				"least one of these before using this command.")
	} else {
		for _, c := range []string{"vim", "vi", "emacs", "nano"} {
			if path, err := exec.LookPath(c); err == nil {
				str = path
				break
			}
		}
	}

	// remove space and ':'
	editorCmd = strings.Trim(str, ": ")
	return editorCmd
}

// EditText starts an editor to edit data, and returns the edited data.
func EditText(data string) string {
	var (
		err    error
		editor string
	)

	editor = Editor()
	if !sysutil.IsTerminal(os.Stdout.Fd()) {
		slog.Println("no editor, input data unchanged")
		fmt.Println(data)
		return data
	}

	tmpfile, err := ioutil.TempFile("", "go-git-edit-file-*")
	if err != nil {
		slog.Fatal(err)
	}

	//goland:noinspection GoUnhandledErrorResult
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(data)
	if err != nil {
		slog.Fatal(err)
	}
	err = tmpfile.Close()
	if err != nil {
		slog.Fatal(err)
	}

	cmdArgs := editorCommands(editor, tmpfile.Name())
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		slog.Errorf("fail to run '%s' to edit script: %s",
			strings.Join(cmdArgs, " "),
			err)
	}

	f, err := os.Open(tmpfile.Name())
	if err != nil {
		slog.Fatal(err)
	}

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		slog.Fatal(err)
	}
	return string(buf)
}

func editorCommands(editor string, args ...string) []string {
	var cmdArgs []string

	if sysutil.IsWindows() {
		// Split on spaces, respecting quoted strings
		if len(editor) > 0 && (editor[0] == '"' || editor[0] == '\'') {
			cmdArgs = cliutil.ParseLine(editor)

			// if err != nil {
			// 	log.Errorf("fail to parse editor '%s': %s", editor, err)
			// 	cmdArgs = append(cmdArgs, editor)
			// }
		} else {
			for i, c := range editor {
				if c == ' ' || c == '\t' {
					if fsutil.PathExists(editor[:i]) {
						cmdArgs = append(cmdArgs, editor[:i])
						inArgs := cliutil.ParseLine(editor[i+1:])
						cmdArgs = append(cmdArgs, inArgs...)

						// inArgs, err := shellwords.Parse(editor[i+1:])
						// if err != nil {
						// 	log.Errorf("fail to parse args'%s': %s", editor[i+1:], err)
						// 	cmdArgs = append(cmdArgs, editor[i+1:])
						// } else {
						// 	cmdArgs = append(cmdArgs, inArgs...)
						// }
						break
					}
				}
			}
			if len(cmdArgs) == 0 {
				cmdArgs = append(cmdArgs, editor)
			}
		}
	} else if regexp.MustCompile(`^.*[$ \t'].*$`).MatchString(editor) {
		// See: https://gerrit-review.googlesource.com/c/git-repo/+/16156
		cmdArgs = append(cmdArgs, "sh", "-c", editor+` "$@"`, "sh")
	} else {
		cmdArgs = append(cmdArgs, editor)
	}

	cmdArgs = append(cmdArgs, args...)
	return cmdArgs
}

// ErrRemoteInfoNil error
var ErrRemoteInfoNil = errorx.Raw("the remote info data cannot be nil")

// ParseRemoteUrl info to the RemoteInfo object.
func ParseRemoteUrl(URL string, r *RemoteInfo) (err error) {
	if r == nil {
		return ErrRemoteInfoNil
	}

	var str string
	hasSfx := strings.HasSuffix(URL, ".git")

	// eg: "git@github.com:gookit/gitwrap.git"
	if strings.HasPrefix(URL, "git@") {
		r.Proto = ProtoSsh
		if hasSfx {
			str = URL[3 : len(URL)-4]
		} else {
			str = URL[3:]
		}

		host, path, ok := strutil.Cut(str, ":")
		if !ok {
			return errorx.Rawf("invalid git URL: %s", URL)
		}

		group, repo, ok := strutil.Cut(path, "/")
		if !ok {
			return errorx.Rawf("invalid git URL path: %s", URL)
		}

		r.Scheme = "git"
		r.Host, r.Group, r.Repo = host, group, repo
		return nil
	}

	if hasSfx {
		str = URL[0 : len(URL)-4]
	}

	// eg: "https://github.com/gookit/gitwrap.git"
	info, err := url.Parse(str)
	if err != nil {
		return err
	}

	group, repo, ok := strutil.Cut(strings.Trim(info.Path, "/"), "/")
	if !ok {
		return errorx.Rawf("invalid http URL path: %s", URL)
	}

	r.Proto = ProtoHttp
	r.Scheme = info.Scheme
	r.Host, r.Group, r.Repo = info.Host, group, repo
	return nil
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

func isDebugFromEnv() bool {
	return os.Getenv("GIT_CMD_VERBOSE") != ""
}

func verboseLog(cmd *GitWrap) {
	if debug {
		PrintCmdline(cmd)
	}
}

func isWindows() bool {
	return runtime.GOOS == "windows" || detectWSL()
}

var (
	detectedWSL         bool
	detectedWSLContents string
)

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
