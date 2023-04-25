package logs

type LevelProvider interface {
	Search(loggerName string) Level
}
