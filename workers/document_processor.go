package workers

import "os"

type DocumentProcessor interface {
	Work() error
	SetupSignalHandler() chan os.Signal
	WaitForShutdown(errChan chan error, signalChan chan os.Signal)
}
