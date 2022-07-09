package middlewares

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

// Recoverer middleware uses builtin recover function
func Recoverer(withLoggingService bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if rvr == http.ErrAbortHandler {
						// we don't recover http.ErrAbortHandler so the response
						// to the client is aborted, this should not be logged
						panic(rvr)
					}

					printPrettyStack(rvr) // print the debug stack
					if withLoggingService {
						//callLoggerService(r)
					}

					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

// for ability to test the PrintPrettyStack function
var recovererErrorWriter io.Writer = os.Stderr

//func callLoggerService(req *http.Request) {
//	b := &bytes.Buffer{}
//	_, _ = io.Copy(b, req.Body)
//	defer req.Body.Close()
//
//	logReq := &service.LoggingApiRequest{
//		Method:    req.Method,
//		Path:      req.URL.Path,
//		PostBody:  b.String(),
//		Origin:    req.Header.Get("origin"),
//		UserAgent: req.UserAgent(),
//	}
//	go grpcClient.Logging.Api(context.Background(), logReq)
//}

func printPrettyStack(rvr any) {
	debugStack := debug.Stack()
	s := prettyStack{}
	out, err := s.parse(debugStack, rvr)
	if err == nil {
		recovererErrorWriter.Write(out)
	} else {
		// print stdlib output as a fallback
		os.Stderr.Write(debugStack)
	}
}

func consoleWrite(w io.Writer, s string, args ...any) {
	fmt.Fprintf(w, s, args...)
}

type prettyStack struct {
}

func (s prettyStack) parse(debugStack []byte, rvr any) ([]byte, error) {
	var err error
	buf := &bytes.Buffer{}

	consoleWrite(buf, "\n")
	consoleWrite(buf, " panic: ")
	consoleWrite(buf, "%v", rvr)
	consoleWrite(buf, "\n \n")

	// process debug stack info
	stack := strings.Split(string(debugStack), "\n")
	lines := []string{}

	// locate panic line, as we may have nested panics
	for i := len(stack) - 1; i > 0; i-- {
		lines = append(lines, stack[i])
		if strings.HasPrefix(stack[i], "panic(") {
			lines = lines[0 : len(lines)-2] // remove boilerplate
			break
		}
	}

	// reverse
	for i := len(lines)/2 - 1; i >= 0; i-- {
		opp := len(lines) - 1 - i
		lines[i], lines[opp] = lines[opp], lines[i]
	}

	// decorate
	for i, line := range lines {
		lines[i], err = s.decorateLine(line, i)
		if err != nil {
			return nil, err
		}
	}

	for _, l := range lines {
		fmt.Fprintf(buf, "%s", l)
	}
	return buf.Bytes(), nil
}

func (s prettyStack) decorateLine(line string, num int) (string, error) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "\t") || strings.Contains(line, ".go:") {
		return s.decorateSourceLine(line, num)
	} else if strings.HasSuffix(line, ")") {
		return s.decorateFuncCallLine(line, num)
	} else {
		if strings.HasPrefix(line, "\t") {
			return strings.Replace(line, "\t", "      ", 1), nil
		} else {
			return fmt.Sprintf("    %s\n", line), nil
		}
	}
}

func (s prettyStack) decorateFuncCallLine(line string, num int) (string, error) {
	idx := strings.LastIndex(line, "(")
	if idx < 0 {
		return "", errors.New("not a func call line")
	}

	buf := &bytes.Buffer{}
	pkg := line[0:idx]
	// addr := line[idx:]
	method := ""

	if idx := strings.LastIndex(pkg, string(os.PathSeparator)); idx < 0 {
		if idx := strings.Index(pkg, "."); idx > 0 {
			method = pkg[idx:]
			pkg = pkg[0:idx]
		}
	} else {
		method = pkg[idx+1:]
		pkg = pkg[0 : idx+1]
		if idx := strings.Index(method, "."); idx > 0 {
			pkg += method[0:idx]
			method = method[idx:]
		}
	}

	if num == 0 {
		consoleWrite(buf, " -> ")
	} else {
		consoleWrite(buf, "    ")
	}
	consoleWrite(buf, "%s", pkg)
	consoleWrite(buf, "%s\n", method)
	return buf.String(), nil
}

func (s prettyStack) decorateSourceLine(line string, num int) (string, error) {
	idx := strings.LastIndex(line, ".go:")
	if idx < 0 {
		return "", errors.New("not a source line")
	}

	buf := &bytes.Buffer{}
	path := line[0 : idx+3]
	lineno := line[idx+3:]

	idx = strings.LastIndex(path, string(os.PathSeparator))
	dir := path[0 : idx+1]
	file := path[idx+1:]

	idx = strings.Index(lineno, " ")
	if idx > 0 {
		lineno = lineno[0:idx]
	}

	if num == 1 {
		consoleWrite(buf, " ->   ")
	} else {
		consoleWrite(buf, "      ")
	}
	consoleWrite(buf, "%s", dir)
	consoleWrite(buf, "%s", file)
	consoleWrite(buf, "%s", lineno)
	if num == 1 {
		consoleWrite(buf, "\n")
	}
	consoleWrite(buf, "\n")

	return buf.String(), nil
}
