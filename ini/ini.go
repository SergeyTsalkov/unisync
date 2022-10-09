package ini

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Unmarshaler interface {
	UnmarshalINI([]byte) error
}

var unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

func Unmarshal(data []byte, ptr any) error {
	fieldMap, err := makeFieldMap(ptr)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		key, value, valid := strings.Cut(line, "=")
		if !valid {
			continue
		}

		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)

		v, ok := fieldMap[key]
		if !ok {
			return fmt.Errorf("invalid field %v", key)
		}

		err = setValue(v, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func setValue(v reflect.Value, str string) error {
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

	setVal, err := getValue(str, v.Type(), true)
	if err != nil {
		return err
	}
	if v.Type().Kind() == reflect.Slice {
		v.Set(reflect.Append(v, setVal))
	} else {
		v.Set(setVal)
	}

	return nil
}

func getValue(str string, typ reflect.Type, canRecurse bool) (reflect.Value, error) {
	kind := typ.Kind()

	if kind == reflect.Slice && canRecurse {
		return getValue(str, typ.Elem(), false)

	} else if kind == reflect.String {
		return reflect.ValueOf(str), nil

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

	return reflect.Value{}, fmt.Errorf("struct contains unknown type %v", typ)
}

func makeFieldMap(ptr any) (map[string]reflect.Value, error) {
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
		if tagName := field.Tag.Get("json"); tagName != "" {
			name, _, _ = strings.Cut(strings.ToLower(tagName), ",")
		} else {
			name = strings.ToLower(field.Name)
		}

		fieldMap[name] = fieldVal
	}

	return fieldMap, nil
}
