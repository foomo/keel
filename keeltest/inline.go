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

func Inline(t *testing.T, skip int, msgAndArgs ...interface{}) (string, bool) {
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
		if err := os.WriteFile(file, []byte(strings.Join(fileLines, "\n")), fileStat.Mode().Perm()); err != nil {
			t.Fatal("failed to write inline", log.FError(err))
		}

		t.Errorf("wrote inline for %s:%d", file, line)
	}

	return "", false
}

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

func InlineJSON(t *testing.T, skip int, target interface{}) {
	t.Helper()

	if inline, ok := Inline(t, skip+1); ok {
		if err := json.Unmarshal([]byte(inline), target); err != nil {
			t.Fatal("failed to unmarshal json", log.FError(err))
		}
	}
}
