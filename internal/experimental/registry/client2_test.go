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
	"encoding/base64"
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
	"regexp"
	"strings"
	"testing"
)

const charty = "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"

var (
	s *httptest.Server
	rc *Client
	cache *Cache
	resolver *Resolver
)

func TestStuff(t *testing.T) {
	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log(r.URL.Path)
		body, err := base64.StdEncoding.DecodeString(charty)
		if err != nil {
			t.Error(err)
		}

		if strings.Contains(r.URL.Path, "manifests") {
			digest := digest.FromBytes(body)
			_ = body
			m := ocispec.Manifest{
				Versioned: specs.Versioned{},
				Config:    ocispec.Descriptor{},
				Layers: []ocispec.Descriptor{ocispec.Descriptor{
					MediaType: "application/tar+gzip",
					Digest:    digest,
					Size:      int64(len(body)),
				}},
				Annotations: nil,
			}
			w.Header().Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
			w.WriteHeader(200)
			data, _ := json.Marshal(&m)
			w.Write(data)
		}

		if strings.Contains(r.URL.Path, "blobs") {
			w.Header().Set("Content-Type", "application/tar+gzip")
			w.WriteHeader(200)
			b, _ := base64.StdEncoding.DecodeString("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")
			w.Write(b[:3279])
		}
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


	re := regexp.MustCompile(`https?://[^:]*(:[0-9]+)`)
	portString := re.ReplaceAllString(s.URL, "$1")
	ref, err := ParseReference(fmt.Sprintf("localhost%s/testrepo/whodis:9.9.9", portString))
	if err != nil {
		t.Fatal(err)
	}

	err = rc.PullChart(ref)
	if err != nil {
		t.Fatal(err)
	}
}

