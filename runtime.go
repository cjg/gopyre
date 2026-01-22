package gopyre

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ebitengine/purego"
)

type pyRuntime struct {
	lib uintptr
	fns pyFns
}

var (
	runtimeOnce sync.Once
	runtimeInst *pyRuntime
	runtimeErr  error
)

func getRuntime() (*pyRuntime, error) {
	runtimeOnce.Do(func() {
		runtimeInst, runtimeErr = initRuntime()
	})
	return runtimeInst, runtimeErr
}

func initRuntime() (*pyRuntime, error) {
	lib, err := openLibPython()
	if err != nil {
		return nil, err
	}

	rt := &pyRuntime{lib: lib}
	rt.bind()

	if rt.fns.Py_IsInitialized() == 0 {
		rt.fns.Py_Initialize()
		// rt.fns.PyEval_InitThreads()
	}

	return rt, nil
}

func openLibPython() (uintptr, error) {
	if path := os.Getenv("GOPYRE_LIBPYTHON"); path != "" {
		lib, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			return 0, fmt.Errorf("gopyre: dlopen %q: %w", path, err)
		}
		return lib, nil
	}

	var lastErr error
	for _, candidate := range libPythonCandidates(runtime.GOOS) {
		lib, err := purego.Dlopen(candidate, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err == nil {
			return lib, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no libpython candidates for %s", runtime.GOOS)
	}
	return 0, fmt.Errorf("gopyre: unable to load libpython: %w", lastErr)
}

func findLibPythonPathFromPythonExec() (string, error) {
	pythonExecutables := []string{"python3", "python"}
	var pythonExecutable string
	for _, exe := range pythonExecutables {
		pythonPath, err := exec.LookPath(exe)
		if err == nil {
			pythonExecutable = pythonPath
			break
		}
	}
	if pythonExecutable == "" {
		return "", fmt.Errorf("gopyre: no python executable found in PATH")
	}

	output, err := exec.Command(pythonExecutable, "-c", "import sysconfig; print(sysconfig.get_config_var('LIBDIR'))").Output()
	if err != nil {
		return "", fmt.Errorf("gopyre: unable to get libdir from python: %w", err)
	}
	libdir := strings.TrimSpace(string(output))

	output, err = exec.Command(pythonExecutable, "-c", "import sysconfig; print(sysconfig.get_config_var('VERSION'))").Output()
	if err != nil {
		return "", fmt.Errorf("gopyre: unable to get version from python: %w", err)
	}
	version := strings.TrimSpace(string(output))

	var libName string
	if runtime.GOOS == "darwin" {
		libName = fmt.Sprintf("libpython%s.dylib", version)
	} else {
		libName = fmt.Sprintf("libpython%s.so", version)
	}

	return filepath.Join(libdir, libName), nil
}

func libPythonCandidates(goos string) []string {
	var candidates []string

	switch goos {
	case "darwin":
		candidates = []string{
			"libpython3.12.dylib",
			"libpython3.11.dylib",
			"libpython3.10.dylib",
			"libpython3.9.dylib",
		}
	default:
		candidates = []string{
			"libpython3.12.so",
			"libpython3.11.so",
			"libpython3.10.so",
			"libpython3.9.so",
		}
	}

	if path, err := findLibPythonPathFromPythonExec(); err == nil {
		candidates = append([]string{path}, candidates...)
	}

	return candidates
}
