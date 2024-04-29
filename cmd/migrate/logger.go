package migrate

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
)

var _ migrate.Logger = (*consoleLogger)(nil)

type consoleLogger struct {
	prefix  string
	verbose bool
}

func (l *consoleLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(l.prefix+format, v...)
}

func (l *consoleLogger) Verbose() bool {
	return l.verbose
}
