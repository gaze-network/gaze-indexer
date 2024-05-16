package automaxprocs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"github.com/Cleverse/go-utilities/utils"
	"github.com/cockroachdb/errors"
	"github.com/gaze-network/indexer-network/pkg/logger"
	"github.com/gaze-network/indexer-network/pkg/logger/slogx"
	"go.uber.org/automaxprocs/maxprocs"
)

var (
	// undo is the undo function returned by maxprocs.Set
	undo func()

	// autoMaxProcs is the value of GOMAXPROCS set by automaxprocs.
	// will be -1 if `automaxprocs` is not initialized.
	autoMaxProcs = -1

	// initialMaxProcs is the initial value of GOMAXPROCS.
	initialMaxProcs = Current()
)

func Init() error {
	logger := logger.With(
		slogx.String("package", "automaxprocs"),
		slogx.String("event", "set_gomaxprocs"),
		slogx.Int("prev_maxprocs", initialMaxProcs),
	)

	// Create a logger function for `maxprocs.Set`.
	setMaxProcLogger := func(format string, v ...any) {
		fields := make([]slog.Attr, 0, 1)

		// `maxprocs.Set` will always pass current GOMAXPROCS value to logger.
		// except when calling `undo` function, it will not pass any value.
		if val, ok := utils.Optional(v); ok {
			// if `GOMAXPROCS` environment variable is set, then `automaxprocs` will honor it.
			if _, exists := os.LookupEnv("GOMAXPROCS"); exists {
				val = Current()
			}

			// add logging field for `set_maxprocs` value if it's present in integer value.
			if setmaxprocs, ok := val.(int); ok {
				fields = append(fields, slogx.Int("set_maxprocs", setmaxprocs))
			}
		}

		logger.LogAttrs(context.Background(), slog.LevelInfo, fmt.Sprintf(format, v...), fields...)
	}

	// Set GOMAXPROCS to match the Linux container CPU quota (if any), returning
	// any error encountered and an undo function.
	//
	// Set is a no-op on non-Linux systems and in Linux environments without a
	// configured CPU quota.
	revert, err := maxprocs.Set(maxprocs.Logger(setMaxProcLogger), maxprocs.Min(1))
	if err != nil {
		return errors.WithStack(err)
	}

	// set the result of `maxprocs.Set` to global variable.
	autoMaxProcs = Current()
	undo = revert
	return nil
}

// Undo restores GOMAXPROCS to its previous value.
// or revert to initial value if `automaxprocs` is not initialized.
//
// returns the current GOMAXPROCS value.
func Undo() int {
	if undo != nil {
		undo()
		return Current()
	}

	runtime.GOMAXPROCS(initialMaxProcs)
	return initialMaxProcs
}

// Current returns the current value of GOMAXPROCS.
func Current() int {
	return runtime.GOMAXPROCS(0)
}

// Value returns the value of GOMAXPROCS set by automaxprocs.
// returns -1 if `automaxprocs` is not initialized.
func Value() int {
	if autoMaxProcs <= 0 {
		return -1
	}
	return autoMaxProcs
}
