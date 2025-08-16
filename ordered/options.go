package ordered

type Option func(*option)

type option struct {
	capacity int
}

func WithCapacity(capacity int) Option {
	return func(o *option) {
		o.capacity = capacity
	}
}
