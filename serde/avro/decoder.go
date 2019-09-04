package avro

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"

	"github.com/pkg/errors"
)

var (
	typeRegistry   = make(map[string]reflect.Type)
	typeRegistryMu sync.RWMutex

	optionalTypes = map[reflect.Kind]string{
		reflect.String:  "string",
		reflect.Int32:   "int",
		reflect.Int64:   "long",
		reflect.Bool:    "boolean",
		reflect.Float32: "float",
		reflect.Float64: "double",
	}
)

func Register(val interface{}) {
	RegisterName("", val)
}

func RegisterName(name string, val interface{}) {

	rt := reflect.TypeOf(val)
	rv := reflect.ValueOf(val)

	if rv.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if name == "" {
		name = rt.Name()
	}

	typeRegistryMu.Lock()
	defer typeRegistryMu.Unlock()

	if _, ok := typeRegistry[name]; ok {
		panic("object with name already registered: " + name)
	}

	typeRegistry[name] = rt
}

func getRegisteredType(name string) (reflect.Type, bool) {

	typeRegistryMu.RLock()
	rt, ok := typeRegistry[name]
	typeRegistryMu.RUnlock()

	if !ok {
		return nil, false
	}

	return rt, true
}

// TODO: decode from io.Reader
func DecodeToStruct(dest interface{}, src map[string]interface{}) error {
	return decodeToStruct(dest, src)
}

func decodeToStruct(dest interface{}, src map[string]interface{}) error {

	rt := reflect.TypeOf(dest)
	rv := reflect.ValueOf(dest)

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
		rt = rt.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return errors.New("the target is not a struct")
	}

	for i := 0; i < rv.NumField(); i++ {
		fv := rv.Field(i)
		if !fv.CanSet() {
			continue
		}

		ft := rt.Field(i)
		tag, ok := ft.Tag.Lookup("avro")
		if !ok || len(tag) == 0 {
			name := []byte(ft.Name)
			copy(name[:1], bytes.ToLower(name[:1]))
			tag = string(name)
		}

		srcVal, ok := src[tag]
		if !ok {
			continue
		} else if srcVal == nil {
			continue
		}

		if err := decodeValue(&fv, srcVal); err != nil {
			return errors.Wrap(err, "field: "+tag)
		}
	}

	return nil
}

func decodeValue(dest *reflect.Value, srcVal interface{}) error {

	srcMap, ok := srcVal.(map[string]interface{})
	if !ok {
		srcMap = nil
	} else if _, ok = srcMap["null"]; ok {
		return nil
	}

	destKind := dest.Kind()
	switch destKind {
	case reflect.Ptr:
		val := reflect.New(dest.Type().Elem())
		dest.Set(val)

		target := dest.Elem()
		if err := decodeValue(&target, srcVal); err != nil {
			return err
		}

	case reflect.Map:

		targetMap := reflect.MakeMap(dest.Type())
		dest.Set(targetMap)

		if len(srcMap) > 0 {
			for k, v := range srcMap {
				targetValPtr := reflect.New(dest.Type().Elem())
				targetVal := targetValPtr.Elem()
				if err := decodeValue(&targetVal, v); err != nil {
					return err
				}

				targetMap.SetMapIndex(reflect.ValueOf(k), targetVal)
			}
		}

	case reflect.Interface, reflect.Struct:

		var tp reflect.Type
		if len(srcMap) == 1 {
			// if a field is optional
			for objectName, objectFields := range srcMap {
				if tp, ok = getRegisteredType(objectName); ok {
					objectFieldsMap, okCast := objectFields.(map[string]interface{})
					if okCast {
						srcMap = objectFieldsMap
					}
				}
			}
		}

		if tp == nil {
			objectName := dest.Type().Name()
			tp, ok = getRegisteredType(objectName)
			if !ok {
				return fmt.Errorf("failed to set %v value: is not registered type with name '%s': %v", destKind, objectName, srcMap)
			}
		}

		target := reflect.New(tp).Interface()
		if err := decodeToStruct(target, srcMap); err != nil {
			return err
		}

		switch destKind {
		case reflect.Interface:
			dest.Set(reflect.ValueOf(target))
		case reflect.Struct:
			dest.Set(reflect.ValueOf(target).Elem())
		}

	case reflect.String, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64, reflect.Bool:
		if srcMap != nil {
			srcVal = srcMap[optionalTypes[destKind]]
		}

		itemVal := reflect.ValueOf(srcVal)
		if itemVal.Kind() != destKind {
			return fmt.Errorf("failed to set %v value: %#v", destKind, itemVal)
		}
		dest.Set(itemVal)

	case reflect.Slice:
		sliceElemKind := dest.Type().Elem().Kind()

		switch sliceElemKind {
		case reflect.Uint8:
			if srcMap != nil {
				srcVal = srcMap["bytes"]
			}

			data, ok := srcVal.([]byte)
			if !ok {
				return fmt.Errorf("failed to set %v value: %#v", destKind, srcVal)
			}
			dest.Set(reflect.ValueOf(data))

		case reflect.String, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64, reflect.Bool:
			if srcMap != nil {
				srcVal = srcMap["array"]
			}

			data, ok := srcVal.([]interface{})
			if !ok {
				return fmt.Errorf("failed to set %v value: %#v", destKind, srcVal)
			}

			targetSlice := reflect.MakeSlice(dest.Type(), len(data), len(data))
			dest.Set(targetSlice)

			dataV := reflect.ValueOf(data)
			for arrI := 0; arrI < dataV.Len(); arrI++ {
				item := dataV.Index(arrI).Interface()
				itemVal := reflect.ValueOf(item)
				if itemVal.Kind() != sliceElemKind {
					return fmt.Errorf("failed to set %v value: %#v(item: %d)", destKind, itemVal, arrI)
				}

				targetSlice.Index(arrI).Set(itemVal)
			}

		case reflect.Interface, reflect.Struct, reflect.Ptr:
			if srcMap != nil {
				srcVal = srcMap["array"]
			}

			data, ok := srcVal.([]interface{})
			if !ok {
				return fmt.Errorf("failed to set %v value: %#v", destKind, srcVal)
			}

			targetSlice := reflect.MakeSlice(dest.Type(), len(data), len(data))
			dest.Set(targetSlice)

			dataV := reflect.ValueOf(data)
			for arrI := 0; arrI < dataV.Len(); arrI++ {
				item := dataV.Index(arrI).Interface()

				targetValPtr := reflect.New(dest.Type().Elem())
				targetVal := targetValPtr.Elem()
				if err := decodeValue(&targetVal, item); err != nil {
					return err
				}

				targetSlice.Index(arrI).Set(targetVal)
			}

		default:
			return fmt.Errorf("failed to set %v[%v] value: unsupported type", destKind, sliceElemKind)
		}

	default:
		return fmt.Errorf("failed to set %v value: unsupported type", destKind)
	}

	return nil
}
