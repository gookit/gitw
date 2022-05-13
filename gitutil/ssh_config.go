package gitutil

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gookit/goutil/sysutil"
)

const (
	hostReStr = "(?i)^[ \t]*(host|hostname)[ \t]+(.+)$"
)

// SSHConfig map
type SSHConfig map[string]string

func newSSHConfigReader() *SSHConfigReader {
	configFiles := []string{
		"/etc/ssh_config",
		"/etc/ssh/ssh_config",
	}
	if hdir := sysutil.HomeDir(); hdir != "" {
		userConfig := filepath.Join(hdir, ".ssh", "config")
		configFiles = append([]string{userConfig}, configFiles...)
	}

	return &SSHConfigReader{
		Files: configFiles,
	}
}

// SSHConfigReader struct
type SSHConfigReader struct {
	Files []string
}

// Read config from files
func (r *SSHConfigReader) Read() SSHConfig {
	config := make(SSHConfig)
	hostRe := regexp.MustCompile(hostReStr)

	for _, filename := range r.Files {
		r.readFile(config, hostRe, filename)
	}

	return config
}

func (r *SSHConfigReader) readFile(c SSHConfig, re *regexp.Regexp, f string) error {
	file, err := os.Open(f)
	if err != nil {
		return err
	}
	defer file.Close()

	hosts := []string{"*"}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		match := re.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		names := strings.Fields(match[2])
		if strings.EqualFold(match[1], "host") {
			hosts = names
		} else {
			for _, host := range hosts {
				for _, name := range names {
					c[host] = expandTokens(name, host)
				}
			}
		}
	}

	return scanner.Err()
}

func expandTokens(text, host string) string {
	re := regexp.MustCompile(`%[%h]`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		switch match {
		case "%h":
			return host
		case "%%":
			return "%"
		}
		return ""
	})
}
