package workers

type DocumentProcessor interface {
	Work(errChan chan error)
}
