package olejunk

import (
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"fmt"
)

func GetPropertyFromIDispatch[T any](disp *ole.IDispatch, name string, params ...any) (*T, error) {
	disp.AddRef()
	variant, err := oleutil.GetProperty(disp, name, params...)
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
		defer disp.Release()
		if err != nil {
			return nil, fmt.Errorf("property %s: %w", name, err)
		}
		result[name] = value
	}
	return result, nil
}

func GetVariantValue[T any](variant *ole.VARIANT) (name *T, err error) {
	defer variant.Clear()
	v, ok := variant.Value().(T)
	if !ok {
		return nil, fmt.Errorf("property %v did not return %T", variant, name)
	}
	return &v, nil
}
