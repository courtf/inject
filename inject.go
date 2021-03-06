package inject

import (
	"errors"
	"fmt"
	"reflect"
)

type Injector interface {
	Applicator
	Invoker
	TypeMapper
	SetParent(Injector)
}

type Applicator interface {
	Apply(interface{}) error
}

type Invoker interface {
	Invoke(interface{}) ([]reflect.Value, error)
}

type TypeMapper interface {
	Map(interface{})
	MapTo(interface{}, interface{})
	Get(reflect.Type) reflect.Value
}

type injector struct {
	values map[reflect.Type]reflect.Value
	parent Injector
}

func InterfaceOf(value interface{}) reflect.Type {
	t := reflect.TypeOf(value)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()

		if t.Kind() == reflect.Interface {
			return t
		}
	}

	panic("Called inject.InterfaceOf with a value that is not a pointer to an interface. (*MyInterface)(nil)")
	return nil
}

func New() Injector {
	return &injector{
		values: make(map[reflect.Type]reflect.Value),
	}
}

func (inj *injector) Invoke(f interface{}) ([]reflect.Value, error) {
	t := reflect.TypeOf(f)

	var in = make([]reflect.Value, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		val := inj.Get(argType)
		if !val.IsValid() {
			return nil, errors.New(fmt.Sprintf("Value not found for type %v", argType))
		}

		in[i] = val
	}

	return reflect.ValueOf(f).Call(in), nil
}

func (inj *injector) Apply(val interface{}) error {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		structField := t.Field(i)
		if f.CanSet() && structField.Tag == "inject" {
			ft := f.Type()
			v := inj.Get(ft)
			if !v.IsValid() {
				return errors.New(fmt.Sprintf("Value not found for type %v", ft))
			}

			f.Set(v)
		}

	}

	return nil
}

func (i *injector) Map(val interface{}) {
	i.values[reflect.TypeOf(val)] = reflect.ValueOf(val)
}

func (i *injector) MapTo(val interface{}, ifacePtr interface{}) {
	i.values[InterfaceOf(ifacePtr)] = reflect.ValueOf(val)
}

func (i *injector) Get(t reflect.Type) reflect.Value {
	val := i.values[t]
	if !val.IsValid() && i.parent != nil {
		val = i.parent.Get(t)
	}
	return val
}

func (i *injector) SetParent(parent Injector) {
	i.parent = parent
}
