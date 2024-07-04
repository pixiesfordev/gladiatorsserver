package main

import (
	gameJson "gladiatorsGoModule/gamejson"
	utility "gladiatorsGoModule/utility"
	"math"

	log "github.com/sirupsen/logrus"
)

type Gladiator struct {
	ID            string // DBGladiator的_id
	LeftSide      bool   // true是左方玩家(第一位), false是右方玩家(第二位)
	JsonGladiator gameJson.JsonGladiator
	JsonSkills    [6]gameJson.JsonSkill
	JsonTraits    []gameJson.TraitJsonData
	JsonEquips    []gameJson.JsonEquip
	Hp            int                               // 生命
	CurHp         int                               // 目前生命
	CurVigor      float64                           // 目前體力
	Str           int                               // 力量
	PDef          int                               // 物理防禦
	MDef          int                               // 魔法防禦
	Crit          float64                           // 爆擊率
	VigorRegen    float64                           // 體力回復
	Knockback     int                               // 擊退
	Init          int                               // 先攻
	Spd           int                               // 移動
	CurUnit       float64                           // 目前位置
	Effects       map[gameJson.EffectType][]*Effect // 狀態清單
}

func NewGladiator(id string, jsonGladiator gameJson.JsonGladiator, jsonSkills [6]gameJson.JsonSkill,
	jsonTraits []gameJson.TraitJsonData, jsonEquips []gameJson.JsonEquip) (Gladiator, error) {

	gladiator := Gladiator{
		ID:            id,
		JsonGladiator: jsonGladiator,
		JsonSkills:    jsonSkills,
		JsonTraits:    jsonTraits,
		JsonEquips:    jsonEquips,
		Hp:            jsonGladiator.HP,
		CurHp:         jsonGladiator.HP,
		CurVigor:      20,
		Str:           jsonGladiator.STR,
		PDef:          jsonGladiator.DEF,
		MDef:          jsonGladiator.MDEF,
		Crit:          jsonGladiator.CRIT,
		VigorRegen:    jsonGladiator.VigorRegen,
		Knockback:     jsonGladiator.Knockback,
		Init:          jsonGladiator.INIT,
		Spd:           jsonGladiator.Speed,
		Effects:       make(map[gameJson.EffectType][]*Effect, 0),
	}
	return gladiator, nil
}

// IsAlive 是否死亡
func (g *Gladiator) IsAlive() bool {
	return (g.CurHp > 0)
}

// CanMove 是否能移動
func (g *Gladiator) CanMove() bool {
	for _, effects := range g.Effects {
		if len(effects) != 0 && effects[0].IsMobileRestriction() {
			return false
		}
	}
	return true
}

// GetStr 取得力量
func (g *Gladiator) GetStr() int {
	str := g.Str
	// 計算影響力量的所有狀態
	addStr := 0
	for _, effects := range g.Effects {
		for _, v := range effects {
			addStr = v.GetStrUpValue()
		}
	}
	str += addStr
	return str
}

// GetPDef 取得物理防禦
func (g *Gladiator) GetPDef() int {
	pdef := g.PDef
	// 計算影響物理防禦的所有狀態
	addPDef := 0
	multiplePDef := 0.0
	for _, effects := range g.Effects {
		for _, v := range effects {
			addPDef += v.GetPDefUpValue()
			multiplePDef += v.GetPDefMultiple()
		}
	}
	pdef = int(math.Round(float64(pdef)*(1+multiplePDef) + float64(addPDef)))
	return pdef
}

// GetMDef 取得魔法防禦
func (g *Gladiator) GetMDef() int {
	mdef := g.MDef
	// 計算影響物理防禦的所有狀態
	addMDef := 0
	multipleMDef := 0.0
	for _, effects := range g.Effects {
		for _, v := range effects {
			addMDef += v.GetMDefUpValue()
			multipleMDef += v.GetPDefMultiple()
		}
	}
	mdef = int(math.Round(float64(mdef)*(1+multipleMDef) + float64(addMDef)))
	return mdef
}

// GetVigorRegen 取得體力回復
func (g *Gladiator) GetVigorRegen() float64 {
	vigorRegen := g.VigorRegen
	multipleVigorRegen := 0.0
	if _, ok := g.Effects[gameJson.Fatigue]; ok {
		multipleVigorRegen += FatigueValue
	}
	vigorRegen = float64(vigorRegen) * (1 + multipleVigorRegen)
	return vigorRegen
}

// AddEffect 新增狀態效果
func (g *Gladiator) AddEffect(effect *Effect) {
	g.Effects[effect.Type] = append(g.Effects[effect.Type], effect)
}

// RemoveEffects 移除一個或多個狀態效果
func (g *Gladiator) RemoveEffects(types ...gameJson.EffectType) {
	if len(types) == 0 {
		return
	}
	for _, t := range types {
		delete(g.Effects, t)
	}
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
			g.RemoveEffects(effectType)
		}
	}
}

// Spell 發動技能
func (myself *Gladiator) Spell(skill *Skill, target *Gladiator) {
	if myself == nil || skill == nil || target == nil {
		return
	}

	for _, effect := range skill.Effects {

		// 如果施法者或目標死亡就跳過
		if effect.Target.IsAlive() || !effect.Speller.IsAlive() {
			continue
		}

		// 如果沒觸發成功就跳過
		if !utility.GetProbResult(effect.Prob) {
			continue
		}

		switch effect.Type {
		case gameJson.PDmg: // 物理攻擊
			multiple := (1 + effect.GetPDmgMultiple())
			effectDmg := float64(effect.GetPDmgValue())
			str := float64(myself.GetStr())
			dmg := int(math.Round(str * effectDmg * multiple))
			myself.Attack(effect.Target, dmg, effect.Target.GetPDef())
		case gameJson.MDmg: // 魔法攻擊
			multiple := (1 + effect.GetMDmgMultiple())
			effectDmg := float64(effect.GetMDmgValue())
			str := float64(myself.GetStr())
			dmg := int(math.Round(str * effectDmg * multiple))
			myself.Attack(effect.Target, dmg, effect.Target.GetMDef())
		case gameJson.RestoreHP: // 回復生命
			effect.Target.AddHp(effect.GetRestoreHPValue())
		case gameJson.RestoreVigor: // 回復體力
			effect.Target.AddVigor(effect.GetRestoreVigorValue())
		default:
			if effect.IsBuffer() { // 賦予狀態
				myself.AddEffect(&effect)
			}
		}
	}

	log.Infof("勇士 %s 發動技能: %v", myself.ID, skill.JsonSkill.ID)
}

// TriggerBuffer_Time 時間性觸發Buffer
func (myself *Gladiator) TriggerBuffer_Time(value int) {
	if !myself.IsAlive() {
		return
	}
	for _, effects := range myself.Effects {
		for _, v := range effects {
			v.Trigger_Time()
		}
	}
}

// TriggerBuffer_BeHit 受擊時觸發Buffer
func (myself *Gladiator) TriggerBuffer_BeHit(value int) {
	if !myself.IsAlive() {
		return
	}
	for _, effects := range myself.Effects {
		for _, v := range effects {
			v.TriggerDmg_BeHit()
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
	if myself.CurHp < 0 {
		myself.CurHp = 0
	} else if myself.CurHp > MaxVigor {
		myself.CurHp = MaxVigor
	}
}

// OnDeath 死亡時觸發
func (myself *Gladiator) OnDeath() {

}

// Attack 造成傷害
func (myself *Gladiator) Attack(target *Gladiator, dmg int, def int) {
	dealDmg := dmg - def
	target.AddHp(dealDmg)
}
