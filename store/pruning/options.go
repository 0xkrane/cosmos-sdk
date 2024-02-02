package pruning

// Options defines the pruning configuration.
type Options struct {
	// KeepRecent sets the number of recent versions to keep.
	KeepRecent uint64

	// Interval sets the number of how often to prune.
	// If set to 0, no pruning will be done.
	Interval uint64

	// Sync when set to true ensure that pruning will be performed
	// synchronously, otherwise by default it will be done asynchronously.
	Sync bool
}

func (o Options) ShouldPrune(height uint64) bool {
	if o.Interval == 0 {
		return false
	}
	if height > o.KeepRecent && height%o.Interval == 0 {
		return true
	}
	return false
}

// DefaultOptions returns the default pruning options.
// Interval is set to 0, which means no pruning will be done.
func DefaultOptions() Options {
	return Options{
		KeepRecent: 0,
		Interval:   0,
		Sync:       false,
	}
}
