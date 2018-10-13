package expr

import (
	"fmt"
	"reflect"
)

func toBool(n Node, v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	}
	panic(fmt.Sprintf("cannot convert %v (type %T) to type bool", n, v))
}

func toText(n Node, v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	}
	panic(fmt.Sprintf("cannot convert %v (type %T) to type string", n, v))
}

func toNumber(n Node, v reflect.Value) float64 {
	f, ok := cast(v)
	if ok {
		return f
	}
	panic(fmt.Sprintf("cannot convert %v (type %T) to type float64", n, v))
}

func cast(v reflect.Value) (float64, bool) {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), true

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), true // TODO: Check if uint64 fits into float64.
	}
	return 0, false
}

func isNumber(v reflect.Value) bool {
	return v.Type().Kind() == reflect.Float64
}

func canBeNumber(v reflect.Value) bool {
	return isNumberType(v.Type())
	return false
}

func equal(left, right reflect.Value) bool {
	if isNumber(left) && canBeNumber(right) {
		right, _ := cast(right)
		return left.Interface() == right
	} else if canBeNumber(left) && isNumber(right) {
		left, _ := cast(left)
		return left == right.Interface()
	} else {
		return reflect.DeepEqual(left.Interface(), right.Interface())
	}
}

func extract(v, i reflect.Value) (reflect.Value, bool) {
	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		n, ok := cast(i)
		if !ok {
			break
		}

		value := v.Index(int(n))
		return value, true

	case reflect.Map:
		value := v.MapIndex(i)
		return value, true
	case reflect.Struct:
		value := v.FieldByName(i.String())
		return value, true
	case reflect.Ptr:
		value := v.Elem()
		return extract(value, i)
	case reflect.Interface:
		value := v.Interface()
		return extract(reflect.ValueOf(value), i)
	}
	return null, false
}

func getFunc(v, i reflect.Value) (reflect.Value, bool) {
	d := v
	if v.Kind() == reflect.Ptr {
		d = v.Elem()
	}

	switch d.Kind() {
	case reflect.Map:
		value := d.MapIndex(i)
		if value.IsValid() && value.CanInterface() {
			return value, true
		}
	case reflect.Struct:
		name := i.String()
		method := v.MethodByName(name)
		return method, true

		// If struct has not method, maybe it has func field.
		// To access this field we need dereferenced value.
		value := d.FieldByName(name)
		return value, true
	}
	return null, false
}

func contains(needle, array reflect.Value) (bool, error) {
	if array.Interface() != nil {
		switch array.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < array.Len(); i++ {
				value := array.Index(i)

				if equal(value, needle) {
					return true, nil
				}

			}
			return false, nil
		case reflect.Map:
			n := reflect.ValueOf(needle)
			if !n.IsValid() {
				return false, fmt.Errorf("cannot use %T as index to %T", needle, array)
			}
			value := array.MapIndex(n)
			if value.IsValid() {
				return true, nil
			}
			return false, nil
		case reflect.Struct:
			n := reflect.ValueOf(needle)
			if !n.IsValid() || n.Kind() != reflect.String {
				return false, fmt.Errorf("cannot use %T as field name of %T", needle, array)
			}
			value := array.FieldByName(n.String())
			if value.IsValid() {
				return true, nil
			}
			return false, nil
		case reflect.Ptr:
			value := array.Elem()
			return contains(needle, value)

			return false, nil
		}
		return false, fmt.Errorf("operator \"in\" not defined on %T", array)
	}
	return false, nil
}

func isNil(v reflect.Value) bool {
	return !v.IsValid() || v.IsNil()
}
