package carte

import (
	"errors"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Errors
var (
	ErrLocationWasNil = errors.New("received a nil time.Location")
)

// Settings
var (
	// -1 returns the entire function name
	// 0  excludes it from the log
	// >0 limits the length
	// TODO: maybe have 0 not set the function name
	functionNameLength       = -1
	EntireFunctionNameLength = -1

	timezone   = time.UTC
	dateFormat = "2006-01-02T15:04:05 MST"

	mux sync.Mutex
)

func SetFunctionNameLength(fnl int) {
	mux.Lock()
	functionNameLength = fnl
	mux.Unlock()
}

func ExcludeFunctionName() {
	mux.Lock()
	functionNameLength = 0
	mux.Unlock()
}

func SetTimezone(tz *time.Location) error {
	if tz == nil {
		return ErrLocationWasNil
	}

	mux.Lock()
	timezone = tz
	mux.Unlock()

	return nil
}

func SetDateFormat(format string) {
	mux.Lock()
	dateFormat = format
	mux.Unlock()
}

// TODO: Consider adding file name and line number
func getFuncName() []byte {
	mux.Lock()
	defer mux.Unlock()

	if functionNameLength == 0 {
		return nil
	}

	callerName := "unavailable"

	// Skip = 3
	// This is called by log
	// log is called by a public Log func
	pc, _, _, ok := runtime.Caller(3)

	if ok {
		callerFunc := runtime.FuncForPC(pc)
		if callerFunc != nil {
			callerName = callerFunc.Name()
			fileNameSeparator := strings.Index(callerName, ".")
			if fileNameSeparator != -1 {
				callerName = callerName[fileNameSeparator+1:]
			}
		}
	}

	// Return the entire function name
	if functionNameLength <= 0 {
		return []byte(callerName)
	}

	// Return a substring of the function name
	if len(callerName) > functionNameLength {
		return []byte(callerName[:functionNameLength])
	}

	return []byte(callerName)
}

func getDate() []byte {
	mux.Lock()
	defer mux.Unlock()

	return []byte(time.Now().In(timezone).Format(dateFormat))
}
