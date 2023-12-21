package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v3"
)

// ValueType is the type of a configuration value.
type ValueType uint16

const (
	// ValueTypeString represents a string.
	ValueTypeString ValueType = iota + 1
	// ValueTypeStrings represents a string slice.
	ValueTypeStrings
	// ValueTypeUint represents an unsigned integer value, on 64 bits.
	ValueTypeUint
	// ValueTypeBool represents a boolean value.
	ValueTypeBool
	// ValueTypeStringsMap represents a map[string][]string.
	ValueTypeStringsMap
)

// StringerFunc is a function that returns a string representing the value.
type StringerFunc func(cmd *cobra.Command) string

// SetterFunc is a function that sets the value from a given string.
type SetterFunc func(cmd *cobra.Command, val string) error

// Value represents a configuration value.
type Value struct {
	defaultValue    interface{}
	flag            string
	flagDescription string
	name            string
	noEnv           bool
	persistentFlag  bool
	setter          SetterFunc
	shortFlag       byte
	stringer        StringerFunc
	valueType       ValueType
}

// ValueOption is a function that can be used to configure a configuration value.
type ValueOption func(*Value) error

var values = map[string]Value{}

// GetValue returns the configuration value with the given name.
func GetValue(name string) (*Value, error) {
	v, ok := values[name]
	if !ok {
		return nil, fmt.Errorf("value %s not found", name)
	}
	return &v, nil
}

// GetValues returns all registered values.
func GetValues() []*Value {
	res := make([]*Value, 0, len(values))
	for _, v := range values {
		newV := &Value{}
		*newV = v
		res = append(res, newV)
	}
	return res
}

var valueNameIsLetter = regexp.MustCompile(`^[a-z][a-z0-9]+$`).MatchString

// Apply applies configuration value on a command.
func (c Value) Apply(cmd *cobra.Command, cfg *viper.Viper) error {
	return errors.Join(
		c.applyFlag(cmd),
		c.applyConfig(cmd, cfg),
	)
}

// GetName gets the configuration value's name.
func (c Value) GetName() string {
	return c.name
}

// GetFlag gets the configuration value's flag.
func (c Value) GetFlag() string {
	return c.flag
}

// AsString gets the configuration value as a string.
func (c Value) AsString(cmd *cobra.Command) string {
	if c.stringer != nil {
		return c.stringer(cmd)
	}
	cfg := GetFromCommandContext(cmd)
	switch c.valueType {
	case ValueTypeString:
		str := cfg.GetString(c.name)
		return fmt.Sprintf(`%q`, str)
	case ValueTypeStrings:
		strs := cfg.GetStringSlice(c.name)
		values := make([]string, len(strs))
		for i, str := range strs {
			values[i] = fmt.Sprintf(`%q`, str)
		}
		return fmt.Sprintf(`[%s]`, strings.Join(values, ", "))
	case ValueTypeUint:
		val := cfg.GetUint64(c.name)
		p := message.NewPrinter(language.English)
		return p.Sprintf("%d", val)
	case ValueTypeBool:
		if cfg.GetBool(c.name) {
			return "true"
		}
		return "false"
	case ValueTypeStringsMap:
		v, err := yaml.Marshal(cfg.GetStringMapStringSlice(c.name))
		if err != nil {
			return err.Error()
		}
		return string(v)
	}
	return ""
}

// AsRawString gets the configuration value as a raw string, no decorator or quotes.
func (c Value) AsRawString(cmd *cobra.Command) string {
	cfg := GetFromCommandContext(cmd)
	switch c.valueType {
	case ValueTypeString:
		return cfg.GetString(c.name)
	case ValueTypeStrings:
		return strings.Join(cfg.GetStringSlice(c.name), "\n")
	case ValueTypeUint:
		return strconv.FormatUint(cfg.GetUint64(c.name), 10)
	case ValueTypeBool:
		if cfg.GetBool(c.name) {
			return "true"
		}
		return "false"
	case ValueTypeStringsMap:
		v, err := json.Marshal(cfg.GetStringMapStringSlice(c.name))
		if err != nil {
			return err.Error()
		}
		return string(v)
	}
	return ""
}

// Set sets the configuration value from the given string.
func (c Value) Set(cmd *cobra.Command, val string) error {
	if c.setter != nil {
		return c.setter(cmd, val)
	}
	cfg := GetFromCommandContext(cmd)
	var v interface{}
	switch c.valueType {
	case ValueTypeString:
		v = val
	case ValueTypeStrings:
		return errors.New("setting this configuration value from string is not supported")
	case ValueTypeUint:
		vv, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		v = vv
	case ValueTypeBool:
		if strings.EqualFold(val, "true") || val == "1" {
			v = true
		} else if strings.EqualFold(val, "false") || val == "0" {
			v = false
		} else {
			return fmt.Errorf("invalid boolean value: %s", val)
		}
	case ValueTypeStringsMap:
		return errors.New("setting this configuration value from string is not supported")
	}
	cfg.Set(c.name, v)
	return nil
}

func (c Value) applyFlag(cmd *cobra.Command) error {
	if c.flag == "" {
		return nil
	}
	flagSet := cmd.Flags()
	if c.persistentFlag {
		flagSet = cmd.PersistentFlags()
	}
	if existing := flagSet.Lookup(c.flag); existing != nil {
		return fmt.Errorf("flag %s already exists", c.flag)
	}
	switch c.valueType {
	case ValueTypeString:
		if c.shortFlag != 0 {
			flagSet.StringP(c.flag, string(c.shortFlag), c.defaultValue.(string), c.getDescription()) //nolint:forcetypeassert
		} else {
			flagSet.String(c.flag, c.defaultValue.(string), c.getDescription()) //nolint:forcetypeassert
		}
	case ValueTypeStrings:
		if c.shortFlag != 0 {
			flagSet.StringArrayP(c.flag, string(c.shortFlag), c.defaultValue.([]string), c.getDescription()) //nolint:forcetypeassert
		} else {
			flagSet.StringArray(c.flag, c.defaultValue.([]string), c.getDescription()) //nolint:forcetypeassert
		}
	case ValueTypeUint:
		if c.shortFlag != 0 {
			flagSet.Uint64P(c.flag, string(c.shortFlag), c.defaultValue.(uint64), c.getDescription()) //nolint:forcetypeassert
		} else {
			flagSet.Uint64(c.flag, c.defaultValue.(uint64), c.getDescription()) //nolint:forcetypeassert
		}
	case ValueTypeBool:
		if c.shortFlag != 0 {
			flagSet.BoolP(c.flag, string(c.shortFlag), c.defaultValue.(bool), c.getDescription()) //nolint:forcetypeassert
		} else {
			flagSet.Bool(c.flag, c.defaultValue.(bool), c.getDescription()) //nolint:forcetypeassert
		}
	}
	return nil
}

func (c Value) getDescription() string {
	description := c.flagDescription
	if description == "" {
		description = c.name
	}
	return description
}

func (c Value) applyConfig(cmd *cobra.Command, cfg *viper.Viper) error {
	if c.flag != "" {
		flagSet := cmd.Flags()
		if c.persistentFlag {
			flagSet = cmd.PersistentFlags()
		}
		if err := cfg.BindPFlag(c.name, flagSet.Lookup(c.flag)); err != nil {
			return err
		}
	} else {
		cfg.SetDefault(c.name, c.defaultValue)
	}
	if !c.noEnv {
		envKey := fmt.Sprintf("DSAK_%s", strings.ReplaceAll(strings.ToUpper(c.name), ".", "_"))
		return cfg.BindEnv(c.name, envKey)
	}
	return nil
}

// RegisterValue registers a configuration value.
// `name` is the global name of the configuration value, must be unique.
// `valueType` is the type of this configuration value.
// Configurations are taken in this order :
// 1. command line flag
// 2. configuration file
// 3. environment variable.
// Names must be composed of words of lowercase letters separated by period.
// Examples : "timeout", "global.verbose"
// RegisterValue should be called in the init() function of each command.
func RegisterValue(name string, valueType ValueType, opts ...ValueOption) {
	if name == "" {
		panic(fmt.Errorf("empty configuration name"))
	}
	if name == "configfile" {
		panic(fmt.Errorf("configfile is a reserved name"))
	}
	for _, part := range strings.Split(name, ".") {
		if part == "" || !valueNameIsLetter(part) {
			panic(fmt.Errorf("invalid configuration name: %s", name))
		}
	}
	if _, ok := values[name]; ok {
		panic(fmt.Errorf("configuration name is already registered: %s", name))
	}
	v := Value{
		flagDescription: fmt.Sprintf("set the value for the configuration %s", name),
		name:            name,
		valueType:       valueType,
	}
	switch v.valueType {
	case ValueTypeString:
		v.defaultValue = ""
	case ValueTypeStrings:
		v.defaultValue = make([]string, 0)
	case ValueTypeUint:
		v.defaultValue = uint64(0)
	case ValueTypeBool:
		v.defaultValue = false
	case ValueTypeStringsMap:
		v.defaultValue = make(map[string][]string)
		v.noEnv = true
	default:
		panic(fmt.Errorf("unknown configuration value type for %s: %d", name, valueType))
	}
	for _, opt := range opts {
		if err := opt(&v); err != nil {
			panic(fmt.Errorf("invalid option for configuration value %s: %w", name, err))
		}
	}
	values[name] = v
}

var valueFlagIsValid = regexp.MustCompile(`^[a-z][a-z0-9\-]+$`).MatchString

// Flag indicates that this configuration value is controlled by a flag.
func Flag(flag string) ValueOption {
	return func(v *Value) error {
		if !valueFlagIsValid(flag) {
			return fmt.Errorf("invalid configuration flag: %s", flag)
		}
		switch v.valueType {
		case ValueTypeStringsMap:
			return errors.New("ValueTypeStringsMap cannot be set with flags")
		default:
		}
		v.flag = flag
		return nil
	}
}

// ShortFlag indicates that this configuration value is controlled by a short flag.
func ShortFlag(flag byte) ValueOption {
	return func(v *Value) error {
		if (flag < 'a' || flag > 'z') && (flag < 'A' || flag > 'Z') && (flag < '0' || flag > '9') {
			return fmt.Errorf("invalid configuration short flag")
		}
		switch v.valueType {
		case ValueTypeStringsMap:
			return errors.New("ValueTypeStringsMap cannot be set with flags")
		default:
		}
		v.shortFlag = flag
		return nil
	}
}

// FlagIsPersistant indicates that this configuration is controlled by flag that is persistent in subcommands.
func FlagIsPersistent() ValueOption {
	return func(v *Value) error {
		switch v.valueType {
		case ValueTypeStringsMap:
			return errors.New("ValueTypeStringsMap cannot be set with flags")
		default:
		}
		v.persistentFlag = true
		return nil
	}
}

// IgnoreEnv indicates that this configuration is not controlled by an environment variable.
func IgnoreEnv() ValueOption {
	return func(v *Value) error {
		v.noEnv = true
		return nil
	}
}

// Description sets configuration description.
func Description(description string) ValueOption {
	return func(v *Value) error {
		v.flagDescription = description
		return nil
	}
}

// DefaultValue sets the default value for this configuration value.
func DefaultValue(v interface{}) ValueOption {
	return func(c *Value) error {
		switch c.valueType {
		case ValueTypeString:
			vv, ok := v.(string)
			if !ok {
				return fmt.Errorf("invalid value for string: %T", v)
			}
			c.defaultValue = vv
		case ValueTypeStrings:
			vv, ok := v.([]string)
			if !ok {
				return fmt.Errorf("invalid value for []string: %T", v)
			}
			c.defaultValue = vv
		case ValueTypeUint:
			vv, ok := v.(uint64)
			if !ok {
				return fmt.Errorf("invalid value for uint: %T", v)
			}
			c.defaultValue = vv
		case ValueTypeBool:
			vv, ok := v.(bool)
			if !ok {
				return fmt.Errorf("invalid value for bool: %T", v)
			}
			c.defaultValue = vv
		case ValueTypeStringsMap:
			vv, ok := v.(map[string][]string)
			if !ok {
				return fmt.Errorf("invalid value for map[string][]string: %T", v)
			}
			c.defaultValue = vv
		}
		return nil
	}
}

// Stringer sets a function to be used when value should be represented as string
// instead of using fmt.Sprintf("%v").
func Stringer(f StringerFunc) ValueOption {
	return func(c *Value) error {
		c.stringer = f
		return nil
	}
}

// Setter sets a function to be used when setting value from dsak instead of default func.
func Setter(f SetterFunc) ValueOption {
	return func(c *Value) error {
		c.setter = f
		return nil
	}
}
