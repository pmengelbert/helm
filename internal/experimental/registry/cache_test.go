package registry

import (
	"crypto/rand"
	"fmt"
	"helm.sh/helm/v3/pkg/chart"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestChartListWillNotCrash(t *testing.T) {
	tmpDir, err := ioutil.TempDir("/home/pme/.cache/helm/", "helm-chart-list-test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpDir)

	cache, err := NewCache(CacheOptRoot(tmpDir))
	if err != nil {
		t.Error(err)
	}

	client, err := NewClient(
		ClientOptWriter(os.Stderr),
		ClientOptCache(cache),
	)
	if err != nil {
		t.Error(err)
	}

	data := make([]byte, 96000)
	rand.Read(data)

	numCharts := 5000
	for i := 0; i < numCharts; i++ {
		ref, err := ParseReference(fmt.Sprintf("localhost:5000/chart%d:latest", i))
		if err != nil {
			t.Error(err)
		}

		s := fmt.Sprintf("%d", i)
		d := append(data, []byte(s)...)

		ch := &chart.Chart{
			Metadata: &chart.Metadata{
				APIVersion: chart.APIVersionV1,
				Name:       s,
				Version:    "1.2.3",
			}, Files: []*chart.File{
				{Name: "scheherazade/shahryar" + s + ".txt", Data: d},
			}, Templates: []*chart.File{
				{Name: filepath.Join(tmpDir, "nested", "dir", "thing"+s+".yaml"), Data: []byte(`abc123`)},
			},
		}

		err = client.SaveChart(ch, ref)
		if err != nil {
			t.Error(err)
		}
	}

	err = client.PrintChartTable()
	if err != nil {
		t.Fatal(err)
	}
}
