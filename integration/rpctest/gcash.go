// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpctest

import (
	"fmt"
	"go/build"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	// compileMtx guards access to the executable path so that the project is
	// only compiled once.
	compileMtx sync.Mutex

	// executablePath is the path to the compiled executable. This is the empty
	// string until gcash is compiled. This should not be accessed directly;
	// instead use the function gcashExecutablePath().
	executablePath string
)

// gcashExecutablePath returns a path to the gcash executable to be used by
// rpctests. To ensure the code tests against the most up-to-date version of
// gcash, this method compiles gcash the first time it is called. After that, the
// generated binary is used for subsequent test harnesses. The executable file
// is not cleaned up, but since it lives at a static path in a temp directory,
// it is not a big deal.
func gcashExecutablePath() (string, error) {
	compileMtx.Lock()
	defer compileMtx.Unlock()

	// If gcash has already been compiled, just use that.
	if len(executablePath) != 0 {
		return executablePath, nil
	}

	testDir, err := baseDir()
	if err != nil {
		return "", err
	}

	// Determine import path of this package. Not necessarily btcsuite/gcash if
	// this is a forked repo.
	_, rpctestDir, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("Cannot get path to gcash source code")
	}
	gcashPkgPath := filepath.Join(rpctestDir, "..", "..", "..")
	gcashPkg, err := build.ImportDir(gcashPkgPath, build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("Failed to build gcash: %v", err)
	}

	// Build gcash and output an executable in a static temp path.
	outputPath := filepath.Join(testDir, "gcash")
	if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", outputPath, gcashPkg.ImportPath)
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Failed to build gcash: %v", err)
	}

	// Save executable path so future calls do not recompile.
	executablePath = outputPath
	return executablePath, nil
}
