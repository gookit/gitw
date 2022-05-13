package chlog

import (
	"strings"

	"github.com/gookit/goutil/strutil"
)

// ItemFilter interface
type ItemFilter interface {
	// Handle filtering
	Handle(li *LogItem) bool
}

// ItemFilterFunc define
// type LineFilterFunc func(line string) bool
type ItemFilterFunc func(li *LogItem) bool

// Handle filtering
func (f ItemFilterFunc) Handle(li *LogItem) bool {
	return f(li)
}

// global filters
// var filters = map[string]ItemFilter{
// 	"msgLen": MsgLenFilter,
// }

// MsgLenFilter handler
func MsgLenFilter(minLen int) ItemFilterFunc {
	return func(li *LogItem) bool {
		return len(li.Msg) > minLen
	}
}

// WordsLenFilter handler
func WordsLenFilter(minLen int) ItemFilterFunc {
	return func(li *LogItem) bool {
		return len(strutil.Split(li.Msg, " ")) > minLen
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
