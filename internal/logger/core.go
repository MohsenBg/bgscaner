package logger

var coreLogger *Logger

func InitCore() error {

	var err error
	coreLogger, err = newLogger("core.log")
	return err
}

func Core() *Logger {
	return coreLogger
}

func CoreInfo(msg string, args ...any) {
	coreLogger.write(LevelInfo, msg, args...)
}

func CoreError(msg string, args ...any) {
	coreLogger.write(LevelError, msg, args...)
}
