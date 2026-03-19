package logger

var uiLogger *Logger

func InitUI() error {

	var err error
	uiLogger, err = newLogger("ui.log")
	return err
}

func UI() *Logger {
	return uiLogger
}

func UIInfo(msg string, args ...any) {
	uiLogger.write(LevelInfo, msg, args...)
}

func UIError(msg string, args ...any) {
	uiLogger.write(LevelError, msg, args...)
}
