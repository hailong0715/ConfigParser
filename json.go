package config

import (
	"encoding/json"
	"errors"
	//"fmt"
	"io/ioutil"
	"os"
	//"reflect"
	"strconv"
	"strings"
	"sync"
)

type JsonConfig struct {
}

func (jc *JsonConfig) Parse(filename string) (Configer, error) {
	return jc.parseFile(filename)
}
func (jc *JsonConfig) parseData(data []byte) (*JsonCfgContainer, error) {
	cfg := &JsonCfgContainer{
		data: make(map[string]interface{}),
	}
	err := json.Unmarshal(data, &cfg.data)
	if err != nil {
		return nil, err
	}
	return cfg, err

}

func (jc *JsonConfig) ParseData(data []byte) (Configer, error) {
	cfg := &JsonCfgContainer{
		data: make(map[string]interface{}),
	}
	err := json.Unmarshal(data, &cfg.data)
	if err != nil {
		return nil, err
	}
	return cfg, err

}

func (jc *JsonConfig) parseFile(filename string) (*JsonCfgContainer, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return jc.parseData(data)
}

type JsonCfgContainer struct {
	data map[string]interface{}
	sync.RWMutex
}

//给配置文件的某个字段设置值，支持sec::key的方式选择key值
func (c *JsonCfgContainer) Set(key, val string) error {
	c.Lock()
	defer c.Unlock()
	sectionKey := strings.Split(key, "::")
	if len(sectionKey) == 2 {
		var section, k string
		section = sectionKey[0]
		k = sectionKey[1]
		_, ok := c.data[section]
		if !ok {
			c.data[section] = make(map[string]interface{})
		}
		if vv, ok := c.data[section].(map[string]interface{}); ok {
			vv[k] = val
		}
	}
	c.data[key] = val
	return nil
}

//返回指定key的Val值得string格式，key支持sec::key的方式
func (c *JsonCfgContainer) String(key string) string {
	c.Lock()
	defer c.Unlock()
	val := c.getdata(key)
	if val != nil {
		return ToString(val)
	}
	return ""
}

//返回指定key值得Val的切片
func (c *JsonCfgContainer) Strings(key string) []string {
	c.Lock()
	defer c.Unlock()
	var resp []string
	val := c.getdata(key)
	switch vv := val.(type) {
	case string:
		resp = append(resp, vv)
	case []string:
		resp = vv
	case []interface{}:
		for _, vvv := range vv {
			resp = append(resp, vvv.(string))
		}
	default:
		resp = nil
	}
	return resp
}

//返回指定key对应val的int值
func (c *JsonCfgContainer) Int(key string) (int, error) {
	c.Lock()
	defer c.Unlock()
	val := c.getdata(key)
	switch vv := val.(type) {
	case string:
		return strconv.Atoi(vv)
	case int:
		return vv, nil
	case float64:
		return int(vv), nil
	case float32:
		return int(vv), nil
	}
	return 0, errors.New("val is not valid")
}

//返回指定key对应val得int64值
func (c *JsonCfgContainer) Int64(key string) (int64, error) {
	c.Lock()
	defer c.Unlock()
	val := c.getdata(key)
	if v, ok := val.(string); ok {
		return strconv.ParseInt(v, 10, 64)
	}
	return 0, errors.New("val is not valid")
}
func (c *JsonCfgContainer) Bool(key string) (bool, error) {
	c.Lock()
	defer c.Unlock()
	return ParseBool(c.getdata(key))
}
func (c *JsonCfgContainer) Float(key string) (float64, error) {
	c.Lock()
	defer c.Unlock()
	val := c.getdata(key)

	switch vv := val.(type) {
	case float64:
		return vv, nil
	case float32:
		return float64(vv), nil
	case string:
		return strconv.ParseFloat(vv, 64)
	}
	return 0, errors.New("val is not valid")
}

//返回指定key的val的值，若key对应的val为空，给该key对应的val设置我defaultVal
func (c *JsonCfgContainer) DefaultString(key, defaultVal string) string {
	v := c.String(key)
	if v != "" {
		return v
	}
	return defaultVal
}
func (c *JsonCfgContainer) DefaultStrings(key string, defaultVals []string) []string {
	v := c.Strings(key)
	if v != nil {
		return v
	}
	return defaultVals
}
func (c *JsonCfgContainer) DefaultInt(key string, defaultVal int) int {
	v, err := c.Int(key)
	if err != nil {
		return defaultVal
	}
	return v
}
func (c *JsonCfgContainer) DefaultInt64(key string, defaultVal int64) int64 {
	v, err := c.Int64(key)
	if err != nil {
		return defaultVal
	}
	return v
}
func (c *JsonCfgContainer) DefaultBool(key string, defaultVal bool) bool {
	v, err := c.Bool(key)
	if err != nil {
		return defaultVal
	}
	return v
}
func (c *JsonCfgContainer) DefaultFloat(key string, defaultVal float64) float64 {
	v, err := c.Float(key)
	if err != nil {
		return defaultVal
	}
	return v
}

//返回给定key的val，并将val转型为interface{}类型
func (c *JsonCfgContainer) GetInerfaceVal(key string) (interface{}, error) {
	if val, ok := c.data[strings.ToLower(key)]; ok {
		return val, nil
	}
	return nil, errors.New("get interface data failed.")
}

//返回某个section下的全部配置
func (c *JsonCfgContainer) GetSection(section string) (map[string]string, error) {
	c.Lock()
	defer c.Unlock()
	var secmap = make(map[string]string)
	if v, ok := c.data[strings.ToLower(section)]; ok {
		for k, val := range v.(map[string]interface{}) {
			secmap[k] = ToString(val)
		}
		return secmap, nil
	}

	return nil, errors.New("section not exist")
}

//将配置信息保存到文件
func (c *JsonCfgContainer) SaveConfigFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(c.data)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *JsonCfgContainer) getdata(key string) interface{} {
	sectionKey := strings.Split(strings.ToLower(key), "::")
	if len(sectionKey) >= 2 {
		value, ok := c.data[sectionKey[0]]
		if !ok {
			return nil
		}
		if v, ok := value.(map[string]interface{}); ok {
			if value, ok = v[sectionKey[1]]; !ok {
				return nil
			}
			return value
		} else {
			return nil
		}

	}

	if val, ok := c.data[key]; ok {
		return val
	}
	return nil
}
func init() {
	Register("json", &JsonConfig{})
}
