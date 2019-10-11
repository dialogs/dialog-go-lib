package logger

type nopWriter struct {
}

func newNopWriter() *nopWriter {
	return &nopWriter{}
}

func (w *nopWriter) Write([]byte) (int, error) {
	return 0, nil
}
