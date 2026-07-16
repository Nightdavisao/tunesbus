package olejunk

import (
	"fmt"
	"reflect"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func GetPropertyFromIDispatch[T any](disp *ole.IDispatch, name string, params ...any) (*T, error) {
	variant, err := oleutil.GetProperty(disp, name, params...)
	if err != nil {
		return nil, err
	}
	property, err := GetVariantValue[T](variant)
	if err != nil {
		return nil, err
	}
	return property, nil
}

// only use this if no property in the list requires a parameter at all
func GetPropertiesFromIDispatch[T any](disp *ole.IDispatch, properties []string) (map[string]*T, error) {
	result := make(map[string]*T, len(properties))
	for _, name := range properties {
		value, err := GetPropertyFromIDispatch[T](disp, name)
		if err != nil {
			return nil, fmt.Errorf("property %s: %w", name, err)
		}
		result[name] = value
	}
	return result, nil
}

func GetVariantValue[T any](variant *ole.VARIANT) (name *T, err error) {
	if variant == nil {
		return nil, fmt.Errorf("property returned no value")
	}
	defer variant.Clear()

	value := reflect.ValueOf(variant.Value())
	targetType := reflect.TypeOf((*T)(nil)).Elem()
	if !value.IsValid() {
		return nil, fmt.Errorf("property %v returned no value", variant)
	}
	if !value.Type().AssignableTo(targetType) {
		if !value.Type().ConvertibleTo(targetType) {
			return nil, fmt.Errorf("property %v did not return %v", variant, targetType)
		}
		value = value.Convert(targetType)
	}

	v, ok := value.Interface().(T)
	if !ok {
		return nil, fmt.Errorf("property %v did not return %T", variant, name)
	}
	return &v, nil
}
