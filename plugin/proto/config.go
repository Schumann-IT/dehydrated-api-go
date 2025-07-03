package proto

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/types/known/structpb"
)

// PluginConfig represents a map of configuration values that can be converted to and from proto values
type PluginConfig struct {
	values map[string]*PluginConfigValue
}

// NewPluginConfig creates a new PluginConfig
func NewPluginConfig() *PluginConfig {
	return &PluginConfig{
		values: make(map[string]*PluginConfigValue),
	}
}

// FromProto populates the PluginConfig from a map of proto.Values, converting each entry to a PluginConfigValue.
func (cm *PluginConfig) FromProto(m map[string]*structpb.Value) {
	for k, v := range m {
		cm.values[k] = FromProto(v)
	}
}

// ToProto converts the PluginConfig to a proto value map
func (cm *PluginConfig) ToProto() (map[string]*structpb.Value, error) {
	result := make(map[string]*structpb.Value)
	for k, v := range cm.values {
		protoVal, err := v.ToProto()
		if err != nil {
			return nil, fmt.Errorf("failed to convert value for key %s: %w", k, err)
		}
		result[k] = protoVal
	}
	return result, nil
}

// Get returns a PluginConfigValue for the given key
func (cm *PluginConfig) Get(key string) *PluginConfigValue {
	return cm.values[key]
}

// Set sets a value for the given key
func (cm *PluginConfig) Set(key string, value any) {
	cm.values[key] = NewConfigValue(value)
}

// GetString returns the value for the given key as a string
func (cm *PluginConfig) GetString(key string) (string, error) {
	if v, ok := cm.values[key]; ok {
		return v.GetString()
	}
	return "", fmt.Errorf("key not found: %s", key)
}

// GetInt returns the value for the given key as an int
func (cm *PluginConfig) GetInt(key string) (int, error) {
	if v, ok := cm.values[key]; ok {
		return v.GetInt()
	}
	return 0, fmt.Errorf("key not found: %s", key)
}

// GetFloat returns the value for the given key as a float64
func (cm *PluginConfig) GetFloat(key string) (float64, error) {
	if v, ok := cm.values[key]; ok {
		return v.GetFloat()
	}
	return 0, fmt.Errorf("key not found: %s", key)
}

// GetBool returns the value for the given key as a bool
func (cm *PluginConfig) GetBool(key string) (bool, error) {
	if v, ok := cm.values[key]; ok {
		return v.GetBool()
	}
	return false, fmt.Errorf("key not found: %s", key)
}

// GetStringSlice returns the value for the given key as a []string
func (cm *PluginConfig) GetStringSlice(key string) ([]string, error) {
	if v, ok := cm.values[key]; ok {
		return v.GetStringSlice()
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

// GetMap returns the value for the given key as a map[string]interface{}
func (cm *PluginConfig) GetMap(key string) (map[string]any, error) {
	if v, ok := cm.values[key]; ok {
		return v.GetMap()
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

// GetStruct converts the value for the given key to the given struct type
func (cm *PluginConfig) GetStruct(key string, target any) error {
	if v, ok := cm.values[key]; ok {
		return v.GetStruct(target)
	}
	return fmt.Errorf("key not found: %s", key)
}

// PluginConfigValue represents a configuration value that can be converted to and from proto values
type PluginConfigValue struct {
	value any
}

// NewConfigValue creates a new PluginConfigValue from a Go value
func NewConfigValue(v any) *PluginConfigValue {
	return &PluginConfigValue{value: v}
}

// ToProto converts the PluginConfigValue to a proto Value
func (cv *PluginConfigValue) ToProto() (*structpb.Value, error) {
	return structpb.NewValue(cv.value)
}

// FromProto converts a proto Value to a PluginConfigValue
func FromProto(v *structpb.Value) *PluginConfigValue {
	return &PluginConfigValue{value: v.AsInterface()}
}

// GetString returns the value as a string
func (cv *PluginConfigValue) GetString() (string, error) {
	if str, ok := cv.value.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("value is not a string: %v", cv.value)
}

// GetInt returns the value as an int
func (cv *PluginConfigValue) GetInt() (int, error) {
	switch v := cv.value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("value is not a number: %v", cv.value)
	}
}

// GetFloat returns the value as a float64
func (cv *PluginConfigValue) GetFloat() (float64, error) {
	switch v := cv.value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("value is not a number: %v", cv.value)
	}
}

// GetBool returns the value as a bool
func (cv *PluginConfigValue) GetBool() (bool, error) {
	if b, ok := cv.value.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("value is not a boolean: %v", cv.value)
}

// GetStringSlice returns the value as a []string
func (cv *PluginConfigValue) GetStringSlice() ([]string, error) {
	if slice, ok := cv.value.([]any); ok {
		result := make([]string, len(slice))
		for i, v := range slice {
			if str, ok := v.(string); ok {
				result[i] = str
			} else {
				return nil, fmt.Errorf("value at index %d is not a string: %v", i, v)
			}
		}
		return result, nil
	}
	return nil, fmt.Errorf("value is not a slice: %v", cv.value)
}

// GetMap returns the value as a map[string]interface{}
func (cv *PluginConfigValue) GetMap() (map[string]any, error) {
	if m, ok := cv.value.(map[string]any); ok {
		return m, nil
	}
	return nil, fmt.Errorf("value is not a map: %v", cv.value)
}

// GetStruct converts the value to the given struct type
// @TODO: Refactor if needed.
//
//nolint:gocyclo // This function is intentionally complex to handle various types and conversions.
func (cv *PluginConfigValue) GetStruct(target any) error {
	if m, ok := cv.value.(map[string]any); ok {
		val := reflect.ValueOf(target)
		if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("target must be a pointer to a struct")
		}
		val = val.Elem()
		typ := val.Type()

		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := typ.Field(i)
			tag := fieldType.Tag.Get("json")
			if tag == "" {
				tag = fieldType.Name
			}

			if value, ok := m[tag]; ok {
				if !field.CanSet() {
					continue
				}

				switch field.Kind() {
				case reflect.String:
					if str, ok := value.(string); ok {
						field.SetString(str)
					}
				case reflect.Int, reflect.Int32, reflect.Int64:
					if num, ok := value.(float64); ok {
						field.SetInt(int64(num))
					}
				case reflect.Float32, reflect.Float64:
					if num, ok := value.(float64); ok {
						field.SetFloat(num)
					}
				case reflect.Bool:
					if b, ok := value.(bool); ok {
						field.SetBool(b)
					}
				case reflect.Slice:
					if slice, ok := value.([]any); ok {
						sliceType := field.Type().Elem()
						newSlice := reflect.MakeSlice(field.Type(), len(slice), len(slice))
						for j, v := range slice {
							switch sliceType.Kind() {
							case reflect.String:
								if str, ok := v.(string); ok {
									newSlice.Index(j).SetString(str)
								}
							case reflect.Int, reflect.Int32, reflect.Int64:
								if num, ok := v.(float64); ok {
									newSlice.Index(j).SetInt(int64(num))
								}
							case reflect.Float32, reflect.Float64:
								if num, ok := v.(float64); ok {
									newSlice.Index(j).SetFloat(num)
								}
							case reflect.Bool:
								if b, ok := v.(bool); ok {
									newSlice.Index(j).SetBool(b)
								}
							}
						}
						field.Set(newSlice)
					}
				}
			}
		}
		return nil
	}
	return fmt.Errorf("value is not a map: %v", cv.value)
}
