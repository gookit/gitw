package chlog

type Formatter interface {
	MatchGroup(msg string) (group string)
	Format(li *LogItem) (group, fmtLine string)
}
