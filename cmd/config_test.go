package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func testConfig(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg.printAll(buf)
	output := string(buf.Bytes())

	phrases := []string{
		"Service Carryout",
		"Name joe",
		"Email nojoe@mail.com",
		"City:Washington DC",
		"Street:1600 Pennsylvania Ave NW",
	}

	for _, phrase := range phrases {
		if !strings.Contains(output, phrase) {
			t.Error("wrong output")
		}
	}
}

func testConfigGet(t *testing.T) {
	t.Error("testConfigGet")
}

func testConfigSet(t *testing.T) {
	t.Error("testConfigSet")
}
