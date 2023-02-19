package gmoji

import (
	"embed"
	"encoding/json"
	"errors"
	"regexp"
	"strings"

	"github.com/gookit/goutil"
	"github.com/gookit/goutil/errorx"
	"github.com/gookit/goutil/strutil"
)

//go:embed gitmojis.json gitmojis.zh-CN.json
var emojiFs embed.FS

var codeMatch = regexp.MustCompile(`(:\w+:)`)

// Emoji struct
type Emoji struct {
	Name string
	Code string

	Emoji  string
	Entity string
	Semver string

	Description string
}

// ID name
func (e *Emoji) ID() string {
	return strings.Trim(e.Code, ":")
}

// EmojiMap data. key is code name: Emoji.ID()
type EmojiMap map[string]*Emoji

// Get by code name
func (em EmojiMap) Get(name string) *Emoji {
	return em[name]
}

// First emoji get
func (em EmojiMap) First() *Emoji {
	for _, e := range em {
		return e
	}
	return nil
}

// Lookup by code name
func (em EmojiMap) Lookup(name string) (*Emoji, bool) {
	e, ok := em[name]
	return e, ok
}

// CodeToEmoji convert
func (em EmojiMap) CodeToEmoji(code string) string {
	return em.NameToEmoji(strings.Trim(code, ": "))
}

// NameToEmoji convert
func (em EmojiMap) NameToEmoji(name string) string {
	if e := em.Get(name); e != nil {
		return e.Emoji
	}
	return name
}

// RenderCodes to emojis
func (em EmojiMap) RenderCodes(text string) string {
	// not contains emoji name.
	if strings.IndexByte(text, ':') == -1 {
		return text
	}

	return codeMatch.ReplaceAllStringFunc(text, func(code string) string {
		return em.CodeToEmoji(code) // + " "
	})
}

// FindOne by keywords
func (em EmojiMap) FindOne(keywords ...string) *Emoji {
	sub := em.Search(keywords, 1)
	if len(sub) > 0 {
		for _, e := range sub {
			return e
		}
	}
	return nil
}

// Search by keywords, will match name and description
func (em EmojiMap) Search(keywords []string, limit int) EmojiMap {
	if limit <= 0 {
		limit = 10
	}

	sub := make(EmojiMap, limit)
	for name, emoji := range em {
		var matched = true
		for _, pattern := range keywords {
			if strutil.QuickMatch(pattern, name) {
				continue
			}
			if strutil.QuickMatch(pattern, emoji.Description) {
				continue
			}

			matched = false
			break
		}

		if matched {
			sub[name] = emoji
			if len(sub) >= limit {
				break
			}
		}
	}

	return sub
}

// Len of map
func (em EmojiMap) Len() int {
	return len(em)
}

// String format
func (em EmojiMap) String() string {
	var sb strutil.Builder
	sb.Grow(len(em) * 16)

	for _, emoji := range em {
		sb.Writef("%28s %s  %s\n", emoji.Code, emoji.Emoji, emoji.Description)
	}
	return sb.String()
}

// languages
const (
	LangEN = "en"
	LangZH = "zh-CN"
)

// key is language
var cache = make(map[string]EmojiMap, 2)

func tryLoad(lang string) (err error) {
	if _, ok := cache[lang]; ok {
		return nil
	}

	var bs []byte
	if lang == LangEN {
		bs, err = emojiFs.ReadFile("gitmojis.json")
	} else if lang == LangZH {
		bs, err = emojiFs.ReadFile("gitmojis." + lang + ".json")
	} else {
		err = errors.New("git-emoji: unsupported lang " + lang)
	}

	if err == nil {
		ls := make([]*Emoji, 64)
		err = json.Unmarshal(bs, &ls)
		if err == nil {
			em := make(EmojiMap)
			for _, e := range ls {
				em[e.ID()] = e
			}
			cache[lang] = em
		}
	}
	return
}

// Emojis for given language
func Emojis(lang string) (EmojiMap, error) {
	if lang == "" {
		lang = LangEN
	}

	err := tryLoad(lang)
	if err != nil {
		return nil, err
	}

	em, ok := cache[lang]
	if !ok {
		err = errorx.Rawf("emoji map data not found for lang %q", lang)
	}
	return em, err
}

// MustEmojis load and get
func MustEmojis(lang string) EmojiMap {
	return goutil.Must(Emojis(lang))
}
