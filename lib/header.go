// (c) 2012 Alexander Solovyov
// under terms of ISC license

package gostatic

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type PageHeader struct {
	Title string
	Tags  []string
	Date  time.Time
	Hide  bool
	Other map[string]string
}

var DATEFORMATS = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05Z-07:00", // ISO 8601
	"2006-01-02T15:04:05Z",       // ISO 8601
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05 -07",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02 15",
	"2006-01-02",
	"06-01-02",
}

func NewPageHeader() *PageHeader {
	return &PageHeader{Other: make(map[string]string)}
}

func (cfg *PageHeader) ParseLine(line string, s *reflect.Value) {
	// Skip empty lines
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return
	}

	// Split line in actual name and value
	bits := TrimSplitN(line, ":", 2)
	if len(bits) < 2 {
		errhandle(fmt.Errorf("could not parse '%s' as 'key: value' string",
			line))
		return
	}

	key := strings.ToUpper(bits[0][0:1]) + bits[0][1:]
	cfg.SetValue(key, bits[1], s)
}

var FalsyValues = map[string]bool{
	"false": true,
	"False": true,
	"FALSE": true,
	"f":     true,
}

func (cfg *PageHeader) SetValue(key string, value string, s *reflect.Value) {
	// put unknown fields into a map
	if _, ok := s.Type().FieldByName(key); !ok {
		cfg.Other[Capitalize(key)] = strings.TrimSpace(value)
		return
	}

	// Set value
	f := s.FieldByName(key)
	switch typ := f.Interface().(type) {
	default:
		errhandle(fmt.Errorf("unknown type of field %s (is type '%v')", key, typ))
	case string:
		f.SetString(value)
	case bool:
		_, ok := FalsyValues[value]
		f.SetBool(!ok)
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

func ParseHeader(source string) *PageHeader {
	cfg := NewPageHeader()

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

func ParseYamlHeader(source string) *PageHeader {
	cfg := NewPageHeader()

	s := reflect.ValueOf(cfg).Elem()

	// Set default values
	t := s.Type()
	for i := 0; i < s.NumField(); i++ {
		def := t.Field(i).Tag.Get("default")
		if len(def) != 0 {
			cfg.SetValue(t.Field(i).Name, def, &s)
		}
	}

	m := make(map[string]string)
	yaml.Unmarshal([]byte(source), &m)

	for key, value := range m {
		key := strings.ToUpper(key[0:1]) + key[1:]
		cfg.SetValue(key, value, &s)
	}

	return cfg
}
