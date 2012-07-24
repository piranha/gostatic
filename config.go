// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"strings"
	"reflect"
	"time"
	"fmt"
)

type PageConfig struct {
	Title string `default:"Unknown Title"`
	Type  string `default:"page"`
	Tags  []string
	Date  time.Time
	Other map[string]string
}

var DATEFORMATS = []string{
	"2006-01-02 15:04:05 -07",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02 15",
	"2006-01-02",
	"06-01-02",
}

func NewPageConfig() *PageConfig {
	return &PageConfig{Other: make(map[string]string)}
}

func (cfg *PageConfig) ParseLine(line string, s *reflect.Value) {
	// Skip empty lines and comments
	line = strings.SplitN(line, "//", 2)[0]
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return
	}

	// Split line in actual name and value
	parts := strings.SplitN(line, ":", 2)
	key := strings.ToUpper(parts[0][0:1]) + strings.TrimSpace(parts[0][1:])
	value := strings.TrimSpace(parts[1])

	cfg.SetValue(key, value, s)
}

func (cfg *PageConfig) SetValue(key string, value string, s *reflect.Value) {
	// put unknown fields into a map
	if _, ok := s.Type().FieldByName(key); !ok {
		cfg.Other[key] = strings.TrimSpace(value)
		return
	}

	// Set value
	f := s.FieldByName(key)
	switch f.Interface().(type) {
	default:
		errhandle(fmt.Errorf("Unknown type of field %s", key))
	case string:
		f.SetString(value)
	case []string:
		values := strings.Split(value, ",")
		for i, v := range values {
			values[i] = strings.TrimSpace(v)
		}
		f.Set(reflect.ValueOf(values))
	case time.Time:
		var t time.Time
		var err error
		for _, fmt := range DATEFORMATS {
			t, err = time.Parse(fmt, value)
			if err == nil {
				break
			}
		}
		errhandle(err)
		f.Set(reflect.ValueOf(t))
	}
}

func ParseConfig(source string) *PageConfig {
	cfg := NewPageConfig()
	s := reflect.ValueOf(cfg).Elem()

	// Set default values
	t := s.Type()
	for i := 0; i < s.NumField(); i++ {
		def := t.Field(i).Tag.Get("default")
		if len(def) != 0 {
			cfg.SetValue(t.Field(i).Name, def, &s)
		}
	}

	for _, line := range strings.Split(source, "\n") {
		cfg.ParseLine(line, &s)
	}

	return cfg
}
