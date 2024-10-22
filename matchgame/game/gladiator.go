package game

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
	"gladiatorsGoModule/utility"
	"matchgame/packet"

	log "github.com/sirupsen/logrus"
)

type Gladiator struct {
	Owner                 Gamer      // 擁有Gamer
	ID                    string     // DBGladiator的_id
	LeftSide              bool       // 是否是左方玩家
	Opponent              *Gladiator // 對手
	JsonGladiator         gameJson.JsonGladiator
	JsonSkills            [GladiatorSkillCount]gameJson.JsonSkill // 起始打亂的技能牌
	JsonTraits            []gameJson.TraitJson
	JsonEquips            []gameJson.JsonEquip
	HandSkills            [HandSkillCount]gameJson.JsonSkill // 手牌
	Deck                  []gameJson.JsonSkill               // 牌庫的牌
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

func NewTestGladiator(owner Gamer) (Gladiator, error) {
	testGladiatorIdx := IDAccumulator.GetNextIdx()
	testGladiatorID := fmt.Sprintf("gladiator%v", testGladiatorIdx)

	gJson, _ := gameJson.GetJsonGladiator(1)
	var jsonSkills [GladiatorSkillCount]gameJson.JsonSkill
	jsonSkill1, _ := gameJson.GetJsonSkill(1)
	jsonSkill2, _ := gameJson.GetJsonSkill(1001)
	jsonSkill3, _ := gameJson.GetJsonSkill(1002)
	jsonSkill4, _ := gameJson.GetJsonSkill(1003)
	jsonSkill5, _ := gameJson.GetJsonSkill(1004)
	jsonSkill6, _ := gameJson.GetJsonSkill(1005)
	jsonSkills[0] = jsonSkill1
	jsonSkills[1] = jsonSkill2
	jsonSkills[2] = jsonSkill3
	jsonSkills[3] = jsonSkill4
	jsonSkills[4] = jsonSkill5
	jsonSkills[5] = jsonSkill6

	return NewGladiator(owner, testGladiatorID, gJson, jsonSkills, []gameJson.TraitJson{}, []gameJson.JsonEquip{})
}

func NewGladiator(owner Gamer, id string, jsonGladiator gameJson.JsonGladiator, jsonSkills [GladiatorSkillCount]gameJson.JsonSkill,
	jsonTraits []gameJson.TraitJson, jsonEquips []gameJson.JsonEquip) (Gladiator, error) {
	pos := -InitGladiatorPos
	leftSide := true
	if MyRoom.GamerCount() > 1 {
		pos = InitGladiatorPos
		leftSide = false
	}

	// 取得亂洗後的手牌與牌庫
	handSkills, deck := GetHandsAndDeck(jsonSkills, true)
	gladiator := Gladiator{
		ID:            id,
		Owner:         owner,
		JsonGladiator: jsonGladiator,
		JsonSkills:    jsonSkills,
		JsonTraits:    jsonTraits,
		JsonEquips:    jsonEquips,
		HandSkills:    handSkills,
		Deck:          deck,
		LeftSide:      leftSide,
		Hp:            jsonGladiator.Hp,
		CurHp:         jsonGladiator.Hp,
		CurVigor:      DefaultVigor,
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

// AddPassiveEffect  賦予被動效果
func (g *Gladiator) AddPassiveEffect(effect *Effect) {
}

// AddEffect 賦予狀態效果
func (g *Gladiator) AddEffect(effect *Effect) {
	log.Infof("AddEffect1 %v To %v", effect.Type, g.ID)
	// 不是狀態就返回
	if effect.BelongTo(NOTBUFFER) {
		return
	}
	//對要被賦予的狀態免疫時就返回
	for tag, _ := range effect.Tags {
		if g.ImmuneTo(tag) {
			return
		}
	}

	//當賦予的狀態會免疫某些類型狀態時, 要移除被免疫的狀態
	removeEffectTypes := make([]gameJson.EffectType, 0)
	for immueTag, _ := range effect.ImmueTags {
		for _, v := range g.Effects {
			if len(v) != 0 && v[0].BelongTo(immueTag) {
				removeEffectTypes = append(removeEffectTypes, v[0].Type)
			}
		}
	}
	g.RemoveEffects(removeEffectTypes...)
	log.Infof("AddEffect2 %v To %v", effect.Type, g.ID)
	switch effect.MyStackType {
	case STACKABLE:
		if len(g.Effects[effect.Type]) > 0 {
			g.Effects[effect.Type][0].Duration += effect.Duration
		} else {
			g.Effects[effect.Type] = append(g.Effects[effect.Type], effect)
		}
	case OVERRIDING:
		if len(g.Effects[effect.Type]) > 0 {
			g.Effects[effect.Type][0] = effect
		} else {
			g.Effects[effect.Type] = append(g.Effects[effect.Type], effect)
		}
	case ADDITIVE:
		g.Effects[effect.Type] = append(g.Effects[effect.Type], effect)
	default:
		log.Errorf("尚未定義的StackType %v", effect.MyStackType)
	}

}

// RemoveEffectsByTag 根據標籤移除Effect
func (g *Gladiator) RemoveEffectsByTag(tag Tag) {
	removeEffectTypes := make([]gameJson.EffectType, 0)
	for _, v := range g.Effects {
		if len(v) != 0 && v[0].BelongTo(tag) {
			removeEffectTypes = append(removeEffectTypes, v[0].Type)
		}
	}
	g.RemoveEffects(removeEffectTypes...)
}

// RemoveEffects 移除多個狀態效果
func (g *Gladiator) RemoveEffects(types ...gameJson.EffectType) {
	if len(types) == 0 {
		return
	}
	for _, t := range types {
		log.Infof("Remove Effect: %v", t)
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
	log.Infof("TriggerBuffer_AfterBeAttack")
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
func (myself *Gladiator) AddHp(value int, sendPack bool) {
	if !myself.IsAlive() {
		return
	}
	myself.CurHp += value
	if myself.CurHp <= 0 {
		myself.CurHp = 0
		myself.OnDeath()
	}

	// 送client封包
	if sendPack {
		packState := packet.Pack{
			CMD: packet.HP_TOCLIENT,
			Content: packet.Hp_ToClient{
				PlayerID: myself.Owner.GetID(),
				HPChange: value,
				CurHp:    myself.CurHp,
				MaxHp:    myself.Hp,
			},
		}
		MyRoom.BroadCastPacket(-1, packState)
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

// GetSkill 傳入skillID取得目標JsonSkill與索引
func (g *Gladiator) GetSkill(skillID int) (gameJson.JsonSkill, int, error) {

	for i, v := range g.HandSkills {
		if i < 3 {
			if skillID == v.ID {
				return v, i, nil
			}
		}
	}
	log.Errorf("使用技能%v 手牌%v", skillID, g.HandSkills)
	log.Errorf("玩家選擇的技能不存在手牌技能中: %v", skillID)
	return gameJson.JsonSkill{}, -1, fmt.Errorf("玩家選擇的技能不存在手牌技能中: %v", skillID)
}

// GetPackPlayerState 取得玩家的狀態封包
func (g *Gladiator) GetPackHp(myselfPack bool) packet.Hp_ToClient {
	packPlayerState := packet.Hp_ToClient{
		PlayerID: g.Owner.GetID(),
		HPChange: 0,
		CurHp:    g.CurHp,
		MaxHp:    g.Hp,
	}
	return packPlayerState
}

func (g *Gladiator) GetEffectStrs() []string {
	effectTypes := []string{}
	for k := range g.Effects {
		effectTypes = append(effectTypes, string(k))
	}
	return effectTypes
}

// GetActiveSkill 取得啟用中的肉搏技能
func (g *Gladiator) GetActiveSkill() *Skill {
	if g.ActivedMeleeJsonSkill != nil {
		skill, err := NewSkill(g, g.Opponent, *g.ActivedMeleeJsonSkill)
		if err != nil {
			log.Errorf("NewSkill錯誤")
		}
		return skill
	}
	return nil
}

// GetHandSkills 取得手牌技能IDs
func (g *Gladiator) GetHandSkills() [4]int {
	var handSkills [4]int
	for i, v := range g.HandSkills {
		if i < 4 {
			handSkills[i] = v.ID
		} else {
			break
		}
	}
	return handSkills
}

func (g *Gladiator) createSkill(jsonSKill gameJson.JsonSkill) (float64, *Skill, error) {
	spellInit := g.GetInit()
	var skill *Skill
	err := g.UseSkill(jsonSKill.ID)
	if err != nil {
		log.Errorf("%s.UseSkill錯誤", g.ID)
		return spellInit, nil, err
	}
	skill, err = NewSkill(g, g.Opponent, jsonSKill)
	if err != nil {
		log.Errorf("NewSkill錯誤")
		return spellInit, nil, err
	}
	spellInit += skill.JsonSkill.Init
	return spellInit, skill, nil
}

// SetSkills 根據技能ID設定角鬥士技能
func (g *Gladiator) SetSkillByIDs(skillIDs [6]int) error {
	var jsonSkills [GladiatorSkillCount]gameJson.JsonSkill
	skillSlice, err := gameJson.GetJsonSkillsByIDs(skillIDs[:])
	if err != nil {
		return fmt.Errorf("獲取技能失敗: %v", err)
	}

	if len(skillSlice) != GladiatorSkillCount {
		return fmt.Errorf("技能數量不正確，期望 %d，實際 %d", GladiatorSkillCount, len(skillSlice))
	}

	for i, skill := range skillSlice {
		jsonSkills[i] = skill
	}

	handSkills, deck := GetHandsAndDeck(jsonSkills, false)
	g.HandSkills = handSkills
	g.Deck = deck

	return nil
}

// 傳入角鬥士技能, 取得手牌與牌庫
func GetHandsAndDeck(jsonSkills [GladiatorSkillCount]gameJson.JsonSkill, shuffle bool) ([HandSkillCount]gameJson.JsonSkill, []gameJson.JsonSkill) {
	var shuffledSkills []gameJson.JsonSkill
	if shuffle {
		// 亂洗技能順序
		shuffledSkills = utility.Shuffle(jsonSkills[:])
	} else {
		shuffledSkills = jsonSkills[:]
	}

	var handSkills [HandSkillCount]gameJson.JsonSkill // 手牌
	var deck []gameJson.JsonSkill                     // 牌庫的牌
	for i := 0; i < GladiatorSkillCount; i++ {
		if i < HandSkillCount {
			handSkills[i] = shuffledSkills[i]
		} else {
			deck = append(deck, shuffledSkills[i])
		}
	}
	return handSkills, deck
}
