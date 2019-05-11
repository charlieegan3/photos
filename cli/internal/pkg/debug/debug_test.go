package debug

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestRunDebug(t *testing.T) {
	dir, err := ioutil.TempDir(".", "test_output")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	command := CreateDebugCmd()
	command.Flags().Set("output", dir)
	command.Flags().Set("source", "../../../test")

	RunDebug(command, []string{})

	content, err := ioutil.ReadFile(filepath.Join(dir, "index.json"))
	if err != nil {
		t.Error(err)
	}

	var index []struct {
		ID      string
		IsVideo bool
	}
	err = json.Unmarshal(content, &index)
	if err != nil {
		t.Error(err)
	}

	// assertions
	if len(index) != 5 {
		t.Errorf("unexpected number of items %v", len(index))
	}
	expected := []string{
		"2019-04-24-2029394066281649921",
		"2018-04-26-1765647895489858959",
		"2017-04-26-1501486602281729368",
		"2017-03-25-1478214926051861624",
		"2013-09-11-542730817640294126",
	}
	for i, v := range index {
		if v.ID != expected[i] {
			t.Errorf("unexpected id, got %v, want: %v", v.ID, expected[i])
		}
	}
}

func TestMissingOutputFolder(t *testing.T) {
	dir, err := ioutil.TempDir(".", "test_output")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	command := CreateDebugCmd()
	command.Flags().Set("output", filepath.Join(dir, "missing_output"))
	command.Flags().Set("source", "../../../test")

	RunDebug(command, []string{})

	if _, err := os.Stat(filepath.Join(dir, "missing_output/index.json")); os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestCopiesStaticFiles(t *testing.T) {
	dir, err := ioutil.TempDir(".", "test_output")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	command := CreateDebugCmd()
	command.Flags().Set("output", dir)
	command.Flags().Set("source", "../../../test")

	RunDebug(command, []string{})

	if _, err := os.Stat(filepath.Join(dir, "index.html")); os.IsNotExist(err) {
		t.Error(err)
	}
}
