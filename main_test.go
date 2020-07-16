package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/variantdev/vals"
)

func newVals(t *testing.T) *vals.Runtime {
	t.Helper()
	v, err := vals.New(vals.Options{})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func Test_renderFiles(t *testing.T) {
	v := newVals(t)

	wantVal := "myvalue"
	fileRef := "ref+echo://" + wantVal
	fileName := "/path/to/testfile"
	want := map[string]string{
		fileName: wantVal,
	}

	valsFiles := []string{
		fileName + ":" + fileRef,
	}

	rendered, err := renderFiles(v, valsFiles)
	assert.NoError(t, err)

	assert.Equal(t, want, rendered)

}
func Test_renderVars(t *testing.T) {
	v := newVals(t)

	wantVal := "myvalue"
	varRefKey := "TEST"
	varRefVal := "ref+echo://" + wantVal

	os.Setenv(varRefKey, varRefVal)
	defer os.Unsetenv(varRefKey)

	rendered, err := renderVars(v)
	assert.NoError(t, err)
	assert.Equal(t, wantVal, rendered[varRefKey])
}

func Test_buildEnvVars(t *testing.T) {
	envKey := "FOO"
	envVal := "bar"
	env := map[string]string{
		envKey: envVal,
	}
	os.Setenv(envKey, envVal)

	want := "FOO=bar"

	assert.Contains(t, buildEnvVars(env), want)
}

func Test_writeFiles(t *testing.T) {
	f, err := ioutil.TempFile("", "valstest")
	defer os.Remove(f.Name())
	assert.NoError(t, err)

	want := "myvalue"
	files := map[string]string{
		f.Name(): want,
	}

	err = writeFiles(files)
	assert.NoError(t, err)

	results, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)

	assert.Equal(t, want, string(results))
}
