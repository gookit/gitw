package gmoji_test

import (
	"testing"

	"github.com/gookit/gitw/gmoji"
	"github.com/gookit/goutil/testutil/assert"
)

func TestEmojiMap_String(t *testing.T) {
	em, err := gmoji.Emojis("")
	assert.NoErr(t, err)
	assert.NotEmpty(t, em)

	// fmt.Println(em.String())

	em, err = gmoji.Emojis(gmoji.LangZH)
	assert.NoErr(t, err)
	assert.NotEmpty(t, em)
	// fmt.Println(em.String())
}
