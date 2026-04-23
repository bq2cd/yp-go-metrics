// Package option implements "functional options" pattern with generics.
package option

// Option defines a function that takes a pointer to an object of the type `T`
// and possibly alters the object internals (typically for configuration purposes).
type Option[T any] func(*T)

// Apply take the target object of type `T` and applies provided options to it
// (that is, calling [Option] function and passing the object to it).
func Apply[T any](obj *T, opts ...Option[T]) {
	for _, opt := range opts {
		opt(obj)
	}
}
