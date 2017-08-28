package config

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var (
	DEFAULT_SECTION = "default" //配置文件中的某些不在某个sec下，默认将其放在defaultSection下
	NUM_COMMENT     = []byte{'#'}
	SEM_COMMENT     = []byte{';'} //ini文件注释
	EMPTY           = []byte{}
	EQUAL           = []byte{'='}
	QUOTE           = []byte{'"'}
	SEC_START       = []byte{'['}
	SEC_END         = []byte{']'}
	LINE_BREAK      = "\n"
)

type IniConfig struct {
}

type IniConfigContainer struct {
	data       map[string]map[string]string //保存配置数据，sec-->key:val
	secComment map[string]string            //保存注释 sec-->comment
	keyComment map[string]string            // key --> comment 某个配置的注释
	sync.RWMutex
}

func (ini *IniConfig) Parse(filename string) (Configer, error) {
	return ini.parseFile(filename)
}

func (ini *IniConfig) parseFile(filename string) (*IniConfigContainer, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return ini.parseData(filepath.Dir(filename), data)
}

func (ini *IniConfig) parseData(dir string, data []byte) (*IniConfigContainer, error) {
	cfg := &IniConfigContainer{
		data:       make(map[string]map[string]string),
		secComment: make(map[string]string),
		keyComment: make(map[string]string),
		RWMutex:    sync.RWMutex{},
	}
	cfg.Lock()
	defer cfg.Unlock()

	var comment bytes.Buffer
	buf := bufio.NewReader(bytes.NewBuffer(data))

	//由于 unicode编码的文档在windows系统下会被自动在文档头部加入三个字节的BOM，因此解析前
	//需要判断前三个字节是不是BOM，是的话去掉BOM，否则无法解析
	head, err := buf.Peek(3)
	if err == nil && head[0] == 239 && head[1] == 187 && head[2] == 191 {
		for i := 0; i < 3; i++ {
			buf.ReadByte()
		}
	}

	section := DEFAULT_SECTION
	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		line = bytes.TrimSpace(line)
		if bytes.Equal(line, EMPTY) {
			//空行
			continue
		}

		//读取注释
		var bcomment []byte

		switch {
		case bytes.HasPrefix(line, NUM_COMMENT):
			bcomment = NUM_COMMENT
		case bytes.HasPrefix(line, SEM_COMMENT):
			bcomment = SEM_COMMENT
		}
		if bcomment != nil {
			line = bytes.TrimLeft(line, string(bcomment))
			if comment.Len() > 0 {
				comment.WriteByte('\n')
			}
			comment.Write(line)
			continue
		}

		//读取section
		if bytes.HasPrefix(line, SEC_START) && bytes.HasSuffix(line, SEC_END) {
			section = string(bytes.ToLower(line[1 : len(line)-1]))
			if comment.Len() > 0 {
				cfg.secComment[section] = comment.String()
				comment.Reset()
			}
			if _, ok := cfg.data[section]; !ok {
				cfg.data[section] = make(map[string]string)
			}
			continue
		}

		if _, ok := cfg.data[section]; !ok {
			cfg.data[section] = make(map[string]string)
		}

		//解析配置项
		keyValue := bytes.SplitN(line, EQUAL, 2)
		key := string(bytes.TrimSpace(keyValue[0]))
		key = strings.ToLower(key)

		//判断文件是否包含其他配置文件，是的话先解析被包含的配置文件 include "other.conf"
		if len(keyValue) == 1 && strings.HasPrefix(key, "include") {
			includefiles := strings.Fields(key)
			if includefiles[0] == "include" && len(includefiles) == 2 {
				otherfile := strings.Trim(includefiles[1], "\"")
				if !filepath.IsAbs(otherfile) {
					otherfile = filepath.Join(dir, otherfile)
				}

				i, err := ini.parseFile(otherfile)
				if err != nil {
					return nil, err
				}

				for sec, dt := range i.data {
					if _, ok := cfg.data[sec]; !ok {
						cfg.data[sec] = make(map[string]string)
					}
					for k, v := range dt {
						cfg.data[sec][k] = v
					}
				}

				for sec, comm := range i.secComment {
					cfg.secComment[sec] = comm
				}

				for k, comm := range i.keyComment {
					cfg.keyComment[k] = comm
				}
				continue
			}
		}

		if len(keyValue) != 2 {
			return nil, errors.New("read content error," + string(line) + " format should be key = value")
		}
		val := bytes.TrimSpace(keyValue[1])
		if bytes.HasPrefix(val, QUOTE) {
			val = bytes.Trim(val, `"`)
		}
		cfg.data[section][key] = string(val)
		if comment.Len() > 0 {
			cfg.keyComment[section+"."+key] = comment.String()
			comment.Reset()
		}
	}
	return cfg, nil
}

func (ini *IniConfig) ParseData(data []byte) (Configer, error) {
	dir := "tmp"
	currentUser, err := user.Current()
	if err == nil {
		dir = "tmp-" + currentUser.Username
	}
	dir = filepath.Join(os.TempDir(), dir)
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	return ini.parseData(dir, data)
}

func (c *IniConfigContainer) Bool(key string) (bool, error) {
	return ParseBool(c.getdata(key))
}

func (c *IniConfigContainer) DefaultBool(key string, defaultVal bool) bool {
	v, err := c.Bool(key)
	if err != nil {
		return defaultVal
	}
	return v
}

func (c *IniConfigContainer) DefaultInt(key string, defaultVal int) int {
	v, err := c.Int(key)
	if err != nil {
		return defaultVal
	}
	return v
}

func (c *IniConfigContainer) Int(key string) (int, error) {
	return strconv.Atoi(c.getdata(key))
}

func (c *IniConfigContainer) Int64(key string) (int64, error) {
	return strconv.ParseInt(c.getdata(key), 10, 64)
}

func (c *IniConfigContainer) DefaultInt64(key string, defaultval int64) int64 {
	v, err := c.Int64(key)
	if err != nil {
		return defaultval
	}
	return v
}

func (c *IniConfigContainer) Float(key string) (float64, error) {
	return strconv.ParseFloat(c.getdata(key), 64)
}
func (c *IniConfigContainer) DefaultFloat(key string, defaultval float64) float64 {
	v, err := c.Float(key)
	if err != nil {
		return defaultval
	}
	return v
}

func (c *IniConfigContainer) String(key string) string {
	return c.getdata(key)
}

func (c *IniConfigContainer) DefaultString(key string, defaultval string) string {
	v := c.String(key)
	if v == "" {
		return defaultval
	}
	return v
}
func (c *IniConfigContainer) Strings(key string) []string {
	v := c.String(key)
	if v == "" {
		return nil
	}
	v = strings.Replace(v, "\"", "", -1)
	str := strings.Split(v, ";")
	return str
}

func (c *IniConfigContainer) DefaultStrings(key string, defaultval []string) []string {
	v := c.Strings(key)
	if v == nil {
		return defaultval
	}
	return v
}

func (c *IniConfigContainer) GetSection(section string) (map[string]string, error) {
	if v, ok := c.data[section]; ok {
		return v, nil
	}
	return nil, errors.New("section not exist")
}

func (c *IniConfigContainer) SaveConfigFile(filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	getCommentStr := func(section, key string) string {
		var (
			comment string
			ok      bool
		)
		if len(key) == 0 {
			comment, ok = c.secComment[section]
		} else {
			comment, ok = c.keyComment[key]
		}

		if ok {
			if len(comment) == 0 || len(strings.TrimSpace(comment)) == 0 {
				return string(NUM_COMMENT)
			}

			//增加注释头"#"
			prefix := string(NUM_COMMENT)
			return prefix + strings.Replace(comment, LINE_BREAK, LINE_BREAK+prefix, -1)
		}
		return ""
	}

	buf := bytes.NewBuffer(nil)
	//先保存defaultsection下的默认全局配置
	if dt, ok := c.data[DEFAULT_SECTION]; ok {
		for key, val := range dt {
			if key != " " {
				//写入配置项注释
				if v := getCommentStr(DEFAULT_SECTION, key); len(v) > 0 {
					if _, err = buf.WriteString(v + LINE_BREAK); err != nil {
						return err
					}
				}
			}

			//写入配置项
			if _, err = buf.WriteString(key + string(EQUAL) + val + LINE_BREAK); err != nil {
				return err
			}
		}
	}

	//保存section下的配置
	for section, dt := range c.data {
		if section != DEFAULT_SECTION {
			if v := getCommentStr(section, ""); len(v) > 0 {
				if _, err = buf.WriteString(v + LINE_BREAK); err != nil {
					return err
				}
			}

			for key, val := range dt {
				if key != " " {
					if v := getCommentStr(section, key); len(v) > 0 {
						if _, err = buf.WriteString(v + LINE_BREAK); err != nil {
							return err
						}
					}
				}

				if _, err = buf.WriteString(key + string(EQUAL) + val + LINE_BREAK); err != nil {
					return err
				}
			}
			if _, err = buf.WriteString(LINE_BREAK); err != nil {
				return err
			}
		}
	}
	_, err = buf.WriteTo(f)
	return err
}

func (c *IniConfigContainer) Set(key, value string) error {
	c.Lock()
	defer c.Unlock()
	if len(key) == 0 {
		return errors.New("Key can not be empty")
	}

	var (
		section, k string
		sectionKey = strings.Split(strings.ToLower(key), "::")
	)

	if len(sectionKey) == 2 {
		section = sectionKey[0]
		k = sectionKey[1]
	} else {
		section = DEFAULT_SECTION
		k = sectionKey[0]
	}
	if _, ok := c.data[section]; !ok {
		c.data[section] = make(map[string]string)
	}
	c.data[section][k] = value
	return nil
}

func (c *IniConfigContainer) GetInerfaceVal(key string) (interface{}, error) {
	if v, ok := c.data[strings.ToLower(key)]; ok {
		return v, nil
	}
	return nil, errors.New("key not exist")
}

func (c *IniConfigContainer) getdata(key string) string {
	if len(key) == 0 {
		return ""
	}
	c.Lock()
	defer c.Unlock()
	key = strings.ToLower(key)
	var (
		section, k string
		sectionKey = strings.Split(key, "::")
	)
	if len(sectionKey) == 2 {
		section = sectionKey[0]
		k = sectionKey[1]
	} else {
		section = DEFAULT_SECTION
		k = sectionKey[0]
	}
	if v, ok := c.data[section]; ok {
		if vv, ok := v[k]; ok {
			return vv
		}
	}
	return ""
}
func (c *IniConfigContainer) GetCfgData() interface{} {
	return c.data
}
func init() {
	Register("ini", &IniConfig{})
}
