package utility

import "math"

// 將角度轉換為弧度
func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}
