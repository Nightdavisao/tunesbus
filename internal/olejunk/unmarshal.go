package olejunk

import (
	"errors"
	"fmt"
	"reflect"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

func UnmarshalCOM(disp *ole.IDispatch, dst any) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("UnmarshalCOM: dst must be a pointer to a struct")
	}
	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fv := rv.Field(i)
		tag, ok := field.Tag.Lookup("com")

		if ok && tag == "self" {
			if fv.Type() != reflect.TypeOf(disp) {
				return fmt.Errorf("field %q tagged com:\"self\" must be *ole.IDispatch", field.Name)
			}
			disp.AddRef()
			fv.Set(reflect.ValueOf(disp))
			continue
		}
		if ok && tag == "-" {
			continue
		}
		
		propName := field.Name
		if ok && tag != "" {
			propName = tag
		}

		variant, err := oleutil.GetProperty(disp, propName)
		if err != nil {
			return fmt.Errorf("get property %q for field %q: %w", tag, field.Name, err)
		}

		if err := assign(fv, variant); err != nil {
			variant.Clear()
			return fmt.Errorf("assign property %q to field %q: %w", tag, field.Name, err)
		}
		variant.Clear()
	}
	return nil
}

func assign(fv reflect.Value, variant *ole.VARIANT) error {
	val := reflect.ValueOf(variant.Value())
	if !val.IsValid() {
		return nil
	}

	if val.Type().AssignableTo(fv.Type()) {
		fv.Set(val)
		return nil
	}

	if val.Type().ConvertibleTo(fv.Type()) {
		fv.Set(val.Convert(fv.Type()))
		return nil
	}

	return fmt.Errorf("cannot assign COM value of type %s to field of type %s", val.Type(), fv.Type())
}

func GetCOMObjectFromVariant[T any](object *ole.VARIANT, iid string) (*T, error) {
	if object != nil {
		return GetCOMObject[T](object.ToIDispatch(), iid)
	}
	return nil, errors.New("object is nil")
}

func GetCOMObject[T any](iDispatch *ole.IDispatch, iid string) (*T, error) {
	if iDispatch == nil {
		return nil, errors.New("iDispatch is nil")
	}
	
	guid := ole.NewGUID(iid)
	if guid == nil {
		return nil, errors.New("guid is nil")
	}
	
	disp, err := iDispatch.QueryInterface(guid); if err != nil {
		return nil, fmt.Errorf("query interface %s: %w", iid, err)
	}
	disp.AddRef()
	defer disp.Release()

	var result T
	if err := UnmarshalCOM(disp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}