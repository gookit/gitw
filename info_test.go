package gitw_test

import (
	"testing"

	"github.com/gookit/gitw"
	"github.com/stretchr/testify/assert"
)

func TestNewRemoteInfo(t *testing.T) {
	URL := "https://github.com/gookit/gitw"

	rt, err := gitw.NewRemoteInfo("origin", URL, gitw.RemoteTypePush)
	assert.NoError(t, err)
	assert.True(t, rt.Valid())
	assert.False(t, rt.Invalid())
	assert.Equal(t, "origin", rt.Name)
	assert.Equal(t, gitw.RemoteTypePush, rt.Type)
	assert.Equal(t, "github.com", rt.Host)
	assert.Equal(t, "gookit/gitw", rt.RepoPath())
	assert.Equal(t, gitw.SchemeHTTPS, rt.Scheme)
	assert.Equal(t, gitw.ProtoHTTP, rt.Proto)
	assert.Equal(t, rt.URL, rt.RawURLOfHTTP())

	URL = "git@github.com:gookit/gitw.git"
	rt, err = gitw.NewRemoteInfo("origin", URL, gitw.RemoteTypePush)
	assert.NoError(t, err)
	assert.Equal(t, "github.com", rt.Host)
	assert.Equal(t, "gookit/gitw", rt.Path())
	assert.Equal(t, gitw.SchemeGIT, rt.Scheme)
	assert.Equal(t, gitw.ProtoSSH, rt.Proto)
	assert.Equal(t, "https://github.com/gookit/gitw", rt.RawURLOfHTTP())
}
