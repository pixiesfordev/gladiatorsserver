package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	"math"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetEffectValue 取得第X個T類數值
func GetEffectValue[T any](e *Effect, idx int) (T, error) {
	var defaultT T
	values := strings.Split(e.ValueStr, ",")
	if len(values) <= idx {
		return defaultT, fmt.Errorf("取得Effect Value失敗 type: %v idx: %v", e.Type, idx)
	}

	valueStr := values[idx]
	var result T

	switch any(result).(type) {
	case int:
		intValue, err := strconv.Atoi(valueStr)
		if err != nil {
			return defaultT, fmt.Errorf("無法將值轉換為int: %v", err)
		}
		return any(intValue).(T), nil
	case float64:
		floatValue, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return defaultT, fmt.Errorf("無法將值轉換為float64: %v", err)
		}
		return any(floatValue).(T), nil
	case string:
		return any(valueStr).(T), nil
	default:
		return defaultT, fmt.Errorf("不支持的類型: %T", result)
	}
}

// GetStrUp 取得力量提升值
func (e *Effect) GetStrUpValue() int {
	addValue := 0
	var err error
	switch e.Type {
	case gameJson.StrUp:
		addValue, err = GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return 0
		}
	case gameJson.StrUpByHp:
		value, err := GetEffectValue[float64](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return 0
		}
		loseHpRatio := 1.0 - ((float64)(e.Target.CurHp) / (float64)(e.Target.Hp))
		addValue = int(math.Round(loseHpRatio * value))
	case gameJson.Berserk:
		addValue = int(math.Round(float64(e.Target.Str) * VALUE_BERSERK))
	}

	return addValue
}

// GetPDefUpValue 取得物理防禦提升值
func (e *Effect) GetPDefUpValue() int {
	addValue := 0
	var err error
	switch e.Type {
	case gameJson.PDefUp:
		addValue, err = GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return 0
		}
	}
	return addValue
}

// GetMDefUpValue 取得魔法防禦提升值
func (e *Effect) GetMDefUpValue() int {
	addValue := 0
	var err error
	switch e.Type {
	case gameJson.MDefUp:
		addValue, err = GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return 0
		}
	}
	return addValue
}

// GetPDefMultiple 取得物理防禦加乘
func (e *Effect) GetPDefMultiple() float64 {
	value := 0.0
	switch e.Type {
	case gameJson.Weak:
		value = VALUE_WEAK
	}
	return value
}

// GetMDefMultiple 取得魔法防禦加乘
func (e *Effect) GetMDefMultiple() float64 {
	value := 0.0
	switch e.Type {
	case gameJson.Weak:
		value = VALUE_WEAK
	}
	return value
}

// GetMDefMultiple 取得受到的傷害調整百分比
func (e *Effect) GetTakeDmgMultiple() float64 {
	value := 0.0
	switch e.Type {
	case gameJson.Protection:
		value = VALUE_PROTECTION
	}
	return value
}

// GetCritUpValue 取得爆擊率提升值
func (e *Effect) GetCritUpValue() float64 {
	addValue := 0.0
	var err error
	switch e.Type {
	case gameJson.CriticalUp:
		addValue, err = GetEffectValue[float64](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return 0
		}
	}
	return addValue
}

// GetInitUpValue 取得先攻提升值
func (e *Effect) GetInitUpValue() float64 {
	addValue := 0.0
	var err error
	switch e.Type {
	case gameJson.InitUp:
		addValue, err = GetEffectValue[float64](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return 0
		}
	}
	return addValue
}

// GetKnockbackUpValue 取得擊退提升值
func (e *Effect) GetKnockbackUpValue() float64 {
	addValue := 0.0
	var err error
	switch e.Type {
	case gameJson.KnockbackUp:
		addValue, err = GetEffectValue[float64](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return 0
		}
	}
	return addValue
}

// GetPDmgMultiple 取得物理傷害加乘
func (e *Effect) GetPDmgMultiple() float64 {
	value := 0.0
	switch e.Type {
	case gameJson.Vulnerable:
		value = VALUE_VULNERABLE
	}
	return value
}

// GetMDmgMultiple 取得魔法傷害加乘
func (e *Effect) GetMDmgMultiple() float64 {
	value := 0.0
	switch e.Type {
	case gameJson.Vulnerable:
		value = VALUE_VULNERABLE
	}
	return value
}

// GetRestoreHPValue 取得治癒值
func (e *Effect) GetRestoreHPValue() int {
	if e.Type != gameJson.RestoreHP {
		return 0
	}
	value, err := GetEffectValue[int](e, 0)
	if err != nil {
		log.Errorf("%v錯誤: %v", e.Type, err)
		return 0
	}
	return value
}

// GetRestoreVigorValue 取得體力回復值
func (e *Effect) GetRestoreVigorValue() float64 {
	if e.Type != gameJson.RestoreVigor {
		return 0
	}
	value, err := GetEffectValue[float64](e, 0)
	if err != nil {
		log.Errorf("%v錯誤: %v", e.Type, err)
		return 0
	}
	return value
}
