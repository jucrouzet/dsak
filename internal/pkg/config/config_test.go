// Package config contains the configuration handling for dsak.
package config //nolint:testpackage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getDefaultConfigFile(t *testing.T) {
	t.Run("uses environment var", func(t *testing.T) {
		require.NoError(t, os.Setenv("DSAK_CONFIGFILE", "foobar"))
		defer os.Unsetenv("DSAK_CONFIGFILE")
		file, err := getDefaultConfigFile()
		require.NoError(t, err)
		assert.Equal(t, "foobar", file)
	})

	t.Run("if no environment var, returns the right file", func(t *testing.T) {
		file, err := getDefaultConfigFile()
		require.NoError(t, err)
		dirname, err := os.UserHomeDir()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dirname, ".dsak.yaml"), file)
	})
}

func Test_getConfigFile(t *testing.T) {
	t.Run("if no flag specified, returns the default file", func(t *testing.T) {
		file, err := getConfigFile(nil)
		require.NoError(t, err)
		dirname, err := os.UserHomeDir()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dirname, ".dsak.yaml"), file)
	})

	t.Run("if a valid flag specified, returns the flag value", func(t *testing.T) {
		file, err := getConfigFile([]string{"-a", "--configfile", "bleh", "arg"})
		require.NoError(t, err)
		assert.Equal(t, "bleh", file)
		file, err = getConfigFile([]string{"--configfile", "bleh2"})
		require.NoError(t, err)
		assert.Equal(t, "bleh2", file)
	})

	t.Run("if an invalid flag specified, error", func(t *testing.T) {
		_, err := getConfigFile([]string{"-a", "--configfile"})
		require.Error(t, err)
		assert.ErrorContains(t, err, "--configfile requires an argument")
	})
}

func TestNew(t *testing.T) {
	t.Run("unexistant config file does not return an error", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "dsak-test-*")
		filename := tmpFile.Name()
		require.NoError(t, err)
		tmpFile.Close()
		require.NoError(t, os.Remove(filename))
		_, err = New([]string{"--configfile", filename})
		require.NoError(t, err)
	})

	t.Run("invalid config file returns an error", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "dsak-test-*")
		filename := tmpFile.Name()
		require.NoError(t, err)
		defer os.Remove(filename)
		_, err = tmpFile.WriteString(`dffdsfds`)
		require.NoError(t, err)
		tmpFile.Close()
		_, err = New([]string{"--configfile", filename})
		require.Error(t, err)
		assert.ErrorContains(t, err, "unmarshal errors")
	})

	t.Run("not yaml config file returns no error", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "dsak-test-*.json")
		filename := tmpFile.Name()
		require.NoError(t, err)
		defer os.Remove(filename)
		_, err = tmpFile.WriteString(`{"foo": "bar"}`)
		require.NoError(t, err)
		tmpFile.Close()
		c, err := New([]string{"--configfile", filename})
		require.NoError(t, err)
		assert.Equal(t, "bar", c.GetString("foo"))
	})

	t.Run("yaml config file returns no error", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "dsak-test-*")
		filename := tmpFile.Name()
		require.NoError(t, err)
		defer os.Remove(filename)
		_, err = tmpFile.WriteString(`a: 1`)
		require.NoError(t, err)
		tmpFile.Close()
		c, err := New([]string{"--configfile", filename})
		require.NoError(t, err)
		assert.Equal(t, 1, c.GetInt("a"))
	})
}

func TestWrite(t *testing.T) {
	t.Run("unexistant config file creates an empty file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "dsak-test-*")
		filename := tmpFile.Name()
		require.NoError(t, err)
		tmpFile.Close()
		require.NoError(t, os.Remove(filename))
		cfg, err := New([]string{"--configfile", filename})
		require.NoError(t, err)
		require.NoError(t, Write(cfg))
		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "{}\n", string(content))
	})

	t.Run("invalid config file path returns an error", func(t *testing.T) {
		filename := "/path/to/nowhere"
		cfg, err := New([]string{"--configfile", filename})
		require.NoError(t, err)
		err = Write(cfg)
		require.Error(t, err)
		assert.ErrorContains(t, err, "no such file or directory")
	})

	t.Run("not yaml config file writes format correctly", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "dsak-test-*.json")
		filename := tmpFile.Name()
		require.NoError(t, err)
		defer os.Remove(filename)
		_, err = tmpFile.WriteString(`{"foo": "bar"}`)
		require.NoError(t, err)
		tmpFile.Close()
		c, err := New([]string{"--configfile", filename})
		require.NoError(t, err)
		c.Set("hello", "world")
		err = Write(c)
		require.NoError(t, err)
		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "{\n  \"foo\": \"bar\",\n  \"hello\": \"world\"\n}", string(content))
	})

	t.Run("yaml config file writes format correctly", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "dsak-test-*")
		filename := tmpFile.Name()
		require.NoError(t, err)
		defer os.Remove(filename)
		_, err = tmpFile.WriteString(`a: 1`)
		require.NoError(t, err)
		tmpFile.Close()
		c, err := New([]string{"--configfile", filename})
		require.NoError(t, err)
		c.Set("hello", "world")
		err = Write(c)
		require.NoError(t, err)
		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "a: 1\nhello: world\n", string(content))
	})
}
