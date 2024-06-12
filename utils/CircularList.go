package utils

// Circular array or list
type CircularList[T interface{}] struct {
	Data     []T `json:"data"`
	Capacity int `json:"capacity"`
	Head     int `json:"head"`
	// could even store cycles or store last array entirely in a file and then begin overriding...
}

func NewCircularList[T interface{}](capacity int) *CircularList[T] {
	return &CircularList[T]{
		Data:     make([]T, capacity),
		Capacity: capacity,
		Head:     0,
	}
}

// Adds if there's enough space, overwrites if no enough space
func (l *CircularList[T]) OverwriteNext(x T) bool {
	l.Data[l.Head%l.Capacity] = x
	l.Head += 1
	return true
}

// Get as a clean list
func (l *CircularList[T]) GetAsCleanList() []T {
	if len(l.Data) == 0 {
		return nil
	}
	if len(l.Data) == 1 {
		return l.Data
	}
	return append(l.Data[l.Head%l.Capacity:], l.Data[:l.Head%l.Capacity]...)
}
