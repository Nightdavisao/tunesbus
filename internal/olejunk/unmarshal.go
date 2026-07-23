package olejunk

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/charmbracelet/log"
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

		tag := field.Tag.Get("com")
		if tag == "-" {
			continue
		}

		parts := strings.Split(tag, ",")
		propName := parts[0]
		if propName == "" {
			propName = field.Name
		}

		allowEmpty := false
		for _, opt := range parts[1:] {
			if opt == "allowempty" {
				allowEmpty = true
			}
		}

		variant, err := oleutil.GetProperty(disp, propName)
		if err != nil {
			if allowEmpty {
				// the code is the DISP_E_UNKNOWNNAME constant apparently...?
				if oleErr, ok := err.(*ole.OleError); ok && oleErr.Code() == 2147614726 {
					continue
				}
			}
			log.Debugf("error getting prop %s for field %q: %v", propName, field.Name, err)
			return fmt.Errorf("get property %q for field %q: %w", propName, field.Name, err)
		}

		if err := assign(fv, variant); err != nil {
			variant.Clear()
			return fmt.Errorf("assign property %q to field %q: %w", propName, field.Name, err)
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

func GetCOMObjectFromVariant[T any](object *ole.VARIANT, iid string, releaser *OleReleaser) (*T, *ole.IDispatch, error) {
	if object != nil {
		disp := object.ToIDispatch()
		obj, err := GetCOMObject[T](disp, iid, releaser)
		if err != nil {
			return nil, nil, err
		}
		return obj, disp, nil
	}
	return nil, nil, errors.New("object is nil")
}

func GetCOMObject[T any](iDispatch *ole.IDispatch, iid string, releaser *OleReleaser) (*T, error) {
	if iDispatch == nil {
		return nil, errors.New("iDispatch is nil")
	}
	if releaser != nil {
		releaser.Add(&iDispatch.IUnknown)
	}
	
	guid := ole.NewGUID(iid)
	if guid == nil {
		return nil, errors.New("guid is nil")
	}
	
	disp, err := iDispatch.QueryInterface(guid)
	if err != nil {
		return nil, fmt.Errorf("query interface %s: %w", iid, err)
	}
	if releaser != nil {
		releaser.Add(&disp.IUnknown)
	}

	var result T
	if err := UnmarshalCOM(disp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
