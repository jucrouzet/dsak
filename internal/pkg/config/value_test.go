package config //nolint:testpackage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValue(t *testing.T) {
	defer func() {
		values = make(map[string]Value)
	}()
	RegisterValue("test.config.getvalue.one", ValueTypeBool)
	v, err := GetValue("test.config.getvalue.one")
	assert.NoError(t, err)
	assert.Equal(t, ValueTypeBool, v.valueType)
	_, err = GetValue("fdsfdsfdse")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "value fdsfdsfdse not found")
}

func TestRegisterValue(t *testing.T) {
	t.Run("valid registration", func(t *testing.T) {
		defer func() {
			values = make(map[string]Value)
		}()
		assert.NotPanics(
			t,
			func() {
				RegisterValue(
					"test",
					ValueTypeBool,
				)
			},
		)
		assert.NotPanics(
			t,
			func() {
				RegisterValue(
					"test.subtest",
					ValueTypeBool,
				)
			},
		)
	})

	t.Run("invalid registrations", func(t *testing.T) {
		defer func() {
			values = make(map[string]Value)
		}()
		assert.NotPanics(
			t,
			func() {
				RegisterValue(
					"test.subtest",
					ValueTypeBool,
				)
			},
		)
		assert.Panics(
			t,
			func() {
				RegisterValue(
					"",
					ValueTypeBool,
				)
			},
		)
		assert.Panics(
			t,
			func() {
				RegisterValue(
					"configfile",
					ValueTypeBool,
				)
			},
		)
		assert.Panics(
			t,
			func() {
				RegisterValue(
					"test lol",
					ValueTypeBool,
				)
			},
		)
		assert.Panics(
			t,
			func() {
				RegisterValue(
					"test.subtest",
					ValueTypeBool,
				)
			},
		)
		assert.Panics(
			t,
			func() {
				RegisterValue(
					"test.subtest",
					ValueType(9999),
				)
			},
		)
	})
}

func TestFlag(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		shouldError bool
	}{
		{
			name: "valid flag",
			arg:  "test",
		},
		{
			name:        "empty flag",
			arg:         "",
			shouldError: true,
		},
		{
			name:        "invalid flag",
			arg:         "lol O",
			shouldError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Value{}
			f := Flag(tt.arg)
			err := f(v)
			if tt.shouldError {
				assert.Error(t, err, fmt.Sprintf("test %s should return an error", tt.name))
				return
			}
			assert.Equal(t, tt.arg, v.flag, fmt.Sprintf("test %s should set flag to %s", tt.name, tt.arg))
		})
	}
}

func TestShortFlag(t *testing.T) {
	tests := []struct {
		name        string
		arg         byte
		shouldError bool
	}{
		{
			name: "valid flag",
			arg:  'c',
		},
		{
			name:        "zero flag",
			arg:         0,
			shouldError: true,
		},
		{
			name:        "invalid flag",
			arg:         10,
			shouldError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Value{}
			f := ShortFlag(tt.arg)
			err := f(v)
			if tt.shouldError {
				assert.Error(t, err, fmt.Sprintf("test %s should return an error", tt.name))
				return
			}
			assert.Equal(t, tt.arg, v.shortFlag, fmt.Sprintf("test %s should set flag to %x", tt.name, tt.arg))
		})
	}
}

func TestFlagIsPersistent(t *testing.T) {
	v := &Value{}
	assert.Equal(t, false, v.persistentFlag)
	f := FlagIsPersistent()
	err := f(v)
	assert.NoError(t, err)
	assert.Equal(t, true, v.persistentFlag)
}

func TestIgnoreEnv(t *testing.T) {
	v := &Value{}
	assert.Equal(t, false, v.noEnv)
	f := IgnoreEnv()
	err := f(v)
	assert.NoError(t, err)
	assert.Equal(t, true, v.noEnv)
}

func TestDescription(t *testing.T) {
	v := &Value{}
	assert.Equal(t, "", v.flagDescription)
	f := Description("coucou")
	err := f(v)
	assert.NoError(t, err)
	assert.Equal(t, "coucou", v.flagDescription)
}

func TestDefaultValue(t *testing.T) {
	tests := []struct {
		arg         interface{}
		name        string
		shouldError bool
		value       *Value
	}{
		{
			name:  "default bool true",
			value: &Value{valueType: ValueTypeBool},
			arg:   true,
		},
		{
			name:  "default bool true",
			value: &Value{valueType: ValueTypeBool},
			arg:   false,
		},
		{
			name:        "default bool invalid",
			value:       &Value{valueType: ValueTypeBool},
			arg:         "false",
			shouldError: true,
		},
		{
			name:  "default empty string",
			value: &Value{valueType: ValueTypeString},
			arg:   "",
		},
		{
			name:  "default string",
			value: &Value{valueType: ValueTypeString},
			arg:   "hello",
		},
		{
			name:        "default string invalid",
			value:       &Value{valueType: ValueTypeString},
			arg:         12,
			shouldError: true,
		},
		{
			name:  "default empty uint",
			value: &Value{valueType: ValueTypeUint},
			arg:   uint64(0),
		},
		{
			name:  "default uint",
			value: &Value{valueType: ValueTypeUint},
			arg:   uint64(12),
		},
		{
			name:        "default uint invalid",
			value:       &Value{valueType: ValueTypeUint},
			arg:         "hello",
			shouldError: true,
		},
		{
			name:        "default uint invalid (not an uint64)",
			value:       &Value{valueType: ValueTypeUint},
			arg:         12,
			shouldError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.arg != nil {
				f := DefaultValue(tt.arg)
				err := f(tt.value)
				if tt.shouldError {
					assert.Error(t, err, fmt.Sprintf("test %s should return an error", tt.name))
					return
				}
			}
			assert.Equal(t, tt.arg, tt.value.defaultValue, fmt.Sprintf("test %s should have a default to flag to %v", tt.name, tt.arg))
		})
	}
}
