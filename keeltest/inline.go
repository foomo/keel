package keeltest

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/foomo/keel/log"
)

// Inline returns the expected value stored as a trailing " // INLINE: "
// comment on the source line of the caller skip frames up (1 is the caller
// of Inline), and true.
//
// If the line has no such comment, msgAndArgs is formatted with
// [fmt.Sprintf], appended to the line as an INLINE comment in the source
// file, the test is marked as failed, and Inline returns "" and false.
// Subsequent runs then compare against the written value. Missing
// msgAndArgs or file access errors abort the test with t.Fatal.
func Inline(t *testing.T, skip int, msgAndArgs ...any) (string, bool) {
	t.Helper()

	// retrieve caller info
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		t.Fatal("failed to retrieve caller")
	}

	fileStat, err := os.Stat(file)
	if err != nil {
		t.Fatal("failed to stat caller file", log.FError(err))
	}

	// read file
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		t.Fatal("failed to read caller file", log.FError(err))
	}

	fileLines := strings.Split(string(fileBytes), "\n")
	fileLine := fileLines[line-1]
	fileLineParts := strings.Split(strings.TrimSpace(fileLine), " // INLINE: ")

	// compare
	if len(fileLineParts) == 2 {
		return fileLineParts[1], true
	} else if len(msgAndArgs) == 0 {
		t.Fatal("missing inline message")
	} else if msg, ok := msgAndArgs[0].(string); !ok {
		t.Fatal("invalid inline message")
	} else if value := fmt.Sprintf(msg, msgAndArgs[1:]...); len(value) == 0 {
		t.Fatal("missing inline message")
	} else {
		fileLines[line-1] = fmt.Sprintf("%s // INLINE: %s", fileLine, value)
		if err := os.WriteFile(file, []byte(strings.Join(fileLines, "\n")), fileStat.Mode().Perm()); err != nil { //nolint:gosec
			t.Fatal("failed to write inline", log.FError(err))
		}

		t.Errorf("wrote inline for %s:%d", file, line)
	}

	return "", false
}

// InlineInt is like [Inline] but parses the stored value as an int. It does
// not write missing values and aborts the test if none is stored.
func InlineInt(t *testing.T, skip int) (int, bool) {
	t.Helper()

	if inline, ok := Inline(t, skip+1); !ok {
		return 0, false
	} else if value, err := strconv.Atoi(inline); err != nil {
		t.Fatal("failed to parse int", log.FError(err))
		return 0, false
	} else {
		return value, true
	}
}

// InlineFloat64 is like [Inline] but parses the stored value as a float64.
// It does not write missing values and aborts the test if none is stored.
func InlineFloat64(t *testing.T, skip int) (float64, bool) {
	t.Helper()

	if inline, ok := Inline(t, skip+1); !ok {
		return 0, false
	} else if value, err := strconv.ParseFloat(inline, 64); err != nil {
		t.Fatal("failed to parse int", log.FError(err))
		return 0, false
	} else {
		return value, true
	}
}

// InlineJSON is like [Inline] but unmarshals the stored JSON value into
// target. It does not write missing values and aborts the test if none is
// stored.
func InlineJSON(t *testing.T, skip int, target any) {
	t.Helper()

	if inline, ok := Inline(t, skip+1); ok {
		if err := json.Unmarshal([]byte(inline), target); err != nil {
			t.Fatal("failed to unmarshal json", log.FError(err))
		}
	}
}
