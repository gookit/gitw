package gitwrap_test

import (
	"testing"

	"github.com/gookit/gitwrap"
	"github.com/gookit/goutil/dump"
	"github.com/stretchr/testify/assert"
)

func TestRepo_RemoteInfos(t *testing.T) {
	r := gitwrap.NewRepo("./")

	rs := r.AllRemoteInfos()
	assert.NoError(t, r.Err())
	assert.NotEmpty(t, rs)

	assert.True(t, r.HasRemote(gitwrap.DefaultRemoteName))

	rt := r.DefaultRemoteInfo()
	assert.NotEmpty(t, rt)
	assert.True(t, rt.Valid())
	assert.False(t, rt.Invalid())
	assert.Equal(t, gitwrap.DefaultRemoteName, rt.Name)
	assert.Equal(t, "", rt.GitUrl())
	assert.Equal(t, "", rt.HttpUrl())
	assert.Equal(t, "", rt.HttpsUrl())

	rt = r.DefaultRemoteInfo(gitwrap.RemoteTypePush)
	assert.NotEmpty(t, rt)

	dump.P(r.Err(), rs)
}
