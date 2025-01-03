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
		triggerDuration := 1.0 // 1秒觸發1次
		e.NextTriggerAt += triggerDuration
		value, err := GetEffectValue[int](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.AddDuration(-triggerDuration)
		e.Target.AddHp(value, e.Type, true)
	case gameJson.RegenVigor: // 回復體力
		triggerDuration := 1.0 // 1秒觸發1次
		e.NextTriggerAt += triggerDuration
		value, err := GetEffectValue[float64](e, 0)
		if err != nil {
			log.Errorf("%v錯誤: %v", e.Type, err)
			return
		}
		e.AddDuration(-triggerDuration)
		e.Target.AddVigor(value)
	case gameJson.Poison: // 中毒
		triggerDuration := 1.0 // 每1秒觸發1次
		e.NextTriggerAt += triggerDuration
		e.Target.AddHp(-int(e.Duration), e.Type, true)
	case gameJson.Burning: // 著火
		triggerDuration := 3.0 // 每3秒觸發1次
		e.NextTriggerAt += triggerDuration
		value := e.Duration                                    // 傷害為層數
		reduce := float64(int(e.Duration) - int(e.Duration/2)) // 每次減半
		e.AddDuration(-reduce)
		e.Target.AddHp(-int(value), e.Type, true)
	case gameJson.Enraged: // 激怒
		e.AddDuration(-TickTimePass)
		if e.Target != nil && !e.Target.IsRush {
			e.Target.SetRush(true)
		}

		// 隨時間消逝但沒有要執行特別效果的Buffer放這裡
	case gameJson.Dizzy, gameJson.Fearing, gameJson.Vulnerable, gameJson.Weak,
		gameJson.Fatigue, gameJson.Protection, gameJson.Indomitable, gameJson.Berserk, gameJson.Chaos, gameJson.PDefUp,
		gameJson.MDefUp, gameJson.StrUp, gameJson.KnockbackUp, gameJson.Barrier, gameJson.Poisoning, gameJson.CriticalUp,
		gameJson.InitUp:
		e.AddDuration(-TickTimePass)
	}

}

// Trigger_AfterBeAttack 受擊後觸發
func (e *Effect) Trigger_AfterBeAttack(dmg int) {
	// log.Infof("Trigger_AfterBeAttack :%v", e.Type)
	switch e.Type {
	case gameJson.Bleeding: // 流血
		// log.Infof("流血: %v", -int(e.Duration))
		e.Target.AddHp(-int(e.Duration), e.Type, true)
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
		e.Target.AddHp(restoreHP, e.Type, true)
	}
}
