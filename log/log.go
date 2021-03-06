package log

import (
	"github.com/RichardKnop/logging"
)

var (
	logger = logging.New(nil, nil, new(logging.ColouredFormatter))

	// INFO ...
	INFO = logger[logging.INFO]
	// WARNING ...
	WARNING = logger[logging.WARNING]
	// ERROR ...
	ERROR = logger[logging.ERROR]
	// FATAL ...
	FATAL = logger[logging.FATAL]
)

// Set sets a custom logger
func Set(l logging.Logger) {
	logger = l
	INFO = logger[logging.INFO]
	WARNING = logger[logging.WARNING]
	ERROR = logger[logging.ERROR]
	FATAL = logger[logging.FATAL]
}
