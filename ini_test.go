package config

import (
	//"os"
	"testing"
)

func TestBool(t *testing.T) {
	config, err := NewConfig("ini", "my.ini")
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

func TestInt(t *testing.T) {
	config, err := NewConfig("ini", "my.ini")
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

func TestFloat(t *testing.T) {
	config, err := NewConfig("ini", "my.ini")
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
func TestString(t *testing.T) {
	config, err := NewConfig("ini", "my.ini")
	if err != nil {
		t.Error(err)
		return
	}
	val := config.String("addr")
	if val != "127.0.0.1" {
		t.Error("Get string failed.")
	}
}

func TestStrings(t *testing.T) {
	config, err := NewConfig("ini", "my.ini")
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
