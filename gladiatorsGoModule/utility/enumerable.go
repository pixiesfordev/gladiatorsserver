package utility

// 移除重複的元素
func RemoveDuplicatesFromSlice[T comparable](slice []T) []T {
	unique := make(map[T]bool)
	var result []T

	for _, v := range slice {
		if _, ok := unique[v]; !ok {
			unique[v] = true
			result = append(result, v)
		}
	}
	return result
}

// 計算數字切片的總和
func SliceSum[T Number](slice []T) T {
	var sum T
	for _, v := range slice {
		sum += v
	}
	return sum
}

// 從map中移除傳入的key陣列
func RemoveFromMapByKeys[K comparable, V any](myMap map[K]*V, keys []K) {
	for _, key := range keys {
		delete(myMap, key)
	}
}

// 從map中移除傳入的key陣列
func RemoveFromMapByKeys2[K comparable, V any](myMap map[K]V, keys []K) {
	for _, key := range keys {
		delete(myMap, key)
	}
}

// 從一個 slice 中移除指定索引的元素
func RemoveFromSliceByIdx[T any](slice []T, idx int) []T {
	return append(slice[:idx], slice[idx+1:]...)
}

// 從一個 slice 中移除多個索引, 多個索引是另一個slice的元素
func RemoveFromSliceByIdxs[T any](slice []T, idxs []int) []T {
	removeSet := make(map[int]bool)
	for _, idx := range idxs {
		removeSet[idx] = true
	}

	var newSlice []T
	for i, v := range slice {
		if !removeSet[i] {
			newSlice = append(newSlice, v)
		}
	}

	return newSlice
}

// Contains 檢查特定元素是否存在於切片中
func Contains[T comparable](slice []T, element T) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

// 某數字是否在數字切片中, 是就返回true, 切片回空返回false
func NumberInSlice[T Number](slice []T, target T) bool {
	if len(slice) == 0 {
		return false
	}
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}

// 切片元素是否都等於某數字, 切片回空返回false
func SliceNumberAllEqualTo[T Number](slice []T, target T) bool {
	if len(slice) == 0 {
		return false
	}
	for _, v := range slice {
		if v != target {
			return false
		}
	}
	return true
}