package workers

type DocumentProcessor interface {
	Work() error
}
