package metric

type IObserver interface {
	// Observe adds a single observation to the histogram.
	Observe(float64)
}

type ICounter interface {
	// Inc increments the counter by 1. non-negative values.
	Inc()

	// Add adds the given value to the counter. It panics if the value is < 0.
	Add(val float64)
}
