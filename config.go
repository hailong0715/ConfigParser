package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

//configer 接口，提供操作配置文件的一系列接口
type Configer interface {
	Set(key, val string) error       //给配置文件的某个字段设置值，支持sec::key的方式选择key值
	String(key string) string        //返回指定key的Val值得string格式，key支持sec::key的方式
	Strings(key string) []string     //返回指定key值得Val的切片
	Int(key string) (int, error)     //返回指定key对应val的int值
	Int64(key string) (int64, error) //返回指定key对应val得int64值
	Bool(key string) (bool, error)
	Float(key string) (float64, error)
	DefaultString(key, defaultVal string) string //返回指定key的val的值，若key对应的val为空，给该key对应的val设置我defaultVal
	DefaultStrings(key string, defaultVals []string) []string
	DefaultInt(key string, defaultVal int) int
	DefaultInt64(key string, defaultVal int64) int64
	DefaultBool(key string, defaultVal bool) bool
	DefaultFloat(key string, defaultVal float64) float64
	GetInerfaceVal(key string) (interface{}, error)       //返回给定key的val，并将val转型为interface{}类型
	GetSection(section string) (map[string]string, error) //返回某个section下的全部配置
	SaveConfigFile(filename string) error                 //将配置信息保存到文件
	GetCfgData() interface{}
}

//Configer的适配器接口，将配置文件或者数据解析，并返回一个Configer的对象
type Config interface {
	Parse(filename string) (Configer, error)
	ParseData(data []byte) (Configer, error)
}

//适配器，保存所有注册的config适配器，key值为配置文件类型，比如ini,xml,conf等
//val 文件类型对用的适配器
var adapters = make(map[string]Config)

func Register(name string, adapter Config) {
	if adapter == nil {
		panic("Config: adapter can not be empty")
	}
	if _, ok := adapters[name]; ok {
		panic("CConfig: adapter for name:" + name + "is only allowed to register once")
	}
	adapters[name] = adapter
}

//返回一个新的Configuer对象，adaptername是文件类型 ini/xml/conf ...
func NewConfig(adaptername, filename string) (Configer, error) {
	adapter, ok := adapters[adaptername]
	if !ok {
		return nil, errors.New("unknown adaptername" + adaptername + ", should register first,then use it")
	}
	return adapter.Parse(filename)
}

func NewConfigData(adapterName string, data []byte) (Configer, error) {
	adapter, ok := adapters[adapterName]
	if !ok {
		return nil, errors.New("unknown adaptername" + adapterName + ", should register first,then use it")
	}
	return adapter.ParseData(data)
}

func ToString(in interface{}) string {
	switch out := in.(type) {
	case time.Time:
		return out.Format("2016-02-05")
	case string:
		return out
	case fmt.Stringer:
		return out.String()
	case error:
		return out.Error()
	case float64:
		return strconv.FormatFloat(out, 'f', -1, 64)
	}
	if val := reflect.ValueOf(in); val.Kind() == reflect.String {
		return val.String()
	}
	return fmt.Sprintf("%s", in)
}

func ParseBool(in interface{}) (out bool, err error) {
	if in != nil {
		switch v := in.(type) {
		case bool:
			return v, nil
		case string:
			switch v {
			case "1", "t", "T", "true", "TRUE", "True", "YES", "yes", "Yes", "Y", "y", "ON", "on", "On":
				return true, nil
			case "0", "f", "F", "false", "FALSE", "False", "NO", "no", "No", "N", "n", "OFF", "off", "Off":
				return false, nil
			}
		case int8, int32, int64:
			str := fmt.Sprintf("%s", v)
			if str == "1" {
				return true, nil
			} else if str == "0" {
				return false, nil
			}
		case float64:
			if v == 1 {
				return true, nil
			} else if v == 0 {
				return false, nil
			}
		}
		return false, fmt.Errorf("parsing %q:invalid syntax", in)
	}
	return false, fmt.Errorf("parsing <nil>: invalid syntax")
}
