package game

import (
	"gladiatorsGoModule/gameJson"
	"math"

	log "github.com/sirupsen/logrus"
)

// Trigger_Time 時間觸發
func (e *Effect) Trigger_Time() {
	if e.BelongTo(NOTBUFFER) || e.NextTriggerAt > GameTime {
		return
	}
	switch e.Type {
	case gameJson.RegenHP: // 回復生命
		e.NextTriggerAt += INTERVAL_EFFECT_TRIGGER_SECS // 更新觸發時間
		value, err := GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.AddDuration(-INTERVAL_EFFECT_TRIGGER_SECS)
		e.Target.AddHp(value, true)
	case gameJson.RegenVigor: // 回復體力
		e.NextTriggerAt += INTERVAL_EFFECT_TRIGGER_SECS // 更新觸發時間
		value, err := GetEffectValue[float64](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.AddDuration(-INTERVAL_EFFECT_TRIGGER_SECS)
		e.Target.AddVigor(value)
	case gameJson.Poison: // 中毒
		e.NextTriggerAt += INTERVAL_EFFECT_TRIGGER_SECS // 更新觸發時間
		value, err := GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.AddDuration(-INTERVAL_EFFECT_TRIGGER_SECS)
		e.Target.AddHp(-value, true)
	case gameJson.Burning: // 著火
		e.NextTriggerAt += INTERVAL_EFFECT_TRIGGER_SECS * 5 // 更新觸發時間
		value, err := GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		reduce := e.Duration / 2
		if reduce == 0 {
			reduce = 1
		}
		e.AddDuration(-reduce)
		log.Infof("著火: %v", value)
		e.Target.AddHp(-value, true)

		// 隨時間消逝但沒有要執行特別效果的Buffer放這裡
	case gameJson.Dizzy, gameJson.Enraged, gameJson.Fearing, gameJson.Vulnerable, gameJson.Weak,
		gameJson.Fatigue, gameJson.Protection, gameJson.Indomitable, gameJson.Berserk, gameJson.Chaos, gameJson.PDefUp,
		gameJson.MDefUp, gameJson.StrUp, gameJson.KnockbackUp, gameJson.Barrier, gameJson.Poisoning, gameJson.CriticalUp,
		gameJson.InitUp:

		e.NextTriggerAt += INTERVAL_EFFECT_TRIGGER_SECS // 更新觸發時間
		e.AddDuration(-INTERVAL_EFFECT_TRIGGER_SECS)

	}

}

// Trigger_AfterBeAttack 受擊後觸發
func (e *Effect) Trigger_AfterBeAttack(dmg int) {
	switch e.Type {
	case gameJson.Bleeding: // 流血
		value, err := GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.Target.AddHp(-value, true)
	}
}

// Trigger_AfterAttack 攻擊後觸發
func (e *Effect) Trigger_AfterAttack(dmg int) {
	switch e.Type {
	case gameJson.Vampire: // 吸血
		value, err := GetEffectValue[float64](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.AddDuration(-1)
		restoreHP := int(math.Round(value * float64(dmg)))
		e.Target.AddHp(restoreHP, true)
	}
}
