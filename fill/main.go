package fill

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

var (
	timeType   = reflect.TypeOf(time.Time{})
	fillerType = reflect.TypeOf((*RandFiller)(nil)).Elem()
)

// RandFiller controls how it is filled.
// Implementing this type allows you
// to control the method in which your type
// is populated with random data
type RandFiller interface {
	RandFill(*rand.Rand)
}

// Fill fills all exported fields of a struct with random data
func Fill(val reflect.Value, rand *rand.Rand) {
	if !val.CanSet() {
		return
	}

	switch {
	case val.Type().Implements(fillerType):
		t := val.Type()
		numIndirects := 0
		for ; t.Kind() == reflect.Ptr; numIndirects++ {
			t = t.Elem()
		}
		v := reflect.New(t).Elem()
		for ; numIndirects > 0; numIndirects-- {
			v = v.Addr()
		}
		v.Interface().(RandFiller).RandFill(rand)
		val.Set(v)
	case val.Type() == timeType:
		val.Set(reflect.ValueOf(time.Date(
			rand.Intn(50)+1990,
			time.Month(rand.Intn(12)),
			rand.Intn(28),
			rand.Intn(24),
			rand.Intn(60),
			rand.Intn(60),
			rand.Intn(1e9),
			time.UTC),
		))
	default:
		switch val.Type().Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val.SetUint(uint64(rand.Int63()))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val.SetInt(rand.Int63())
		case reflect.String:
			val.SetString("string")
		case reflect.Bool:
			val.SetBool(rand.Int()%2 == 0)
		case reflect.Float32, reflect.Float64:
			val.SetFloat(rand.Float64())
		case reflect.Slice:
			length := rand.Intn(5) + 3
			slice := reflect.MakeSlice(val.Type(), length, length)
			for i := 0; i < length; i++ {
				v := reflect.New(val.Type().Elem())
				Fill(v.Elem(), rand)
				slice.Index(i).Set(v.Elem())
			}
			val.Set(slice)
		case reflect.Ptr:
			elem := reflect.New(val.Type().Elem())
			Fill(elem.Elem(), rand)
			val.Set(elem)
		case reflect.Struct:
			numFields := val.NumField()
			for i := 0; i < numFields; i++ {
				Fill(val.Field(i), rand)
			}
		default:
			fmt.Println("Unsupported field: ", val.Type())
		}
	}

}
