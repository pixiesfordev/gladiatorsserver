package main

import (
	"gladiatorsGoModule/gameJson"
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
	CritDmg       float64                           // 爆擊傷害
	VigorRegen    float64                           // 體力回復
	Knockback     float64                           // 擊退
	Init          float64                           // 先攻
	Spd           float64                           // 移動速度
	RushSpd       float64                           // 衝刺增加速度
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
		Hp:            jsonGladiator.Hp,
		CurHp:         jsonGladiator.Hp,
		CurVigor:      20,
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

// Knockbackable 是否能被擊退
func (g *Gladiator) Knockbackable() bool {
	for _, effects := range g.Effects {
		if len(effects) != 0 && effects[0].IsImmuneToKnockback() {
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

// GetCrtit 取得爆擊率
func (g *Gladiator) GetCrit() float64 {
	crit := g.Crit
	// 計算影響爆擊率的所有狀態
	addCirt := 0.0
	for _, effects := range g.Effects {
		for _, v := range effects {
			addCirt = v.GetCritUpValue()
		}
	}
	crit += addCirt
	return crit
}

// GetKnockback 取得擊退值
func (g *Gladiator) GetKnockback() float64 {
	knockback := g.Knockback
	// 計算影響擊退的所有狀態
	addKnockback := 0.0
	for _, effects := range g.Effects {
		for _, v := range effects {
			addKnockback = v.GetKnockbackUpValue()
		}
	}
	knockback += addKnockback
	return knockback
}

// GetInit 取得先攻
func (g *Gladiator) GetInit() float64 {
	init := g.Init
	// 計算影響先攻的所有狀態
	addInit := 0.0
	for _, effects := range g.Effects {
		for _, v := range effects {
			addInit = v.GetInitUpValue()
		}
	}
	init += addInit
	return init
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
			multipleMDef += v.GetMDefMultiple()
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

// GetImmuneTypes 取得體力回復
func (g *Gladiator) GetImmuneTypes() map[ImmuneType]struct{} {

	immuneTypes := make(map[ImmuneType]struct{})

	for _, effects := range g.Effects {
		for _, v := range effects {
			if v.IsImmuneToKnockback() {
				immuneTypes[Immune_Knockback] = struct{}{}
				break
			}
			if v.IsImmuneToMobileRestriction() {
				immuneTypes[Immune_MobileRestriction] = struct{}{}
				break
			}
		}
	}
	return immuneTypes
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

// 執行擊退
func (myself *Gladiator) DoKnockback(knockbackValue float64) {

	//對擊退免疫就返回
	immuneTypes := myself.GetImmuneTypes()
	if _, ok := immuneTypes[Immune_Knockback]; ok {
		return
	}

	// 執行擊退位移
	if myself.LeftSide {
		myself.CurUnit -= knockbackValue
	} else {
		myself.CurUnit += knockbackValue
	}
}

// Spell 發動技能
func (myself *Gladiator) Spell(skill *Skill) {
	if myself == nil || skill == nil {
		return
	}

	// 檢查是否通過技能施法條件
	if !myself.IsPassingSpellCondition(skill) {
		return
	}

	// 執行技能效果
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
		case gameJson.Purge: // 移除負面狀態
			effect.Target.RemoveEffectsByBufferType(Debuff)
		default:
			if effect.IsBuffer() { // 賦予狀態
				myself.AddEffect(&effect)
			}
		}
	}

	log.Infof("勇士 %s 發動技能: %v", myself.ID, skill.JsonSkill.ID)

	// 如果有ComboAttack就重複觸發技能
	if effects, ok := myself.Effects[gameJson.ComboAttack]; ok {
		if len(effects) != 0 {
			effects[0].AddDuration(-1)
			myself.Spell(skill)
		}
	}

}

// IsPassingSpellCondition 施法條件檢查, 全部偷過才返回true(代表可以觸發)
func (myself *Gladiator) IsPassingSpellCondition(skill *Skill) bool {
	if !myself.IsAlive() {
		return false
	}
	for _, effects := range myself.Effects {
		for _, v := range effects {
			switch v.Type {
			case gameJson.Condition_SkillVigorBelow:
				value, err := GetEffectValue[int](v, 0)
				if err != nil {
					log.Errorf("%v錯誤: %v", v.Type, err)
					return false
				}
				if skill.Vigor > value {
					return false
				}
			}
		}
	}
	return true
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

	// 計算目標狀態減傷
	multiple := 0.0
	for _, effects := range target.Effects {
		for _, v := range effects {
			multiple += v.GetTakeDmgMultiple()
		}
	}
	dealDmg = int(math.Round(float64(dealDmg) * (1 + multiple)))

	// 計算爆擊
	crit := myself.GetCrit()
	extraCritDmg := 0.0
	if crit > 1 { // 溢出的爆擊率要加到爆擊傷害上
		extraCritDmg = 1 - crit
	}
	if utility.GetProbResult(myself.GetCrit()) {
		dealDmg = int(math.Round(float64(dealDmg) * (myself.CritDmg + extraCritDmg)))
	}

	target.TriggerBuffer_AfterBeAttack(dealDmg)
	myself.TriggerBuffer_AfterAttack(dealDmg)

	target.AddHp(dealDmg)
}
