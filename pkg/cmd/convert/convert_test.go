package convert_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jenkins-x-plugins/jx-tekton-to-actions/pkg/actions"
	"github.com/jenkins-x-plugins/jx-tekton-to-actions/pkg/cmd/convert"
	"github.com/jenkins-x/jx-helpers/v3/pkg/yamls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// generateTestOutput enable to regenerate the expected output
	generateTestOutput = false
)

func TestConvert(t *testing.T) {
	_, o := convert.NewCmdConvert()

	tmpDir, err := ioutil.TempDir("", "")
	require.NoError(t, err, "could not create temp dir")

	o.Dir = "test_data"
	o.OutDir = tmpDir
	err = o.Run()
	require.NoError(t, err, "Failed to run linter")

	names := []string{"jenkins-x-pr.yaml", "jenkins-x-release.yaml"}
	for _, name := range names {
		f := filepath.Join(tmpDir, name)
		expectedPath := filepath.Join("test_data", "expected", name)

		data, err := ioutil.ReadFile(f)
		require.NoError(t, err, "failed to load %s", f)

		if generateTestOutput {
			err = ioutil.WriteFile(expectedPath, data, 0666)
			require.NoError(t, err, "failed to save file %s", expectedPath)
			continue
		}
		expectedData, err := ioutil.ReadFile(expectedPath)
		require.NoError(t, err, "failed to load file "+expectedPath)

		text := strings.TrimSpace(string(data))
		expectedText := strings.TrimSpace(string(expectedData))

		assert.Equal(t, expectedText, text, "Task loaded for "+name)

		// lets check we can parse the workflow
		workflow := &actions.Workflow{}
		err = yamls.LoadFile(f, workflow)
		require.NoError(t, err, "failed to load file %s", f)
	}
}
