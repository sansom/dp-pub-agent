package testutil

import (
	"reflect"
	"testing"
)

func CompareVar(t *testing.T, sut interface{}, expect interface{}) {
	switch reflect.TypeOf(expect).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(sut)
		sE := reflect.ValueOf(expect)
		if sE.Len() != s.Len() {
			t.Errorf("length %d != expected %d", s.Len(), sE.Len())
		}
		for i := 0; i < sE.Len(); i++ {
			CompareVar(t, s.Index(i).Interface(), sE.Index(i).Interface())
		}
	case reflect.Ptr:
		val := reflect.ValueOf(expect).Elem()
		valE := reflect.ValueOf(sut).Elem()
		for i := 0; i < valE.NumField(); i++ {
			valueField := val.Field(i)
			valueFieldE := valE.Field(i)
			CompareVar(t, valueField.Interface(), valueFieldE.Interface())
		}
	default:
		if sut != expect {
			t.Errorf("%s != expected %s", sut, expect)
		}
	}
}
