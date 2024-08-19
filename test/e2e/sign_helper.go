//go:build e2e
// +build e2e

/*
Copyright 2022 The Kubernetes Authors.

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

package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/publish"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/stretchr/testify/assert"
)

const (
	baseImage  = "cgr.dev/chainguard/static:latest"
	importpath = "golang.org/x/example/hello"
	targetRepo = "localhost:5000"
)

func buildAndPushTestImage(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	b, err := build.NewGo(ctx, ".",
		build.WithPlatforms("linux/amd64"),
		build.WithBaseImages(func(ctx context.Context, _ string) (name.Reference, build.Result, error) {
			ref := name.MustParseReference(baseImage)
			base, err := remote.Index(ref, remote.WithContext(ctx))
			return ref, base, err
		}))
	if err != nil {
		t.Fatalf("NewGo: %v", err)
	}
	r, err := b.Build(ctx, importpath)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	tag := fmt.Sprintf("%s-%d", "test", time.Now().Unix())

	p, err := publish.NewDefault(targetRepo,
		publish.WithTags([]string{string(tag)}),
		publish.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		t.Fatalf("NewDefault: %v", err)
	}
	ref, err := p.Publish(ctx, r, importpath)
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	return ref.String()
}

func genSignKeyPair(t *testing.T) {
	t.Helper()

	filename := "release-sdk-testkey"

	// Check if local signing key already exists
	if _, err := os.Stat(filename + ".key"); err == nil {
		return
	}

	keysBytes, err := cosign.GenerateKeyPair(nil)
	assert.Nil(t, err)

	if err := os.WriteFile(filename+".key", keysBytes.PrivateBytes, 0700); err != nil {
		t.Error(err)
	}

	if err := os.WriteFile(filename+".pub", keysBytes.PublicBytes, 0755); err != nil {
		t.Error(err)
	}
}
