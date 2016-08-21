package structinitializer

import (
	"testing"
	"fmt"
	"reflect"
)

type Validator interface {
	Valid() bool
}

type TNormStruct struct  {
	AInt int
	AStr string
}

type TStruct struct {
	AInt int `default:"10"`
	AStr string `default:"hello"`
}

func (self *TStruct) Valid() bool {
	return self.AStr == "hello" && self.AInt == 10
}

type TWrongStruct struct {
	AInt int `default:"hhhhh"`
	AStr string `default:"hello"`
}

func (self *TWrongStruct) Valid() bool {
	return false
}

type TEmbeddStruct struct  {
	AStruct TStruct
	AInt int `default:"101"`
}

func (self *TEmbeddStruct) Valid() bool {
	return self.AStruct.Valid() && self.AInt == 101
}

type TEmbeddStruct2 struct  {
	AStruct TStruct
	AInt int
}

type TAnonymousStruct struct  {
	TStruct
	AUint uint16 `default:"20"`
}

func (self *TAnonymousStruct) Valid() bool  {
	return self.TStruct.Valid() && self.AUint == 20
}

type TPointerStruct struct {
	AStructPtr *TStruct
	AIntPtr *int32 `default:"20"`
	AStrPtr *string `default:"-"`
	AUintPtr *uint
	ANormStruct *TNormStruct
}

func (self *TPointerStruct) Valid() bool {
	if self.AStructPtr != nil && self.AStructPtr.Valid() && self.AIntPtr != nil && *self.AIntPtr == 20 && self.AStrPtr == nil && self.AUintPtr == nil && self.ANormStruct == nil {
		return true
	}
	return false
}

func AssertTrue(t *testing.T, a bool, msg string) {
	if !a {
		t.Errorf(msg)
	}
}

func TestCheckStructHasDefaultValue(t *testing.T) {
	self := NewStructInitialiser()
	AssertTrue(t, self.checkStructHasDefaultValue(reflect.TypeOf(TStruct{})), "TStruct check failed!")
	AssertTrue(t, !self.checkStructHasDefaultValue(reflect.TypeOf(TNormStruct{})), "TNormStrcut check succeed!")
	AssertTrue(t, self.checkStructHasDefaultValue(reflect.TypeOf(TEmbeddStruct2{})), "TEmbeddStruct2 check failed!")
}

func checkError(t *testing.T, err error, wantReason string) {
	if err == nil {
		t.Errorf("except non pointer err!")
		return
	} else {
		if err1 := err.(*Error); err1.Reason != wantReason {
			t.Errorf("want '%s' but '%s': %s", wantReason, err1.Reason, err1.Message)
			return
		}
	}
	fmt.Printf("%#v\n", err)
}

func TestNonPointer(t *testing.T) {
	aStruct := TStruct{}
	err := InitializeStruct(aStruct)
	checkError(t, err, ERROR_NOT_A_POINTER)
}

func TestPointToNonStruct(t *testing.T) {
	aInt := 100
	err := InitializeStruct(&aInt)
	checkError(t, err, ERROR_POINTE_TO_NON_STRUCT)
}

func TestWrongTag(t *testing.T) {
	aStruct := TWrongStruct{}
	err := InitializeStruct(&aStruct)
	checkError(t, err, ERROR_TYPE_CONVERSION_FAILED)
}

func TestDefaultSetted(t *testing.T) {
	aStruct := TStruct{
		AInt: 103,
	}
	err := InitializeStruct(&aStruct)
	if err != nil || aStruct.AInt != 103 {
		t.Errorf("BOOM!!! rewrite a user setted value? err: %s", err)
	}
	fmt.Printf("%#v\n", aStruct)
}

func checkStruct(t *testing.T, v Validator) {
	err := InitializeStruct(v)
	if err != nil {
		t.Errorf("unexcept err: %#v, aStruct: %#v\n", err, v)
		return
	}

	if !v.Valid() {
		t.Errorf("init incorret: %#v\n", v)
		return
	}
	fmt.Printf("%#v\n", v)
}

func TestEmbeddStruct(t *testing.T) {
	aStruct := TEmbeddStruct{}
	checkStruct(t, &aStruct)
}

func TestAnonymousStrct(t *testing.T) {
	aStruct := TAnonymousStruct{}
	checkStruct(t, &aStruct)
}

func TestNormalStruct(t *testing.T) {
	aStruct := TStruct{}
	checkStruct(t, &aStruct)
}

func TestPointerStruct(t *testing.T) {
	aStruct := TPointerStruct{}
	checkStruct(t, &aStruct)
}
