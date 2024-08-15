package game

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gladiatorsGoModule/gameJson"
)

type Gladiator struct {
	ID                    string // DBGladiator的_id
	LeftSide              bool   // 是否是左方玩家
	JsonGladiator         gameJson.JsonGladiator
	JsonSkills            [GladiatorSkillCount]gameJson.JsonSkill
	JsonTraits            []gameJson.TraitJson
	JsonEquips            []gameJson.JsonEquip
	HandSkills            [HandSkillCount]gameJson.JsonSkill
	Hp                    int
	CurHp                 int
	CurVigor              float64
	Str                   int                               // 力量
	PDef                  int                               // 物理防禦
	MDef                  int                               // 魔法防禦
	Crit                  float64                           // 爆擊率
	CritDmg               float64                           // 爆擊傷害
	VigorRegen            float64                           // 體力回復
	Init                  float64                           // 先攻
	Knockback             float64                           // 擊退
	Spd                   float64                           // 移動速度
	RushSpd               float64                           // 衝刺增加速度
	CurPos                float64                           // 目前位置
	IsRush                bool                              // 是否正在衝刺中
	Effects               map[gameJson.EffectType][]*Effect // 狀態清單
	ActivedMeleeJsonSkill *gameJson.JsonSkill               // 啟用中的肉搏技能, 玩家啟用中的肉搏技能, 如果是0代表沒有啟用中的肉搏技能

}

func NewTestGladiator() (Gladiator, error) {
	testGladiatorIdx := IDAccumulator.GetNextIdx("TestGladiatorID")
	testGladiatorID := fmt.Sprintf("gladiator%v", testGladiatorIdx)

	gJson, _ := gameJson.GetJsonGladiator(3)
	var jsonSkills [GladiatorSkillCount]gameJson.JsonSkill
	jsonSkill1, _ := gameJson.GetJsonSkill(3)
	jsonSkill2, _ := gameJson.GetJsonSkill(1001)
	jsonSkill3, _ := gameJson.GetJsonSkill(1003)
	jsonSkill4, _ := gameJson.GetJsonSkill(1006)
	jsonSkill5, _ := gameJson.GetJsonSkill(1008)
	jsonSkill6, _ := gameJson.GetJsonSkill(1010)
	jsonSkills[0] = jsonSkill1
	jsonSkills[1] = jsonSkill2
	jsonSkills[2] = jsonSkill3
	jsonSkills[3] = jsonSkill4
	jsonSkills[4] = jsonSkill5
	jsonSkills[5] = jsonSkill6

	return NewGladiator(testGladiatorID, gJson, jsonSkills, []gameJson.TraitJson{}, []gameJson.JsonEquip{})
}

func NewGladiator(id string, jsonGladiator gameJson.JsonGladiator, jsonSkills [GladiatorSkillCount]gameJson.JsonSkill,
	jsonTraits []gameJson.TraitJson, jsonEquips []gameJson.JsonEquip) (Gladiator, error) {
	pos := -InitGladiatorPos
	leftSide := true
	if MyRoom.GamerCount() > 1 {
		pos = InitGladiatorPos
		leftSide = false
	}
	gladiator := Gladiator{
		ID:            id,
		JsonGladiator: jsonGladiator,
		JsonSkills:    jsonSkills,
		JsonTraits:    jsonTraits,
		JsonEquips:    jsonEquips,
		LeftSide:      leftSide,
		Hp:            jsonGladiator.Hp,
		CurHp:         jsonGladiator.Hp,
		CurVigor:      MaxVigor,
		Str:           jsonGladiator.Str,
		PDef:          jsonGladiator.PDef,
		MDef:          jsonGladiator.MDef,
		Crit:          jsonGladiator.Crit,
		CritDmg:       jsonGladiator.CritDmg,
		VigorRegen:    jsonGladiator.VigorRegen,
		Knockback:     jsonGladiator.Knockback,
		Init:          jsonGladiator.Init,
		Spd:           jsonGladiator.Spd,
		RushSpd:       jsonGladiator.RushSpd,
		CurPos:        pos,
		IsRush:        false,
		Effects:       make(map[gameJson.EffectType][]*Effect, 0),
	}
	return gladiator, nil
}

// AddEffect 賦予狀態效果
func (g *Gladiator) AddEffect(effect *Effect) {
	//對要被賦予的狀態免疫時就返回
	immuneTypes := g.GetImmuneTypes()
	for i, _ := range immuneTypes {
		switch i {
		case Immune_MobileRestriction:
			if effect.IsImmuneToKnockback() {
				return
			}
		}
	}

	//當賦予的狀態會免疫某些類型狀態時, 要移除被免疫的狀態
	removeEffectTypes := make([]gameJson.EffectType, 0)
	switch effect.Type {
	case gameJson.Indomitable: // 免疫行動控制類的負面效果
		// 移除所有行動限制類負面效果
		for _, effects := range g.Effects {
			for _, v := range effects {
				if v.IsMobileRestriction() {
					removeEffectTypes = append(removeEffectTypes, v.Type)
					break
				}
			}
		}
	}
	g.RemoveEffects(removeEffectTypes)
	g.Effects[effect.Type] = append(g.Effects[effect.Type], effect)
}

// RemoveEffectsByBufferType 移除狀態
func (g *Gladiator) RemoveEffectsByBufferType(bufferType BufferType) {
	effectTypes := make([]gameJson.EffectType, 0)
	for effectType, v := range g.Effects {
		if len(v) != 0 && bufferType == v[0].GetBufferType() {
			effectTypes = append(effectTypes, effectType)
		}
	}
	if len(effectTypes) > 0 {
		g.RemoveEffects(effectTypes)
	}
}

// RemoveEffects 移除多個狀態效果
func (g *Gladiator) RemoveEffects(types []gameJson.EffectType) {
	if len(types) == 0 {
		return
	}
	for _, t := range types {
		delete(g.Effects, t)
	}
}

// RemoveEffects 移除一個或多個狀態效果
func (g *Gladiator) RemoveEffect(effectType gameJson.EffectType) {
	delete(g.Effects, effectType)
}

// RemoveSpecificEffect 移除指定的狀態效果
func (g *Gladiator) RemoveSpecificEffect(targetEffect *Effect) {
	for effectType, effects := range g.Effects {
		for i, effect := range effects {
			if effect == targetEffect {
				g.Effects[effectType] = append(effects[:i], effects[i+1:]...)
				break
			}
		}
		// 如果某個效果類型下沒有剩餘的效果，則從map中移除該類型
		if len(g.Effects[effectType]) == 0 {
			g.RemoveEffect(effectType)
		}
	}
}

// TriggerBuffer_Time 時間性觸發Buffer
func (myself *Gladiator) TriggerBuffer_Time() {
	if !myself.IsAlive() {
		return
	}
	for _, effects := range myself.Effects {
		for _, v := range effects {
			v.Trigger_Time()
		}
	}
}

// TriggerBuffer_AfterBeAttack 受擊後觸發Buffer
func (myself *Gladiator) TriggerBuffer_AfterBeAttack(dmg int) {
	if !myself.IsAlive() {
		return
	}
	for _, effects := range myself.Effects {
		for _, v := range effects {
			v.Trigger_AfterBeAttack(dmg)
		}
	}
}

// TriggerBuffer_AfterBeAttack 攻擊後觸發Buffer
func (myself *Gladiator) TriggerBuffer_AfterAttack(dmg int) {
	if !myself.IsAlive() {
		return
	}
	for _, effects := range myself.Effects {
		for _, v := range effects {
			v.Trigger_AfterAttack(dmg)
		}
	}
}

// AddHp 增加生命
func (myself *Gladiator) AddHp(value int) {
	if !myself.IsAlive() {
		return
	}
	myself.CurHp += value
	if myself.CurHp <= 0 {
		myself.CurHp = 0
		myself.OnDeath()
	}
}

// AddVigor 增加體力
func (myself *Gladiator) AddVigor(value float64) {
	if !myself.IsAlive() {
		return
	}
	myself.CurVigor += value
	if myself.CurVigor < 0 {
		myself.CurVigor = 0
	} else if myself.CurVigor > MaxVigor {
		myself.CurVigor = MaxVigor
	}
}

// OnDeath 死亡時觸發
func (myself *Gladiator) OnDeath() {

}

// GetSkill 傳入skillID取得目標JsonSkill
func (g *Gladiator) GetSkill(skillID int) (gameJson.JsonSkill, error) {
	for _, jsonSkill := range g.HandSkills {
		if jsonSkill.ID == skillID {
			return jsonSkill, nil
		}
	}
	log.Errorf("玩家選擇的技能不存在手牌技能中")
	return gameJson.JsonSkill{}, fmt.Errorf("玩家選擇的技能不存在手牌技能中")
}
