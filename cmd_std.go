package gitw

// debug for std
var debug bool
var std = newStd()

// IsDebug mode
func IsDebug() bool {
	return debug
}

// SetDebug mode
func SetDebug(open bool) {
	debug = open
	if open {
		std.BeforeExec = PrintCmdline
	} else {
		std.BeforeExec = nil
	}
}

// Std instance get
func Std() *GitWrap { return std }

// RestStd instance
func RestStd() {
	std = newStd()
}

// GlobalFlags for run git command
var GlobalFlags []string

func gitCmd(args ...string) *GitWrap {
	// with global flags
	return std.New(GlobalFlags...).WithArgs(args)
}

func cmdWithArgs(subCmd string, args ...string) *GitWrap {
	// with global flags
	return std.Cmd(subCmd, GlobalFlags...).WithArgs(args)
}

func newStd() *GitWrap {
	gw := New()

	// load debug setting.
	debug = isDebugFromEnv()
	if debug {
		gw.BeforeExec = PrintCmdline
	}
	return gw
}

// -------------------------------------------------
// git commands use std
// -------------------------------------------------

// Branch command of git
func Branch(args ...string) *GitWrap { return std.Branch(args...) }

// Log command of git
//
// Usage: Log("-2").OutputLines()
func Log(args ...string) *GitWrap { return std.Log(args...) }

// RevList command of git
func RevList(args ...string) *GitWrap { return std.RevList(args...) }

// Remote command of git
func Remote(args ...string) *GitWrap { return std.Remote(args...) }

// Show command of git
func Show(args ...string) *GitWrap { return std.Show(args...) }

// Tag command of git
//
// Usage:
// 	Tag("-l").OutputLines()
func Tag(args ...string) *GitWrap { return std.Tag(args...) }
