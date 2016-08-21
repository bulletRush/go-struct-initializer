package structinitializer

import (
	"reflect"
	"fmt"
	"strconv"
	"strings"
)

const (
	ERROR_NOT_A_POINTER = "NotAPointer"
	ERROR_NOT_IMPLEMENT = "NotImplement"
	ERROR_POINTE_TO_NON_STRUCT = "PointToNonStruct"
	ERROR_TYPE_CONVERSION_FAILED = "TypeConversionFailed"
	ERROR_MULTI_ERRORS_FOUND = "MultiErrorsFound"
)

type Error struct {
	Reason  string
	Message string
	Stack   string
}

func NewError(reason, message string, stack string) error {
	return &Error{reason, message, stack}
}

func mergedError(errors []error, stackName string) error {
	if len(errors) == 0 {
		return nil
	}
	if len(errors) == 1 {
		return errors[0]
	}
	points := make([]string, len(errors))
	for i, err := range errors {
		points[i] = fmt.Sprintf("* %s", err)
	}
	return &Error{
		ERROR_MULTI_ERRORS_FOUND,
		strings.Join(points, "\n"),
		stackName,
	}
}

func (e *Error) Error() string {
	if e.Stack == "" {
		return fmt.Sprintf("%s: %s", e.Reason, e.Message)
	}
	return fmt.Sprintf("%s:(%s) %s", e.Reason, e.Stack, e.Message)
}

type InitialiserConfig struct {
	TagName string
}

type Initialiser struct {
	config *InitialiserConfig
}

func NewStructInitialiser() Initialiser {
	config := InitialiserConfig{
		TagName: "default",
	}
	initialiser := Initialiser{
		config: &config,
	}
	return initialiser
}

func (self *Initialiser) initialiseInt(stackName, tag string, val reflect.Value) error {
	if tag != "" {
		i, err := strconv.ParseInt(tag, 10, 64)
		if err != nil {
			return NewError(
				ERROR_TYPE_CONVERSION_FAILED,
				fmt.Sprintf(`"%s" can't convert to int64`, tag),
				stackName,
			)
		}
		if val.Int() == 0 {
			val.SetInt(int64(i))
		}
	}
	return nil
}

func (self *Initialiser) initialiseUint(stackName, tag string, val reflect.Value) error {
	if tag != "" {
		i, err := strconv.ParseUint(tag, 10, 64)
		if err != nil {
			return NewError(
				ERROR_TYPE_CONVERSION_FAILED,
				fmt.Sprintf(`"%s" can't convert to uint64`, tag),
				stackName,
			)
		}
		if val.Uint() == 0 {
			val.SetUint(uint64(i))
		}
	}
	return nil
}

func (self *Initialiser) initialiseString(stackName, tag string, val reflect.Value) error {
	if tag != "" && val.String() == ""{
		val.SetString(tag)
	}
	return nil
}

func (self *Initialiser) checkStructHasDefaultValue(typ reflect.Type) bool {
	for i:=0; i<typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Tag.Get(self.config.TagName) != "" {
			return true
		}
		if field.Type.Kind() == reflect.Struct && self.checkStructHasDefaultValue(field.Type) {
			return true
		}
	}
	return false
}

func (self *Initialiser) initialisePtr(stackName, tag string, val reflect.Value) error {
	if tag == "-" {
		return nil
	}
	typ := val.Type().Elem()
	if tag == "" {
		if typ.Kind() != reflect.Struct {
			return nil
		} else if !self.checkStructHasDefaultValue(typ) {
			return nil
		}
	}
	realVal := reflect.New(typ)
	if err := self.initialise(stackName, tag, reflect.Indirect(realVal)); err != nil {
		return err
	}
	val.Set(realVal)
	return nil
}

func (self *Initialiser) initialise(stackName string, tag string, val reflect.Value) error {
	typ := val.Type()
	kind := typ.Kind()
	if kind == reflect.Struct {
		return self.initialiseStruct(stackName, val)
	}
	switch kind {
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		return self.initialiseInt(stackName, tag, val)
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		return self.initialiseUint(stackName, tag, val)
	case reflect.String:
		return self.initialiseString(stackName, tag, val)
	case reflect.Ptr:
		return self.initialisePtr(stackName, tag, val)
	default:
		return NewError(ERROR_NOT_IMPLEMENT, "not implement", stackName)
	}
	return NewError(ERROR_NOT_IMPLEMENT, "not implement", stackName)
}

func (self *Initialiser) initialiseStruct(stackName string, structVal reflect.Value) error {
	structs := make([]reflect.Value, 1, 5)
	structs[0] = structVal
	errors := make([]error, 0)
	for len(structs) > 0 {
		structVal := structs[0]
		structs = structs[1:]
		structTyp := structVal.Type()
		for i:=0; i<structTyp.NumField(); i++ {
			structField := structTyp.Field(i)
			tag := structField.Tag
			defaultTag := tag.Get(self.config.TagName);
			if structField.Anonymous {
				structs = append(structs, structVal.Field(i))
				continue
			}
			fieldName := structField.Name
			if stackName != "" {
				fieldName = fmt.Sprintf("%s.%s", stackName, fieldName)
			}
			err := self.initialise(
				fieldName,
				defaultTag,
				structVal.Field(i),
			)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}
	return mergedError(errors, stackName)
}

func (self *Initialiser) Initialise(inf interface{}) error {
	typ := reflect.TypeOf(inf)
	kind := typ.Kind()
	if kind != reflect.Ptr {
		return NewError(
			ERROR_NOT_A_POINTER,
			fmt.Sprintf("%s not a pointer", typ),
			"",
		)
	}
	val := reflect.ValueOf(inf).Elem()
	kind = val.Kind()
	if kind != reflect.Struct {
		return NewError(
			ERROR_POINTE_TO_NON_STRUCT,
			fmt.Sprintf("%s point to non struct", typ),
			"",
		)
	}
	return self.initialiseStruct(val.Type().Name(), val)
}

func InitializeStruct(struct_ interface{}) error {
	initialiser := NewStructInitialiser()
	return initialiser.Initialise(struct_)
}
