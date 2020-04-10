package carte

import (
	"io"
	"runtime"
	"strings"
	"time"
)

type severity []byte

// Predefined severity levels
var (
	Info     severity = []byte("INFO")
	Debug    severity = []byte("DEBG")
	Warn     severity = []byte("WARN")
	Error    severity = []byte("ERR ")
	Critical severity = []byte("CRIT")
)

// Detail is a struct for providing key-value pairs to a log
// Also considered accepting an blank interface and either checking if it matches a specific interface w func -> string string
// else if it doesnt match just setting the name to the type and the value to json marshalling of the object
type Detail struct {
	Name  string
	Value string
}

// LogOut writes to Settings.writerOut
func LogOut(sev severity, message string, details ...*Detail) (int, error) {
	Settings.mux.Lock()
	writer := Settings.writerOut
	Settings.mux.Unlock()

	return log(writer, sev, message, details...)
}

// UlogOut is a no return writer to Settings.writerOut
func UlogOut(sev severity, message string, details ...*Detail) {
	Settings.mux.Lock()
	writer := Settings.writerOut
	Settings.mux.Unlock()

	_, _ = log(writer, sev, message, details...)
}

// LogErr writes to Settings.writerErr
func LogErr(sev severity, message string, details ...*Detail) (int, error) {
	Settings.mux.Lock()
	writer := Settings.writerErr
	Settings.mux.Unlock()

	return log(writer, sev, message, details...)
}

// UlogErr is a no return writer to Settings.writerErr
func UlogErr(sev severity, message string, details ...*Detail) {
	Settings.mux.Lock()
	writer := Settings.writerErr
	Settings.mux.Unlock()

	_, _ = log(writer, sev, message, details...)
}

// LogTo writes to the provided writer
func LogTo(writer io.Writer, sev severity, message string, details ...*Detail) (int, error) {
	return log(writer, sev, message, details...)
}

// UlogTo is a no return writer to the provided writer
func UlogTo(writer io.Writer, sev severity, message string, details ...*Detail) {
	_, _ = log(writer, sev, message, details...)
}

// Made the log one line to allow for a quick synchronous write to a custom writer with mutex locking
func log(writer io.Writer, sev severity, message string, details ...*Detail) (int, error) {
	// Get date
	Settings.mux.Lock()
	date := []byte(time.Now().In(Settings.location).Format(Settings.dateFormat))
	Settings.mux.Unlock()

	// Get func name
	pc, _, _, ok := runtime.Caller(1)
	Settings.mux.Lock()
	funcName := getCallerName(pc, ok, Settings.functionNameLength)
	Settings.mux.Unlock()

	// Rough estimate of all the required wrappers to a log, NOT the info
	// Reduce the number of allocations
	// TODO: calculate the entire size of the log and use copy or just directly assign the bytes (probably benchmark)
	// Base of 42 + (if len dtls > 0) -> 9 + len dtls * 6
	baseLogLen := 42
	if len(details) > 0 {
		baseLogLen += 9
		baseLogLen += len(details) * 6
	}
	jsonLog := make([]byte, 0, baseLogLen)

	// DATE
	jsonLog = append(jsonLog, []byte(`{"Time":"`)...)
	jsonLog = append(jsonLog, date...)
	//_, _ = writer.Write([]byte(`{"Time":"`))
	//_, _ = writer.Write(date)
	// FUNC
	jsonLog = append(jsonLog, []byte(`","Func":"`)...)
	jsonLog = append(jsonLog, funcName...)
	//_, _ = writer.Write([]byte(`","Func":"`))
	//_, _ = writer.Write(funcName)
	// TYPE
	jsonLog = append(jsonLog, []byte(`","Severity":"`)...)
	jsonLog = append(jsonLog, sev...)
	//_, _ = writer.Write([]byte(`","Type":"`))
	//_, _ = writer.Write(logType.name)
	// MESSAGE
	jsonLog = append(jsonLog, []byte(`","Message":"`)...)
	jsonLog = append(jsonLog, message...)
	jsonLog = append(jsonLog, '"')
	//_, _ = writer.Write([]byte(`","Mess":"`))
	//_, _ = writer.Write([]byte(message))
	//_, _ = writer.Write([]byte(`"`))
	// DETAILS
	if len(details) > 0 {
		jsonLog = append(jsonLog, []byte(`,"Dtls":{`)...)
		//_, _ = writer.Write([]byte(`,"Dtls":{`))
		for i, d := range details {
			jsonLog = append(jsonLog, '"')
			jsonLog = append(jsonLog, d.Name...)
			jsonLog = append(jsonLog, []byte(`":"`)...)
			jsonLog = append(jsonLog, d.Value...)
			jsonLog = append(jsonLog, '"')
			//_, _ = writer.Write([]byte(`"`))
			//_, _ = writer.Write([]byte(d.Name))
			//_, _ = writer.Write([]byte(`":"`))
			//_, _ = writer.Write([]byte(d.Value))
			//_, _ = writer.Write([]byte(`"`))

			// If this is not the last detail append a comma
			if i != len(details)-1 {
				jsonLog = append(jsonLog, []byte(",")...)
				//_, _ = writer.Write([]byte(","))
			}
		}
		jsonLog = append(jsonLog, '}')
		//_, _ = writer.Write([]byte("}"))
	}
	jsonLog = append(jsonLog, []byte("}\n")...)
	//_, _ = writer.Write([]byte("}\n"))

	return writer.Write(jsonLog)
}

func getCallerName(pc uintptr, ok bool, maxLen int) []byte {
	callerName := "unavailable"

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

	if len(callerName) > maxLen {
		return []byte(callerName[:maxLen])
	}

	return []byte(callerName)
}