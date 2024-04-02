package utility

import (
	"fmt"
	"math"
	"reflect"
)

type Number interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64
}

// 將傳入浮點四捨五入到指定位數
func RoundToDecimal(value float64, decimalPlaces int) float64 {
	shift := math.Pow(10, float64(decimalPlaces))
	return math.Round(value*shift) / shift
}

// 將傳入小數slice四捨五入
func RoundSlice[T float64 | float32](s []T, decimalPlaces int) []T {
	rounded := make([]T, len(s))
	for i, v := range s {
		rounded[i] = T(RoundToDecimal(float64(v), decimalPlaces))
	}
	return rounded
}

// 通過反射對struct內所有float64和float32欄位進行四捨五入
func RoundDecimalInStruct(target interface{}, places int) (interface{}, error) {
	val := reflect.ValueOf(target)
	var structVal reflect.Value

	if val.Kind() == reflect.Ptr {
		structVal = val.Elem()
	} else if val.Kind() == reflect.Struct {
		structVal = val
	} else {
		return nil, fmt.Errorf("傳入的必須是struct或一個指標")
	}

	// 建立副本
	structCopy := reflect.New(structVal.Type()).Elem()
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		copiedField := structCopy.Field(i)

		switch field.Kind() {
		case reflect.Float32, reflect.Float64: // 處理小數
			rounded := RoundToDecimal(field.Float(), places)
			copiedField.SetFloat(rounded)
		case reflect.Slice: // 處理slice
			newSlice := reflect.MakeSlice(field.Type(), field.Len(), field.Cap())
			for j := 0; j < field.Len(); j++ {
				item := field.Index(j)
				if item.Kind() == reflect.Float32 || item.Kind() == reflect.Float64 {
					rounded := RoundToDecimal(item.Float(), places)
					newSlice.Index(j).SetFloat(rounded)
				} else {
					newSlice.Index(j).Set(item)
				}
			}
			copiedField.Set(newSlice)
		case reflect.Array: //處理array
			// 需要建立一個暫時用的slice並修改這個slice值, 最後再把值設回去原來的array
			arrayLen := field.Len()
			tmpSliceType := reflect.SliceOf(field.Type().Elem())
			tmpSlice := reflect.MakeSlice(tmpSliceType, arrayLen, arrayLen)

			// 設定暫時用的slice值
			for j := 0; j < arrayLen; j++ {
				item := field.Index(j)
				if item.Kind() == reflect.Float32 || item.Kind() == reflect.Float64 {
					rounded := RoundToDecimal(item.Float(), places)
					tmpSlice.Index(j).SetFloat(rounded)
				} else {
					tmpSlice.Index(j).Set(item)
				}
			}

			// 把值設回去原來的array
			for j := 0; j < arrayLen; j++ {
				copiedField.Index(j).Set(tmpSlice.Index(j))
			}
		default:
			copiedField.Set(field)
		}
	}
	// 根據類型返回
	if val.Kind() == reflect.Ptr {
		return structCopy.Addr().Interface(), nil
	}
	return structCopy.Interface(), nil
}
