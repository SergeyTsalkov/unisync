package ini

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type typeMapFn func(string) (reflect.Value, error)

type Parser struct {
	typeMap map[string]typeMapFn
}

type Unmarshaler interface {
	UnmarshalINI([]byte) error
}

var unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

func New() *Parser {
	return &Parser{
		typeMap: map[string]typeMapFn{},
	}
}

func (p *Parser) AddTypeMap(key string, fn typeMapFn) {
	p.typeMap[key] = fn
}

func (p *Parser) Unmarshal(data []byte, ptr any) error {
	fieldMap, err := p.makeFieldMap(ptr)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		key, value, valid := strings.Cut(line, "=")
		if !valid {
			continue
		}

		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)

		// commented out field?
		if strings.ContainsAny(key, "#; ") {
			continue
		}

		v, ok := fieldMap[key]
		if !ok {
			return fmt.Errorf("%v <-- invalid setting", line)
		}

		err = p.setValue(v, value)
		if err != nil {
			return fmt.Errorf("%v <-- %v", line, err)
		}
	}

	return nil
}

func (p *Parser) setValue(v reflect.Value, str string) error {
	if v.CanConvert(unmarshalerType) {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		unmarshaler := (v.Interface()).(Unmarshaler)
		return unmarshaler.UnmarshalINI([]byte(str))
	}
	if ptr := v.Addr(); ptr.CanConvert(unmarshalerType) {
		unmarshaler := (ptr.Interface()).(Unmarshaler)
		return unmarshaler.UnmarshalINI([]byte(str))
	}

	if v.Type().Kind() == reflect.Slice {
		setVal, err := p.getValue(str, v.Type().Elem())
		if err != nil {
			return err
		}

		v.Set(reflect.Append(v, setVal))
	} else {
		setVal, err := p.getValue(str, v.Type())
		if err != nil {
			return err
		}

		v.Set(setVal)
	}

	return nil
}

func (p *Parser) getValue(str string, typ reflect.Type) (reflect.Value, error) {
	kind := typ.Kind()

	if fn, ok := p.typeMap[typ.String()]; ok {
		return fn(str)

	} else if kind == reflect.String {
		return reflect.ValueOf(str), nil

	} else if kind == reflect.Bool {
		bool, err := strconv.ParseBool(str)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("must be true or false")
		}
		return reflect.ValueOf(bool), nil

	} else if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 {
		i, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(i).Convert(typ), nil

	} else if kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
		i, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(i).Convert(typ), nil

	}

	return reflect.Value{}, fmt.Errorf("unknown type %v", typ)
}

func (p *Parser) makeFieldMap(ptr any) (map[string]reflect.Value, error) {
	fieldMap := map[string]reflect.Value{}

	v := reflect.ValueOf(ptr)
	for v.Type().Kind() == reflect.Pointer {
		v = v.Elem()
	}

	typ := v.Type()
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v instead", v.Type().Kind())
	}

	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldVal := v.Field(i)
		if !fieldVal.CanSet() {
			return nil, fmt.Errorf("can't set struct values -- did you remember to pass a pointer?")
		}

		var name string
		if tagName := field.Tag.Get("ini"); tagName != "" {
			name, _, _ = strings.Cut(strings.ToLower(tagName), ",")
		} else {
			name = strings.ToLower(field.Name)
		}

		fieldMap[name] = fieldVal
	}

	return fieldMap, nil
}
