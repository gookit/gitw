package chlog

import "strings"

type LineFilterFunc func(line string) bool
type ItemFilterFunc func(li *LogItem) bool

// KeywordFilter filter log item by keyword
func KeywordFilter(kw string, exclude bool) ItemFilterFunc{
	return func(li *LogItem) bool {
		has := strings.Contains(li.Msg, kw)

		if exclude {
			return !has
		}
		return has
	}
}

// KeywordsFilter filter log item by keywords
func KeywordsFilter(kws []string, exclude bool) ItemFilterFunc{
	return func(li *LogItem) bool {
		for _, kw := range kws {
			if strings.Contains(li.Msg, kw) {
				return !exclude
			}
		}

		return exclude
	}
}
