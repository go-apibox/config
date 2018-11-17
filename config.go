// 配置

package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Delimiter string
	cfgData   map[interface{}]interface{}
}

// FromFile create a config with specified config file.
func FromFile(configFile string) (*Config, error) {
	cfgBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	config, err := newConfigWithBytes(cfgBytes)
	if err != nil {
		return nil, err
	}

	// include sub config
	incItems := make([]string, 0, 5)
	if v, ok := config.cfgData["include"]; ok {
		switch vv := v.(type) {
		case string:
			incItems = append(incItems, vv)
		case []interface{}:
			for _, vvv := range vv {
				if item, ok := vvv.(string); !ok {
					return nil, errors.New("unrecoginzed config value of `include`")
				} else {
					incItems = append(incItems, item)
				}
			}
		default:
			return nil, errors.New("unrecoginzed config value of `include`")
		}
	}

	configDir := filepath.Dir(configFile)
	for _, incItem := range incItems {
		incFile := filepath.Join(configDir, incItem+".yaml")
		incCfgBytes, err := ioutil.ReadFile(incFile)
		if err != nil {
			return nil, err
		}
		incCfgData := make(map[interface{}]interface{})
		err = yaml.Unmarshal(incCfgBytes, &incCfgData)
		if err != nil {
			return nil, err
		}
		configDeepMerge(config.cfgData, incCfgData)
	}

	return config, nil
}

// FromString create a config by specified yaml string.
func FromString(yamlStr string) (*Config, error) {
	cfgBytes := []byte(yamlStr)

	return newConfigWithBytes(cfgBytes)
}

func newConfigWithBytes(cfgBytes []byte) (*Config, error) {
	// slice the BOM
	if len(cfgBytes) >= 3 && cfgBytes[0] == 239 && cfgBytes[1] == 187 && cfgBytes[2] == 191 {
		cfgBytes = cfgBytes[3:]
	}

	cfgData := make(map[interface{}]interface{})
	err := yaml.Unmarshal(cfgBytes, &cfgData)
	if err != nil {
		return nil, err
	}

	return &Config{".", cfgData}, nil
}

// merge two config maps
func configDeepMerge(dst map[interface{}]interface{}, src map[interface{}]interface{}) {
	for k, v := range src {
		// 目标不存在该key，则直接复制
		if _, ok := dst[k]; !ok {
			dst[k] = v
			continue
		}

		// 目标存在该key，但不是map，则直接覆盖
		dstV, ok := dst[k].(map[interface{}]interface{})
		if !ok {
			dst[k] = v
			continue
		}

		// 源不是map，直接覆盖
		srcV, ok := v.(map[interface{}]interface{})
		if !ok {
			dst[k] = v
			continue
		}

		// 目标存在该key，而且是map，且源是map，则递归合并
		configDeepMerge(dstV, srcV)
	}
}

// GetString returns the string value for a given key.
func (c *Config) GetString(key string) (string, error) {
	v, err := c.Get(key)
	if err != nil {
		return "", err
	}

	str, ok := v.(string)
	if !ok {
		return "", errors.New("value of `" + key + "` is not string")
	}

	return str, nil
}

// GetDefaultString returns the string value for a given key.
// if error occur, return defaultVal
func (c *Config) GetDefaultString(key string, defaultVal string) string {
	if v, err := c.GetString(key); err != nil {
		return defaultVal
	} else {
		return v
	}
}

// GetStringArray returns the []string value for a given key.
func (c *Config) GetStringArray(key string) ([]string, error) {
	v, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	t, ok := v.([]interface{})
	if !ok {
		return nil, errors.New("value of `" + key + "` is not a string list")
	}

	vArr := make([]string, 0, len(t))
	for _, vv := range t {
		if vvv, ok := vv.(string); ok {
			vArr = append(vArr, vvv)
		} else {
			return nil, errors.New("some value in key `" + key + "` is not string")
		}
	}
	return vArr, nil
}

// GetDefaultStringArray returns the []string value for a given key.
// if error occur, return defaultVal
func (c *Config) GetDefaultStringArray(key string, defaultVal []string) []string {
	if v, err := c.GetStringArray(key); err != nil {
		return defaultVal
	} else {
		return v
	}
}

// GetInt returns the int value for a given key.
func (c *Config) GetInt(key string) (int, error) {
	v, err := c.Get(key)
	if err != nil {
		return 0, err
	}

	switch vv := v.(type) {
	case int:
		return vv, nil
	case string:
		if vvv, err := strconv.Atoi(vv); err == nil {
			return vvv, nil
		} else {
			return 0, errors.New("value of `" + key + "` is not int")
		}
	default:
		return 0, errors.New("value of `" + key + "` is not int")
	}
}

// GetDefaultInt returns the int value for a given key.
// if error occur, return defaultVal
func (c *Config) GetDefaultInt(key string, defaultVal int) int {
	if v, err := c.GetInt(key); err != nil {
		return defaultVal
	} else {
		return v
	}
}

// GetIntArray returns the []int value for a given key.
func (c *Config) GetIntArray(key string) ([]int, error) {
	v, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	t, ok := v.([]interface{})
	if !ok {
		return nil, errors.New("value of `" + key + "` is not a int list")
	}

	vArr := make([]int, 0, len(t))
	for _, vv := range t {
		switch vvv := vv.(type) {
		case int:
			vArr = append(vArr, vvv)
		case string:
			if vvvv, err := strconv.Atoi(vvv); err == nil {
				vArr = append(vArr, vvvv)
			} else {
				return nil, errors.New("some value in key `" + key + "` is not int")
			}
		default:
			return nil, errors.New("some value in key `" + key + "` is not int")
		}
	}
	return vArr, nil
}

// GetDefaultIntArray returns the []int value for a given key.
// if error occur, return defaultVal
func (c *Config) GetDefaultIntArray(key string, defaultVal []int) []int {
	if v, err := c.GetIntArray(key); err != nil {
		return defaultVal
	} else {
		return v
	}
}

// GetBool returns the boolean value for a given key.
func (c *Config) GetBool(key string) (bool, error) {
	v, err := c.Get(key)
	if err != nil {
		return false, err
	}

	switch vv := v.(type) {
	case bool:
		return vv, nil
	case string:
		if vvv, err := strconv.ParseBool(vv); err == nil {
			return vvv, nil
		} else {
			return false, errors.New("value of `" + key + "` is not boolean")
		}
	default:
		return false, errors.New("value of `" + key + "` is not boolean")
	}
}

// GetDefaultBool returns the boolean value for a given key.
// if error occur, return defaultVal
func (c *Config) GetDefaultBool(key string, defaultVal bool) bool {
	if v, err := c.GetBool(key); err != nil {
		return defaultVal
	} else {
		return v
	}
}

// GetBoolArray returns the []bool value for a given key.
func (c *Config) GetBoolArray(key string) ([]bool, error) {
	v, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	t, ok := v.([]interface{})
	if !ok {
		return nil, errors.New("value of `" + key + "` is not a boolean list")
	}

	vArr := make([]bool, 0, len(t))
	for _, vv := range t {
		switch vvv := vv.(type) {
		case bool:
			vArr = append(vArr, vvv)
		case string:
			if vvvv, err := strconv.ParseBool(vvv); err == nil {
				vArr = append(vArr, vvvv)
			} else {
				return nil, errors.New("some value in key `" + key + "` is not boolean")
			}
		default:
			return nil, errors.New("some value in key `" + key + "` is not boolean")
		}
	}
	return vArr, nil
}

// GetDefaultBoolArray returns the []bool value for a given key.
// if error occur, return defaultVal
func (c *Config) GetDefaultBoolArray(key string, defaultVal []bool) []bool {
	if v, err := c.GetBoolArray(key); err != nil {
		return defaultVal
	} else {
		return v
	}
}

// GetFloat returns the float64 value for a given key.
func (c *Config) GetFloat(key string) (float64, error) {
	v, err := c.Get(key)
	if err != nil {
		return 0, err
	}

	switch vv := v.(type) {
	case int:
		return float64(vv), nil
	case float64:
		return vv, nil
	case string:
		if vvv, err := strconv.ParseFloat(vv, 64); err == nil {
			return vvv, nil
		} else {
			return 0, errors.New("value of `" + key + "` is not float64")
		}
	default:
		return 0, errors.New("value of `" + key + "` is not float64")
	}
}

// GetDefaultFloat returns the float64 value for a given key.
// if error occur, return defaultVal
func (c *Config) GetDefaultFloat(key string, defaultVal float64) float64 {
	if v, err := c.GetFloat(key); err != nil {
		return defaultVal
	} else {
		return v
	}
}

// GetFloatArray returns the []float64 value for a given key.
func (c *Config) GetFloatArray(key string) ([]float64, error) {
	v, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	t, ok := v.([]interface{})
	if !ok {
		return nil, errors.New("value of `" + key + "` is not a boolean list")
	}

	vArr := make([]float64, 0, len(t))
	for _, vv := range t {
		switch vvv := vv.(type) {
		case int:
			vArr = append(vArr, float64(vvv))
		case float64:
			vArr = append(vArr, vvv)
		case string:
			if vvvv, err := strconv.ParseFloat(vvv, 64); err == nil {
				vArr = append(vArr, vvvv)
			} else {
				return nil, errors.New("some value in key `" + key + "` is not float64")
			}
		default:
			return nil, errors.New("some value in key `" + key + "` is not float64")
		}
	}
	return vArr, nil
}

// GetDefaultFloatArray returns the []float64 value for a given key.
// if error occur, return defaultVal
func (c *Config) GetDefaultFloatArray(key string, defaultVal []float64) []float64 {
	if v, err := c.GetFloatArray(key); err != nil {
		return defaultVal
	} else {
		return v
	}
}

// GetMap returns the map[string]interface{} value for a given key.
func (c *Config) GetMap(key string) (map[string]interface{}, error) {
	v, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	t, ok := v.(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("value of `" + key + "` is not a map")
	}

	vMap := make(map[string]interface{})
	for kk, vv := range t {
		vMap[kk.(string)] = vv
	}
	return vMap, nil
}

// GetDefaultMap returns the map[string]interface{} value for a given key.
// if error occur, return defaultVal
func (c *Config) GetDefaultMap(key string, defaultVal map[string]interface{}) map[string]interface{} {
	if v, err := c.GetMap(key); err != nil {
		return defaultVal
	} else {
		return v
	}
}

// GetSubKeys returns the subkey array of a given key.
// Support multi-level key which concat with '.'.
func (c *Config) GetSubKeys(key string) ([]string, error) {
	v, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	vv, ok := v.(map[interface{}]interface{})
	if !ok {
		return []string{}, nil
	}

	keys := make([]string, 0, len(vv))
	for k, _ := range vv {
		if key, ok := k.(string); ok {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// Len returns the value length of a given key.
// Support multi-level key which concat with '.'.
func (c *Config) Len(key string) (int, error) {
	v, err := c.Get(key)
	if err != nil {
		return 0, err
	}
	switch vv := v.(type) {
	case map[interface{}]interface{}:
		return len(vv), nil
	case []interface{}:
		return len(vv), nil
	default:
		return 0, nil
	}
}

// Get returns the interface{} value for a given key.
// Support multi-level key which concat with '.'.
func (c *Config) Get(key string) (interface{}, error) {
	if len(key) == 0 {
		return nil, errors.New("key should not be empty")
	}
	if key[0] == '[' {
		return nil, errors.New("wrong key format")
	}

	// 每级使用.分隔
	tKeyArr := strings.Split(key, c.Delimiter)
	keyArr := make([]interface{}, 0, len(tKeyArr))

	// 将[数字]形式的进行分隔
	for _, v := range tKeyArr {
		indexs := make([]uint16, 0)
		for {
			if !strings.HasSuffix(v, "]") {
				keyArr = append(keyArr, v)
				goto next
			}
			startPos := strings.LastIndex(v, "[")
			if startPos == -1 {
				keyArr = append(keyArr, v)
				goto next
			}
			indexStr := v[startPos+1 : len(v)-1]
			index, err := strconv.ParseUint(indexStr, 10, 16)
			if err != nil { // 非uint16
				keyArr = append(keyArr, v)
				goto next
			}

			indexs = append(indexs, uint16(index))
			v = v[:startPos]
		}

	next:
		for _, index := range indexs {
			keyArr = append(keyArr, index)
		}
		continue
	}

	var pKey, cKey string
	var tNode interface{} = c.cfgData
	lasti := len(keyArr) - 1

	for i, v := range keyArr {
		switch key := v.(type) {
		case string:
			if pKey == "" {
				cKey = key
			} else {
				cKey = pKey + c.Delimiter + key
			}

			tMap, ok := tNode.(map[interface{}]interface{})
			if !ok {
				return nil, errors.New("key `" + pKey + "` is not a map")
			}

			// 检测类型，必须为map或slice
			switch t := tMap[key].(type) {
			case map[interface{}]interface{}:
				tNode = interface{}(t)
			case []interface{}:
				tNode = interface{}(t)
			case nil:
				return nil, errors.New("key `" + cKey + "` is not exists")
			default:
				if i == lasti {
					// path最后一个部分
					return t, nil
				}
				return nil, errors.New("key `" + cKey + "` is not a map or slice")
			}

		case uint16:
			cKey = pKey + fmt.Sprint("[%d]", key)

			tSlice, ok := tNode.([]interface{})
			if !ok {
				return nil, errors.New("key `" + pKey + "` is not a map")
			}

			// 检测类型，必须为slice
			switch t := tSlice[key].(type) {
			case map[interface{}]interface{}:
				tNode = interface{}(t)
			case []interface{}:
				tNode = interface{}(t)
			case nil:
				return nil, errors.New("key `" + cKey + "` is not exists")
			default:
				if i == lasti {
					// path最后一个部分
					return t, nil
				}
				return nil, errors.New("key `" + cKey + "` is not a map or slice")
			}
		}
		pKey = cKey
	}

	return tNode, nil
}
