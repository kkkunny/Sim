package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kkkunny/Sim/compiler/config"
)

var simBin string
var repoRoot string

func TestMain(m *testing.M) {
	_, file, _, _ := runtime.Caller(0)
	repoRoot = filepath.Dir(filepath.Dir(file))

	tmp, err := os.MkdirTemp("", "sim-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)

	simBin = filepath.Join(tmp, "sim")
	cmd := exec.Command("go", "build", "-tags", "compile", "-o", simBin, ".")
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestSim(t *testing.T) {
	runCases(t, "success", true)
	runCases(t, "failed", false)
}

func runCases(t *testing.T, dir string, expectSuccess bool) {
	matches, err := filepath.Glob(filepath.Join(dir, "*"+config.SourceCodeFileExtName))
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range matches {
		src := src
		name := strings.TrimSuffix(filepath.Base(src), config.SourceCodeFileExtName)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			outBin := filepath.Join(t.TempDir(), "main.out")
			absSrc, err := filepath.Abs(src)
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(simBin, absSrc)
			cmd.Dir = repoRoot
			cmd.Env = append(os.Environ(), "SIM_OUTPUT="+outBin)
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if expectSuccess {
				if err != nil {
					t.Fatalf("compile failed: %v", err)
				}
			}

			_, err = exec.Command(outBin).Output()
			if expectSuccess {
				if err != nil {
					t.Fatalf("run failed: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected run failure but succeeded")
				}
			}
		})
	}
}
