package commander //nolint:testpackage

import (
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jucrouzet/dsak/internal/pkg/config"
)

func Test_buildCommandTree(t *testing.T) {
	emptyCreator := func() *cobra.Command {
		return &cobra.Command{}
	}

	t.Run("no root command", func(t *testing.T) {
		list := map[string]registeredCommand{
			"b": {
				creator: emptyCreator,
				configs: []string{},
			},
		}
		_, _, err := buildCommandTree(list)
		require.Error(t, err)
		assert.ErrorContains(t, err, "root command not found")
	})

	t.Run("unlogical hierarchy", func(t *testing.T) {
		list := map[string]registeredCommand{
			"": {
				creator: emptyCreator,
				configs: []string{},
			},
			"a>b": {
				creator: emptyCreator,
				configs: []string{},
			},
		}
		_, _, err := buildCommandTree(list)
		require.Error(t, err)
		assert.ErrorContains(t, err, "parent command of a>b not found")
		list = map[string]registeredCommand{
			"": {
				creator: emptyCreator,
				configs: []string{},
			},
			"a": {
				creator: emptyCreator,
				configs: []string{},
			},
			"a>b": {
				creator: emptyCreator,
				configs: []string{},
			},
			"a>b>c": {
				creator: emptyCreator,
				configs: []string{},
			},
			"a>c>d": {
				creator: emptyCreator,
				configs: []string{},
			},
		}
		_, _, err = buildCommandTree(list)
		require.Error(t, err)
		assert.ErrorContains(t, err, "parent command of a>c>d not found")
	})

	t.Run("valid list", func(t *testing.T) {
		list := map[string]registeredCommand{
			"a>ab>c": {
				creator: emptyCreator,
				configs: []string{},
			},
			"": {
				creator: emptyCreator,
				configs: []string{},
			},
			"a": {
				creator: emptyCreator,
				configs: []string{},
			},
			"a>aa": {
				creator: emptyCreator,
				configs: []string{},
			},
			"b": {
				creator: emptyCreator,
				configs: []string{},
			},
			"a>ab": {
				creator: emptyCreator,
				configs: []string{},
			},
		}
		_, cmds, err := buildCommandTree(list)
		assert.NoError(t, err)
		assert.Equal(t, 6, len(cmds))
	})

	t.Run("order respected", func(t *testing.T) {
		cobra.EnableTraverseRunHooks = true
		markers := []string{}
		markersMtx := sync.Mutex{}
		addMarker := func(name string) {
			markersMtx.Lock()
			defer markersMtx.Unlock()
			markers = append(markers, name)
		}
		creator := func(name, use string) func() *cobra.Command {
			return func() *cobra.Command {
				return &cobra.Command{
					Use: use,
					PersistentPreRun: func(_ *cobra.Command, _ []string) {
						addMarker(name)
					},
					Run: func(_ *cobra.Command, _ []string) {
					},
				}
			}
		}

		list := map[string]registeredCommand{
			"a>b": {
				creator: creator("a>b", "b"),
				configs: []string{},
			},
			"": {
				creator: creator("root", "test"),
				configs: []string{},
			},
			"b": {
				creator: creator("b", "b"),
				configs: []string{},
			},
			"a": {
				creator: creator("a", "a"),
				configs: []string{},
			},
		}
		rootCmd, _, err := buildCommandTree(list)
		require.NoError(t, err)
		rootCmd.SetArgs([]string{"a", "b"})
		require.NoError(t, rootCmd.Execute())
		assert.Equal(t, []string{"root", "a", "a>b"}, markers)
	})
}

func TestRegister(t *testing.T) {
	t.Run("valid registration", func(t *testing.T) {
		defer func() {
			commandsRegistered = make(map[string]registeredCommand)
		}()
		assert.NotPanics(
			t,
			func() {
				Register(
					"test",
					func() *cobra.Command {
						return &cobra.Command{}
					},
				)
			},
		)
		assert.NotPanics(
			t,
			func() {
				Register(
					"test2",
					func() *cobra.Command {
						return &cobra.Command{}
					},
					WithConfig("test"),
				)
			},
		)
	})

	t.Run("invalid registrations", func(t *testing.T) {
		assert.Panics(
			t,
			func() {
				Register(
					" test",
					func() *cobra.Command {
						return &cobra.Command{}
					},
				)
			},
		)
		assert.NotPanics(
			t,
			func() {
				Register(
					"test",
					func() *cobra.Command {
						return &cobra.Command{}
					},
					WithConfig("test"),
				)
			},
		)
		assert.Panics(
			t,
			func() {
				Register(
					"test",
					func() *cobra.Command {
						return &cobra.Command{}
					},
					WithConfig("test"),
				)
			},
		)
	})
}

func Test_applyConfigs(t *testing.T) {
	t.Run("all valid", func(t *testing.T) {
		emptyCreator := func() *cobra.Command {
			return &cobra.Command{}
		}
		config.RegisterValue("test.commander.applyconfigs.one", config.ValueTypeBool)
		list := map[string]registeredCommand{
			"": {
				creator: emptyCreator,
				configs: []string{"test.commander.applyconfigs.one"},
			},
			"a": {
				creator: emptyCreator,
				configs: []string{},
			},
		}
		cmds := map[string]*cobra.Command{
			"":  emptyCreator(),
			"a": emptyCreator(),
		}
		require.NoError(t, applyConfigs(list, cmds, viper.New()))
	})
	t.Run("unknown config", func(t *testing.T) {
		emptyCreator := func() *cobra.Command {
			return &cobra.Command{}
		}
		list := map[string]registeredCommand{
			"": {
				creator: emptyCreator,
				configs: []string{"dsqdsqsqdsq"},
			},
			"a": {
				creator: emptyCreator,
				configs: []string{},
			},
		}
		cmds := map[string]*cobra.Command{
			"":  emptyCreator(),
			"a": emptyCreator(),
		}
		require.Error(t, applyConfigs(list, cmds, viper.New()))
		assert.ErrorContains(t, applyConfigs(list, cmds, viper.New()), "value dsqdsqsqdsq not found")
	})
}
