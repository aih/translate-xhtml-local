package integration

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	locations = []struct {
		Lat, Lon float64
		Name     string
	}{
		{33.448, -112.073, "phoenix_az"},
		{40.712, -74.006, "new_york_ny"},
		{34.052, -118.243, "los_angeles_ca"},
		{41.878, -87.629, "chicago_il"},
		{29.760, -95.369, "houston_tx"},
		{39.952, -75.165, "philadelphia_pa"},
		{29.424, -98.493, "san_antonio_tx"},
		{32.715, -117.161, "san_diego_ca"},
		{32.776, -96.797, "dallas_tx"},
		{37.338, -121.886, "san_jose_ca"},
		{30.267, -97.743, "austin_tx"},
		{30.332, -81.655, "jacksonville_fl"},
		{37.774, -122.419, "san_francisco_ca"},
		{39.961, -82.998, "columbus_oh"},
		{39.768, -86.158, "indianapolis_in"},
		{32.755, -97.330, "fort_worth_tx"},
		{35.227, -80.843, "charlotte_nc"},
		{47.606, -122.332, "seattle_wa"},
		{39.739, -104.990, "denver_co"},
		{38.907, -77.036, "washington_dc"},
	}
	dataDir = "data"
)

func downloadTestData(t *testing.T) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("Failed to create data dir: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, loc := range locations {
		filename := filepath.Join(dataDir, fmt.Sprintf("%s.xhtml", loc.Name))
		if _, err := os.Stat(filename); err == nil {
			t.Logf("File %s already exists, skipping download", filename)
			continue
		}

		url := fmt.Sprintf("https://forecast.weather.gov/MapClick.php?lat=%.3f&lon=%.3f", loc.Lat, loc.Lon)
		t.Logf("Downloading %s from %s", filename, url)

		// Simple retry logic
		var resp *http.Response
		var err error
		for i := 0; i < 3; i++ {
			resp, err = client.Get(url)
			if err == nil && resp.StatusCode == http.StatusOK {
				break
			}
			time.Sleep(1 * time.Second)
		}
		if err != nil {
			t.Logf("Failed to download %s: %v", url, err)
			continue // Skip failing downloads to avoid blocking the whole test suite if one fails
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Logf("Failed to read body for %s: %v", url, err)
			continue
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", filename, err)
		}
	}
}
