package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	projectDir string
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	projectDir = filepath.Join(path.Dir(filename), "..", "..")
}

func setup() error {
	return nil
}

func teardown() error {
	return nil
}

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		teardown()
		panic(err)
	}

	result := m.Run()

	err = teardown()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	os.Exit(result)
}

func TestRunExample(t *testing.T) {
	cmd := exec.Command("make", "run-example")
	cmd.Dir = projectDir
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Start()
	assert.NoError(t, err)

	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	deadline := 30 * time.Second
	select {
	case <-time.After(deadline):
		err = cmd.Process.Kill()
		if err != nil {
			fmt.Printf("failed to kill the process: %v\n", err)
		}
		t.Fatalf("failed to complete the example within %s", deadline.String())
	case err = <-done:
		if err != nil {
			t.Fatalf("failed to execute the process: %s", stderr.String())
		}
		break
	}

	output := stdout.String()
	assert.Contains(t, output, "Queue 0")
}
