package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
)

type Effect struct {
	Type          gameJson.EffectType // 效果類型
	ValueStr      string              // 效果數值
	Duration      float64             // 持續時間/次數
	Speller       *Gladiator          // 施法者
	Target        *Gladiator          // 目標
	Prob          float64             // 觸發機率
	NextTriggerAt float64             // 下次觸發時間
	MyStackType   StackType           // Buffer堆疊類型
	Tags          map[Tag]bool        // 所屬標籤
	ImmueTags     map[Tag]bool        // 免疫標籤
}

// 效果數值設定
const (
	VALUE_VULNERABLE float64 = -0.3 // 無力造成傷害減少百分比
	VALUE_WEAK       float64 = -0.3 // 虛弱造成防禦減少百分比
	VALUE_FATIGUE    float64 = -0.3 // 疲勞造成體力回復減少
	VALUE_PROTECTION float64 = -0.3 // 加護減傷百分比
	VALUE_BERSERK    float64 = 0.5  // 狂暴增加力量百分比
)

// StackType Buffer堆疊類型
type StackType string

const (
	STACKABLE  StackType = "STACKABLE"  // 堆疊型-同樣的新Buffer會累加Duration
	OVERRIDING           = "OVERRIDING" // 覆蓋型-同樣的新Buffer會完全取代舊Buffer
	ADDITIVE             = "ADDITIVE"   // 擴增型-同樣的新Buffer會擴增一個Buffer
)

func NewEffect(effectType gameJson.EffectType, valueStr string, speller *Gladiator, target *Gladiator, prob float64, isPassive bool) (*Effect, error) {
	duration := 0.0
	tags := make(map[Tag]bool)
	immueTags := make(map[Tag]bool)
	var stackType StackType = STACKABLE

	values, err := utility.Split_FLOAT(valueStr, ",")
	if err != nil {
		return nil, err
	}

	switch effectType {
	case gameJson.PDmg:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[PDMG] = true
	case gameJson.MDmg:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[MDMG] = true
	case gameJson.TrueDmg:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[TDMG] = true
	case gameJson.RestoreHP:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[RESTORE_HP] = true
	case gameJson.RestoreVigor:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[RESTORE_VIGOR] = true
	case gameJson.Rush:
		tags[MOVE] = true
	case gameJson.Pull:
	case gameJson.Purge:
	case gameJson.Shuffle:
	case gameJson.Fortune:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
	case gameJson.PermanentHp:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		// 以下為Buffer類
	case gameJson.MeleeDmgReflect:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[1]
		tags[BUFF] = true
	case gameJson.Block:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[BUFF] = true
	case gameJson.Intuition:
		tags[BUFF] = true
		stackType = OVERRIDING
	case gameJson.Enraged:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[DEBUFF] = true
		stackType = STACKABLE
	case gameJson.Dodge_RangeAttack:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.RegenHP:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[1]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.RegenVigor:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[1]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.Dizzy:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[DEBUFF] = true
		stackType = STACKABLE
		immueTags[MOVE] = true
	case gameJson.Poison:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[DEBUFF] = true
		stackType = STACKABLE
	case gameJson.Bleeding:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[DEBUFF] = true
		stackType = ADDITIVE
	case gameJson.Burning:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[DEBUFF] = true
		stackType = ADDITIVE
	case gameJson.Fearing:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[DEBUFF] = true
		stackType = STACKABLE
	case gameJson.Protection:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.MeleeSkillReflect:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.RangeSkillReflect:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.PDefUp:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[1]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.MDefUp:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[1]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.StrUp:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[1]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.KnockbackUp:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[1]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.Barrier:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[BUFF] = true
		stackType = STACKABLE
		immueTags[KNOCKBACK] = true
	case gameJson.Poisoning:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.ComboAttack:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.Vampire:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[1]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.CriticalUp:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[1]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.InitUp:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[1]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.Indomitable:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.Berserk:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[NEUTRAL] = true
		immueTags[SKILL_DIVINE] = true
		stackType = STACKABLE
	case gameJson.StrUpByHp:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[BUFF] = true
		stackType = ADDITIVE
	case gameJson.Chaos:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		duration = values[0]
		tags[DEBUFF] = true
		stackType = STACKABLE
	case gameJson.SkillVigorUp:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		if values[1] > 0 {
			tags[DEBUFF] = true
		} else if values[1] < 0 {
			tags[BUFF] = true
		} else {
			tags[NEUTRAL] = true
		}
		stackType = ADDITIVE
	case gameJson.StrBurst:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.TriggerEffect_BeAttack_StrUp:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.TriggerEffect_Time_Fortune:
		if len(values) != 1 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[BUFF] = true
		stackType = STACKABLE
	case gameJson.TriggerEffect_WaitTime_RestoreVigor:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[BUFF] = true
		stackType = ADDITIVE
	case gameJson.TriggerEffect_BattleResult_PermanentHp:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[BUFF] = true
		stackType = ADDITIVE
	case gameJson.TriggerEffect_SkillVigorBelow_ComboAttack:
		if len(values) != 2 {
			return nil, fmt.Errorf("effectype %v 參數填入錯誤", effectType)
		}
		tags[BUFF] = true
		stackType = ADDITIVE
	case gameJson.TriggerEffect_FirstAttack_Dodge:
		tags[BUFF] = true
		stackType = OVERRIDING
	}
	if isPassive {
		tags[PASSIVE] = true
	}

	// 如果不是Buff類型Effect就加入NOTBUFFER標籤
	if !tags[BUFF] && !tags[DEBUFF] && !tags[NEUTRAL] {
		tags[NOTBUFFER] = true
	}

	effect := Effect{
		Type:          effectType,
		ValueStr:      valueStr,
		Duration:      duration,
		Speller:       speller,
		Target:        target,
		Prob:          prob,
		NextTriggerAt: GameTime,
		MyStackType:   stackType,
		Tags:          tags,
		ImmueTags:     immueTags,
	}
	return &effect, nil
}

// BelongTo 此Buffer是否屬於某Tag類型
func (e *Effect) BelongTo(tag Tag) bool {
	return e.Tags[tag]
}

// ImmuneTo 此Buffer是否免疫某Tag類型
func (e *Effect) ImmuneTo(tags ...Tag) bool {
	for _, tag := range tags {
		if e.ImmueTags[tag] {
			return true
		}
	}
	return false
}

// IsExpired 效果是否過期
func (e *Effect) IsExpired() bool {
	return e.Duration <= 0
}

// AddDuration 增加持續時間
func (e *Effect) AddDuration(value float64) {
	e.Duration += value
	if e.IsExpired() && e.Target != nil {
		e.Target.RemoveSpecificEffect(e)
	}
}
