package logs

type LevelProvider interface {
	Search(pkg string) Level
}
