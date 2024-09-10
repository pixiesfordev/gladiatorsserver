package game

import (
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"matchgame/packet"
	"math"
	// log "github.com/sirupsen/logrus"
)

func (gladiator *Gladiator) GetPackGladiator(myselfPack bool) packet.PackGladiator {
	var jsonSkillIDs [GladiatorSkillCount]int
	var handSkills [HandSkillCount]int
	curVigor := 0.0
	for i, v := range gladiator.JsonSkills {
		jsonSkillIDs[i] = v.ID
	}

	// 玩家自己才需要知道資料
	if myselfPack {
		for i, v := range gladiator.HandSkills {
			handSkills[i] = v.ID
		}
		curVigor = gladiator.CurVigor
	}

	packGladiator := packet.PackGladiator{
		DBID:         gladiator.ID,
		JsonID:       gladiator.JsonGladiator.ID,
		SkillIDs:     jsonSkillIDs,
		HandSkillIDs: handSkills,
		MaxHP:        gladiator.Hp,
		CurHp:        gladiator.CurHp,
		CurVigor:     curVigor,
		CurSpd:       gladiator.Spd,
		CurPos:       gladiator.CurPos,
		EffectTypes:  []string{},
	}
	return packGladiator
}

func (gladiator *Gladiator) GetPackGladiatorState(myselfPack bool) packet.PackGladiatorState {
	var handSkills [HandSkillCount]int
	activedMeleeJsonSkillID := 0
	if myselfPack {
		for i, _ := range handSkills {
			handSkills[i] = gladiator.HandSkills[i].ID
		}
		if gladiator.ActivedMeleeJsonSkill != nil {
			activedMeleeJsonSkillID = gladiator.ActivedMeleeJsonSkill.ID
		}
	}

	effectTypes := []string{}
	for k := range gladiator.Effects {
		effectTypes = append(effectTypes, string(k))
	}

	packGladiator := packet.PackGladiatorState{
		HandSkillIDs:            handSkills,
		CurHp:                   gladiator.CurHp,
		CurVigor:                utility.RoundToDecimal(gladiator.CurVigor, 3),
		CurSpd:                  utility.RoundToDecimal(gladiator.Spd, 3),
		CurPos:                  utility.RoundToDecimal(gladiator.CurPos, 3),
		EffectTypes:             effectTypes,
		ActivedMeleeJsonSkillID: activedMeleeJsonSkillID,
	}
	return packGladiator
}

// IsAlive 是否死亡
func (g *Gladiator) IsAlive() bool {
	return (g.CurHp > 0)
}

func (g *Gladiator) GetDir() float64 {
	if g.LeftSide {
		return 1
	} else {
		return -1
	}
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

// GetSpd 取得速度
func (g *Gladiator) GetSpd() float64 {
	spd := g.Spd
	if g.IsRush {
		spd += g.RushSpd
	}
	return spd
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
