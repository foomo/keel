package keeltestutil

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/foomo/keel/log"
)

func Inline(t *testing.T, actual interface{}, skip int) (string, bool) {
	t.Helper()

	// retrieve caller info
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		t.Fatal("failed to retrieve caller")
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
	}

	fileLines[line-1] = fmt.Sprintf("%s // INLINE: %v", fileLine, actual)
	if err := os.WriteFile(file, []byte(strings.Join(fileLines, "\n")), 0644); err != nil {
		t.Fatal("failed to write caller file", log.FError(err))
	}

	t.Errorf("Inline annotaion created for %s:%d", file, line)
	return "", false
}
