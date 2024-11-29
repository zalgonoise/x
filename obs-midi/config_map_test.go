package obs_midi

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	c := &Config{}

	data, err := os.ReadFile("./testing/testdata/local.json")

	if errors.Is(err, os.ErrNotExist) {
		t.Skip()

		return
	}

	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatal(err)
	}

	m := NewConfigMap(c)

	data, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))

	if path := os.Getenv("OBS_MIDI_CONFIG_OUTPUT"); path != "" {
		f, err := os.Create(path + `obs-midi-config.json`)
		if err != nil {
			t.Fatal(err)
		}

		defer f.Close()

		if _, err := f.Write(data); err != nil {
			t.Fatal(err)
		}

		t.Log("created file in " + path + `obs-midi-config.json`)
	}
}
