//go:build mage
// +build mage

/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"path/filepath"

	"github.com/magefile/mage/sh"
	"github.com/uwu-tools/magex/pkg"

	"sigs.k8s.io/release-utils/mage"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var (
	Default = Verify
)

const (
	binDir    = "bin"
	scriptDir = "scripts"
)

var boilerplateDir = filepath.Join(scriptDir, "boilerplate")

// All runs all targets for this repository
func All() error {
	if err := Verify(); err != nil {
		return err
	}

	if err := Test(); err != nil {
		return err
	}

	return nil
}

// Test runs various test functions
func Test() error {
	if err := mage.TestGo(true); err != nil {
		return err
	}

	return nil
}

// IntegrationTest runs the integration test functions
func IntegrationTest() error {
	if err := mage.EnsureGitConfig(); err != nil {
		return err
	}

	if err := mage.TestGoWithTags(true, "integration", "test/integration"); err != nil {
		return err
	}

	return nil
}

// E2ETest runs the end-to-end test functions.
func E2ETest() error {
	if err := mage.EnsureGitConfig(); err != nil {
		return err
	}

	if err := mage.TestGoWithTags(true, "e2e", "test/e2e"); err != nil {
		return err
	}

	return nil
}

// LocalE2ETest runs the end-to-end test functions locally.
func LocalE2ETest() error {
	// Setup local registry to push images to
	if err := sh.Run("docker", "run", "--name", "reg", "-d", "-p", "5000:5000", "registry:2"); err != nil {
		return err
	}

	return E2ETest()
}

// Verify runs repository verification scripts
func Verify() error {
	fmt.Println("Ensuring mage is available...")
	if err := pkg.EnsureMage(""); err != nil {
		return err
	}

	fmt.Println("Running copyright header checks...")
	if err := mage.VerifyBoilerplate("v0.2.5", binDir, boilerplateDir, false); err != nil {
		return err
	}

	fmt.Println("Running external dependency checks...")
	if err := mage.VerifyDeps("v0.3.0", "", "", true); err != nil {
		return err
	}

	fmt.Println("Running go module linter...")
	if err := mage.VerifyGoMod(); err != nil {
		return err
	}

	fmt.Println("Running golangci-lint...")
	if err := mage.RunGolangCILint("", false); err != nil {
		return err
	}

	fmt.Println("Running go build...")
	if err := mage.VerifyBuild(scriptDir); err != nil {
		return err
	}

	return nil
}

func Clean() {
	fmt.Println("Cleaning workspace...")
	toClean := []string{"bin", "test/e2e/release-sdk-testkey.key", "test/e2e/release-sdk-testkey.pub"}

	for _, clean := range toClean {
		sh.Rm(clean)
	}

	fmt.Println("Done.")
}
