package logger

func CloseAll() {

	if coreLogger != nil {
		coreLogger.Close()
	}

	if debugLogger != nil {
		debugLogger.Close()
	}

	if uiLogger != nil {
		uiLogger.Close()
	}
}
