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

package getter

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/internal/experimental/registry"
	"helm.sh/helm/v3/pkg/cli"
	"net/http"
	"os"
	"strings"

	"helm.sh/helm/v3/internal/tlsutil"
	"helm.sh/helm/v3/internal/urlutil"
)

// OCIGetter is the default HTTP(/S) backend handler
type OCIGetter struct {
	opts options
}

//Get performs a Get from repo.Getter and returns the body.
func (g *OCIGetter) Get(href string, options ...Option) (*bytes.Buffer, error) {
	for _, opt := range options {
		opt(&g.opts)
	}
	return g.get(href)
}

func (g *OCIGetter) get(href string) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	settings := cli.New()

	client, err := registry.NewClient(
		registry.ClientOptDebug(settings.Debug),
		registry.ClientOptWriter(os.Stdout),
		registry.ClientOptCredentialsFile(settings.RegistryConfig),
	)
	if err != nil {
		return nil, err
	}

	ref := strings.TrimPrefix(href, "oci://")
	if tag := g.opts.tagname; tag != "" {
		ref = fmt.Sprintf("%s:%s", ref, tag)
	}

	r, err := registry.ParseReference(ref)
	if err != nil {
		return nil, err
	}

	buf, err = client.PullChart2(r)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// NewOCIGetter constructs a valid http/https client as a Getter
func NewOCIGetter(options ...Option) (Getter, error) {
	var client OCIGetter

	for _, opt := range options {
		opt(&client.opts)
	}

	return &client, nil
}

func (g *OCIGetter) httpClient() (*http.Client, error) {
	transport := &http.Transport{
		DisableCompression: true,
		Proxy:              http.ProxyFromEnvironment,
	}
	if (g.opts.certFile != "" && g.opts.keyFile != "") || g.opts.caFile != "" {
		tlsConf, err := tlsutil.NewClientTLS(g.opts.certFile, g.opts.keyFile, g.opts.caFile)
		if err != nil {
			return nil, errors.Wrap(err, "can't create TLS config for client")
		}
		tlsConf.BuildNameToCertificate()

		sni, err := urlutil.ExtractHostname(g.opts.url)
		if err != nil {
			return nil, err
		}
		tlsConf.ServerName = sni

		transport.TLSClientConfig = tlsConf
	}

	if g.opts.insecureSkipVerifyTLS {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

	}

	client := &http.Client{
		Transport: transport,
		Timeout:   g.opts.timeout,
	}

	return client, nil
}
