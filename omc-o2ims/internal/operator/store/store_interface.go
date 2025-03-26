package store

type Store[T any] interface {
	List() ([]T, error)
	ReconcileList() error
	UpdateStatus(string, map[string]interface{}) error
}
