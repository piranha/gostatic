// (c) 2012 Alexander Solovyov
// under terms of ISC license

package main

import (
	"strings"
	// "fmt"
	"reflect"
)

type PageConfig struct {
	Title string `default:"unknown title"`
	Type  string `default:"page"`
	Tags  []string
	Other map[string]string
}

func (cfg *PageConfig) ParseLine(line string, elemptr *reflect.Value) {
	var s reflect.Value
	if elemptr != nil {
		s = *elemptr
	} else {
		s = reflect.ValueOf(cfg).Elem()
	}

	// Cleanup line
	line = strings.SplitN(line, "//", 2)[0]
	line = strings.TrimSpace(line)

	// Skip empty lines and comments
	if len(line) == 0 {
		return
	}

	// Split line in actual name and value
	parts := strings.SplitN(line, ":", 2)
	name := strings.ToUpper(parts[0][0:1]) + strings.TrimSpace(parts[0][1:])
	value := strings.TrimSpace(parts[1])

	// put unknown fields into a map
	if _, ok := s.Type().FieldByName(name); !ok {
		cfg.Other[name] = strings.TrimSpace(value)
		return
	}

	// Set value
	f := s.FieldByName(name)
	switch f.Kind() {
	case reflect.String:
		f.SetString(value)
	case reflect.Slice:
		values := strings.Split(value, ",")
		for i, v := range values {
			values[i] = strings.TrimSpace(v)
		}
		f.Set(reflect.ValueOf(values))
	}
}

func ParseConfig(source string) *PageConfig {
	cfg := &PageConfig{}
	s := reflect.ValueOf(cfg).Elem()

	// Set default values
	t := s.Type()
	for i := 0; i < s.NumField(); i++ {
		def := t.Field(i).Tag.Get("default")
		f := s.Field(i)

		if len(f.String()) == 0 && len(def) > 0 {
			f.SetString(def)
		}
	}

	for _, line := range strings.Split(source, "\n") {
		cfg.ParseLine(line, &s)
	}

	return cfg
}
