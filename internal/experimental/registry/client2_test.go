/*
Copyright The Helm Authors.

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

package registry

import (
	"encoding/json"
	"fmt"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	s *httptest.Server
	rc *Client
	cache *Cache
	resolver *Resolver
)

func TestStuff(t *testing.T) {
	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log(r.URL.Path)
		body := []byte("nobody wanted this")
		digest := digest.FromBytes(body)
		_ = body
		m := ocispec.Manifest{
			Versioned:   specs.Versioned{},
			Config:      ocispec.Descriptor{},
			Layers:      []ocispec.Descriptor{ocispec.Descriptor{
				MediaType:   "application/tar+gzip",
				Digest:      digest,
			}},
			Annotations: nil,
		}
		w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
		w.WriteHeader(200)
		data, _ := json.Marshal(&m)
		w.Write(data)
	}))

	tmpDir, err := ioutil.TempDir("", "helm-pull-digest-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cache, err := NewCache(
		CacheOptDebug(true),
		CacheOptWriter(os.Stdout),
		CacheOptRoot(filepath.Join(tmpDir, CacheRootDir)),
	)

	rc, err = NewClient(
		ClientOptDebug(true),
		ClientOptWriter(os.Stdout),
		ClientOptCache(cache),
	)

	url := "localhost" + strings.TrimPrefix(s.URL, "http://127.0.0.1")
	ref, err := ParseReference(fmt.Sprintf("%s/testrepo/whodis:9.9.9", url))
	if err != nil {
		t.Fatal(err)
	}

	err = rc.PullChart(ref)
	if err != nil {
		t.Fatal(err)
	}
}

