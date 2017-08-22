package config

import (
	//"os"
	"testing"
)

func TestJsonBool(t *testing.T) {
	config, err := NewConfig("json", "my.json")
	if err != nil {
		t.Error(err)
		return
	}
	val, err := config.Bool("IsOpen")
	if err != nil {
		t.Error(err)
		return
	}
	if val != false {
		t.Error("get data failed.")
	}
}

func TestJsonInt(t *testing.T) {
	config, err := NewConfig("json", "my.json")
	if err != nil {
		t.Error(err)
		return
	}
	val, err := config.Int("num")
	if err != nil {
		t.Error(err)
		return
	}
	if val != 5 {
		t.Error("Get int failed.")
	}
}

func TestJsonFloat(t *testing.T) {
	config, err := NewConfig("json", "my.json")
	if err != nil {
		t.Error(err)
		return
	}
	val, err := config.Float("float")
	if err != nil {
		t.Error(err)
		return
	}
	if val != 3.1415 {
		t.Error("Get float failed.")
	}
}
func TestJsonString(t *testing.T) {
	config, err := NewConfig("json", "my.json")
	if err != nil {
		t.Error(err)
		return
	}
	val := config.String("addr")
	if val != "127.0.0.1" {
		t.Error("Get string failed.")
	}
}

func TestJsonStrings(t *testing.T) {
	config, err := NewConfig("json", "my.json")
	if err != nil {
		t.Error(err)
		return
	}
	val := config.Strings("addrs")
	if len(val) != 2 {
		t.Error("Get strings failed.")
	}
	if val[0] != "127.0.0.1" && val[1] != "192.168.1.1" {
		t.Error("Get strings failed.")
	}
}

func TestJsonSection(t *testing.T) {
	config, err := NewConfig("json", "my.json")
	if err != nil {
		t.Error(err)
		return
	}
	var val string
	val = config.String("mysql::addr")
	if val != "127.0.0.1" {
		t.Error("get section data failed.")
	}
	valInt, err := config.Int("mysql::port")
	if valInt != 3306 || err != nil {
		t.Error("Get int data failed.")
	}
}

func TestSaveJsonFile(t *testing.T) {
	config, err := NewConfig("json", "my.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.SaveConfigFile("test.json")
}

func TestJsonSet(t *testing.T) {
	config, err := NewConfig("json", "my.json")
	if err != nil {
		t.Error(err)
		return
	}
	err = config.Set("listenport", "19996")
	if err != nil {
		t.Error("Set failed.")
	}
	val, err := config.Int("listenport")
	if err != nil || val != 19996 {
		t.Error("Get failed,")
	}
}

func TestSectionJsonSet(t *testing.T) {
	config, err := NewConfig("json", "my.json")
	if err != nil {
		t.Error(err)
		return
	}
	err = config.Set("mysql::tablename", "familyinfo")
	if err != nil {
		t.Error("Set failed.")
	}
	val := config.String("mysql::tablename")

	if val != "familyinfo" {
		t.Error("Get failed,")
	}
}
