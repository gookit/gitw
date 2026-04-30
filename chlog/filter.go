package chlog

import (
	"strings"
	"unicode"

	"github.com/gookit/goutil/strutil"
)

// ItemFilter interface
type ItemFilter interface {
	// Handle filtering
	Handle(li *LogItem) bool
}

// ItemFilterFunc define. return False to filter(discard) item.
type ItemFilterFunc func(li *LogItem) bool

// Handle filtering
func (f ItemFilterFunc) Handle(li *LogItem) bool {
	return f(li)
}

// built in filters
const (
	FilterMsgLen   = "msg_len"
	FilterWordsLen = "words_len"
	FilterKeyword  = "keyword"
	FilterKeywords = "keywords"
)

// MsgLenFilter handler
func MsgLenFilter(minLen int) ItemFilterFunc {
	return func(li *LogItem) bool {
		return len(li.Msg) > minLen
	}
}

// WordsLenFilter handler
//  - For English text: counts words separated by whitespace
//  - For Chinese text: counts characters (runes) since Chinese doesn't use spaces
//  - For mixed text: counts both English words and Chinese characters
func WordsLenFilter(minLen int) ItemFilterFunc {
	return func(li *LogItem) bool {
		msg := li.Msg
		wordCount := 0
		inWord := false
		hasChinese := false

		for _, r := range msg {
			if unicode.Is(unicode.Han, r) {
				hasChinese = true
				wordCount++
				inWord = false
			} else if unicode.IsSpace(r) {
				inWord = false
			} else {
				if !inWord {
					wordCount++
					inWord = true
				}
			}
		}

		if !hasChinese {
			return len(strutil.Split(msg, " ")) > minLen
		}

		return wordCount > minLen
	}
}

// KeywordFilter filter log item by keyword
func KeywordFilter(kw string, exclude bool) ItemFilterFunc {
	return func(li *LogItem) bool {
		has := strings.Contains(li.Msg, kw)

		if exclude {
			return !has
		}
		return has
	}
}

// KeywordsFilter filter log item by keywords
func KeywordsFilter(kws []string, exclude bool) ItemFilterFunc {
	return func(li *LogItem) bool {
		for _, kw := range kws {
			if strings.Contains(li.Msg, kw) {
				return !exclude
			}
		}

		return exclude
	}
}
