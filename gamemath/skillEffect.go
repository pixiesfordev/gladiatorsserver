package main

import (
	"fmt"
	gameJson "gladiatorsGoModule/gamejson"
	"math"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Effect struct {
	Type          gameJson.EffectType // 效果類型
	Value         interface{}         // 效果數值(根據效果類型定義結構)
	Duration      int                 // 持續時間/次數
	Speller       *Gladiator          // 施法者
	Target        *Gladiator          // 目標
	Prob          float64             // 觸發機率
	NextTriggerAt float64             // 下次觸發時間
}

// 效果數值設定
const (
	VulnerableValue       float64 = -0.3 // 無力造成傷害減少
	WeakValue             float64 = -0.3 // 虛弱造成防禦減少
	FatigueValue          float64 = -0.3 // 疲勞造成體力回復減少
	EffectTriggerInterval int     = 1    // 時間性觸發效果間格時間, 例如流血, 填1就是每秒觸發1次
)

// GetEffectValue 取得第X個T類數值
func GetEffectValue[T any](e *Effect, idx int) (T, error) {
	var defaultT T
	values := strings.Split(e.Value.(string), ",")
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

type BufferType string

// Buffer類型定義
const (
	Buff    BufferType = "Buff"    // 正面效果
	Debuff             = "Debuff"  // 負面效果
	Neutral            = "Neutral" // 中性效果
)

// IsBuffer 是否為狀態效果
func (e *Effect) IsBuffer() bool {
	switch e.Type {
	case gameJson.RegenHP, gameJson.RegenVigor, gameJson.Dizzy, gameJson.Poison, gameJson.Bleeding, gameJson.Burning,
		gameJson.Fearing, gameJson.Vulnerable, gameJson.Weak, gameJson.Fatigue, gameJson.Protection, gameJson.MeleeSkillReflect, gameJson.RangeSkillReflect,
		gameJson.Enraged, gameJson.PDefUp, gameJson.MDefUp, gameJson.StrUp, gameJson.Barrier, gameJson.Poisoning, gameJson.ComboAttack, gameJson.Vampire, gameJson.CriticalUp,
		gameJson.InitUp, gameJson.Indomitable, gameJson.Berserk, gameJson.StrUpByHp, gameJson.Chaos, gameJson.SkillVigorUp:
		return true
	}
	return false
}

// IsMobileRestriction 是否為移動限制類效果
func (e *Effect) IsMobileRestriction() bool {
	if !e.IsBuffer() {
		return false
	}
	switch e.Type {
	case gameJson.Fearing, gameJson.Dizzy, gameJson.Pull:
		return true
	}
	return false
}

// IsInstantSkillRestriction 是否為立即技能限制類效果
func (e *Effect) IsInstantSkillRestriction() bool {
	if !e.IsBuffer() {
		return false
	}
	switch e.Type {
	case gameJson.Dizzy:
		return true
	}
	return false
}

// GetBufferType 取得Buffer類型
func (e *Effect) GetBufferType() BufferType {
	switch e.Type {
	case gameJson.Bleeding, gameJson.Dizzy:
		return Buff
	case gameJson.Vampire:
		return Debuff
	default:
		return Neutral
	}
}

// IsExpired 效果是否過期
func (e *Effect) IsExpired() bool {
	if e.Duration == -1 { // -1是永久性Buffer
		return false
	}
	return e.Duration <= 0
}

// AddStack 增加效果層數或持續時間
func (e *Effect) AddStack(value int) {
	if e.IsExpired() {
		e.Target.RemoveEffects(e.Type)
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
		hpRatio := (float64)(e.Target.CurHp) / (float64)(e.Target.Hp)
		addValue = int(math.Round(hpRatio * value))
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
		value = WeakValue
	}
	return value
}

// GetMDefMultiple 取得魔法防禦加乘
func (e *Effect) GetMDefMultiple() float64 {
	value := 0.0
	switch e.Type {
	case gameJson.Weak:
		value = WeakValue
	}
	return value
}

// GetPDmgValue 取得物理傷害
func (e *Effect) GetPDmgValue() int {
	if e.Type != gameJson.PDmg {
		return 0
	}
	value, err := GetEffectValue[int](e, 0)
	if err != nil {
		log.Errorf("%v錯誤: %v", e.Type, err)
		return 0
	}
	return value
}

// GetMDmgValue 取得魔法傷害
func (e *Effect) GetMDmgValue() int {
	if e.Type != gameJson.PDmg {
		return 0
	}
	value, err := GetEffectValue[int](e, 0)
	if err != nil {
		log.Errorf("%v錯誤: %v", e.Type, err)
		return 0
	}
	return value
}

// GetPDmgMultiple 取得物理傷害加乘
func (e *Effect) GetPDmgMultiple() float64 {
	value := 0.0
	switch e.Type {
	case gameJson.Vulnerable:
		value = VulnerableValue
	}
	return value
}

// GetMDmgMultiple 取得魔法傷害加乘
func (e *Effect) GetMDmgMultiple() float64 {
	value := 0.0
	switch e.Type {
	case gameJson.Vulnerable:
		value = VulnerableValue
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

// Trigger_Time 時間觸發
func (e *Effect) Trigger_Time() {
	if !e.IsBuffer() || e.NextTriggerAt >= GameTimer {
		return
	}
	switch e.Type {
	case gameJson.RegenHP: // 回復生命
		e.NextTriggerAt += float64(EffectTriggerInterval) // 更新觸發時間
		value, err := GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.Duration -= EffectTriggerInterval
		e.Target.AddHp(value)
	case gameJson.RegenVigor: // 回復體力
		e.NextTriggerAt += float64(EffectTriggerInterval) // 更新觸發時間
		value, err := GetEffectValue[float64](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.Duration -= EffectTriggerInterval
		e.Target.AddVigor(value)
	case gameJson.Poison: // 中毒
		e.NextTriggerAt += float64(EffectTriggerInterval) // 更新觸發時間
		value, err := GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.Duration -= EffectTriggerInterval
		e.Target.AddHp(-value)
	case gameJson.Burning: // 著火
		e.NextTriggerAt += float64(EffectTriggerInterval) // 更新觸發時間
		value, err := GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.Duration = e.Duration / 2
		e.Target.AddHp(-value)
	case gameJson.Bleeding: // 不會隨時間消逝的Buffer放這裡

	default:

		e.Duration -= EffectTriggerInterval
	}

}

// TriggerDmg_BeHit 受擊觸發傷害
func (e *Effect) TriggerDmg_BeHit() {
	switch e.Type {
	case gameJson.Bleeding: // 流血
		value, err := GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.Duration -= 1
		e.Target.AddHp(-value)
	}
}
