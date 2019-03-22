package consumer

type Logger interface {
	Info(a ...interface{})
	Infof(format string, a ...interface{})
	Error(a ...interface{})
	Errorf(format string, a ...interface{})
	Warn(a ...interface{})
	Warnf(format string, a ...interface{})
	Debug(a ...interface{})
	Debugf(format string, a ...interface{})
}

type NopLogger struct {
}

func (n *NopLogger) Info(a ...interface{}) {

}

func (n *NopLogger) Infof(format string, a ...interface{}) {

}

func (n *NopLogger) Error(a ...interface{}) {

}

func (n *NopLogger) Errorf(format string, a ...interface{}) {

}

func (n *NopLogger) Warn(a ...interface{}) {

}

func (n *NopLogger) Warnf(format string, a ...interface{}) {

}

func (n *NopLogger) Debug(a ...interface{}) {

}

func (n *NopLogger) Debugf(format string, a ...interface{}) {

}
