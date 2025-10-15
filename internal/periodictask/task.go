package periodictask

// Task defines a generic task which can be run
type Task interface {
	Run() error
}
