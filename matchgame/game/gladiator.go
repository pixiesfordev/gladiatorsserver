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
	IsAlive               bool                               // 是否存活
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
	CurPos                utility.Vector2                   // 目前位置
	FaceDir               utility.Vector2                   // 目前方向
	IsRush                bool                              // 是否正在衝刺中
	Effects               map[gameJson.EffectType][]*Effect // 狀態清單
	ActivedMeleeJsonSkill *gameJson.JsonSkill               // 啟用中的肉搏技能, 玩家啟用中的肉搏技能, 如果是0代表沒有啟用中的肉搏技能

}

func NewTestGladiator(owner Gamer, gladiatorID int, jsonSkillIDs []int) (Gladiator, error) {
	testGladiatorIdx := IDAccumulator.GetNextIdx()
	testGladiatorID := fmt.Sprintf("gladiator%v", testGladiatorIdx)

	gJson, _ := gameJson.GetJsonGladiator(gladiatorID)
	var jsonSkills [GladiatorSkillCount]gameJson.JsonSkill
	if len(jsonSkillIDs) == 6 { // 指定技能
		for i, skillID := range jsonSkillIDs {
			skill, err := gameJson.GetJsonSkill(skillID)
			if err != nil {
				return Gladiator{}, fmt.Errorf("gameJson.GetJsonSkill 錯誤: %v", err)
			}
			jsonSkills[i] = skill
		}
	} else {
		return Gladiator{}, fmt.Errorf("NewTestGladiator 錯誤: 傳入的jsonSkillIDs錯誤")
	}

	return NewGladiator(owner, testGladiatorID, gJson, jsonSkills, []gameJson.TraitJson{}, []gameJson.JsonEquip{})
}

func NewGladiator(owner Gamer, id string, jsonGladiator gameJson.JsonGladiator, jsonSkills [GladiatorSkillCount]gameJson.JsonSkill,
	jsonTraits []gameJson.TraitJson, jsonEquips []gameJson.JsonEquip) (Gladiator, error) {
	pos := GLADIATOR_POS_LEFT
	faceDir := utility.Vector2{X: 1, Y: 0}
	leftSide := true
	if MyRoom.GamerCount() > 1 {
		pos = GLADIATOR_POS_RIGHT
		leftSide = false
		faceDir = utility.Vector2{X: -1, Y: 0}
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
		IsAlive:       true,
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
		FaceDir:       faceDir,
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
	// log.Infof("AddEffect1 %v To %v", effect.Type, g.ID)
	// 不是狀態就返回
	if effect.BelongTo(NOTBUFFER) {
		return
	}
	//對要被賦予的狀態免疫時就返回
	for tag, _ := range effect.Tags {
		if g.ImmuneTo(tag) {
			// log.Infof("對 %v 效果的狀態: %v 免疫", effect.Type, tag)
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
	// log.Infof("AddEffect2 %v To %v", effect.Type, g.ID)
	switch effect.MyStackType {
	case STACKABLE: // 層數疊加
		if len(g.Effects[effect.Type]) > 0 {
			g.Effects[effect.Type][0].Duration += effect.Duration
			g.Effects[effect.Type][0].NextTriggerAt = GameTime
		} else {
			g.Effects[effect.Type] = append(g.Effects[effect.Type], effect)
		}
	case OVERRIDING: // 覆蓋
		if len(g.Effects[effect.Type]) > 0 {
			g.Effects[effect.Type][0] = effect
		} else {
			g.Effects[effect.Type] = append(g.Effects[effect.Type], effect)
		}
	case KEEPMAX: // 保留最大值
		if len(g.Effects[effect.Type]) > 0 {
			if g.Effects[effect.Type][0].Duration < effect.Duration {
				g.Effects[effect.Type][0] = effect
			}
		} else {
			g.Effects[effect.Type] = append(g.Effects[effect.Type], effect)
		}
	case ADDITIVE: // 新效果疊加
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
		// log.Infof("Remove Effect: %v", t)
		delete(g.Effects, t)
	}
}

// RemoveSpecificEffect 移除指定的狀態效果
func (g *Gladiator) RemoveSpecificEffect(targetEffect *Effect) {
	// log.Errorf("移除: %v", targetEffect.Type)
	for effectType, effects := range g.Effects {
		for i, effect := range effects {
			if effect == targetEffect {
				g.Effects[effectType] = append(effects[:i], effects[i+1:]...)
				break
			}
		}
		// 如果某個效果類型下沒有剩餘的效果，則從map中移除該類型
		if len(g.Effects[effectType]) == 0 {
			// log.Errorf("完全移除: %v", targetEffect.Type)
			g.RemoveEffects(effectType)
		}
	}
}

// TriggerBuffer_Time 時間性觸發Buffer
func (myself *Gladiator) TriggerBuffer_Time() {
	if !myself.IsAlive {
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
	if !myself.IsAlive {
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
	if !myself.IsAlive {
		return
	}
	for _, effects := range myself.Effects {
		for _, v := range effects {
			v.Trigger_AfterAttack(dmg)
		}
	}
}

// AddHp 增加生命
func (myself *Gladiator) AddHp(value int, effectType gameJson.EffectType, sendPack bool) {

	// 如果已經死亡就不處理
	if !myself.IsAlive {
		return
	}

	myself.CurHp += value
	if myself.CurHp <= 0 {
		myself.CurHp = 0
	} else if myself.CurHp > myself.Hp {
		myself.CurHp = myself.Hp
	}

	// 送client封包
	if sendPack {
		packState := packet.Pack{
			CMD: packet.HP_TOCLIENT,
			Content: packet.Hp_ToClient{
				PlayerID:   myself.Owner.GetID(),
				HPChange:   value,
				EffectType: string(effectType),
				CurHp:      myself.CurHp,
				MaxHp:      myself.Hp,
			},
		}
		MyRoom.BroadCastPacket(-1, packState)
	}

	if myself.IsAlive && myself.CurHp <= 0 { // 死亡
		myself.OnDeath()
	}

}

// AddVigor 增加體力
func (myself *Gladiator) AddVigor(value float64) {
	if !myself.IsAlive {
		return
	}
	// if value < 0 {
	// 	log.Errorf("AddVigor: %v", value)
	// }
	myself.CurVigor += value
	if myself.CurVigor < 0 {
		myself.CurVigor = 0
	} else if myself.CurVigor > MaxVigor {
		myself.CurVigor = MaxVigor
	}
}

// OnDeath 死亡時觸發
func (myself *Gladiator) OnDeath() {
	myself.IsAlive = false
	ChangeGameState(GAMESTATE_END, true)
	ResetGame("遊戲結束")
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

func (g *Gladiator) GetPackEffects() []packet.PackEffect {
	effectTypes := []packet.PackEffect{}
	for _, v := range g.Effects {
		if len(v) != 0 {
			switch v[0].MyStackType {
			case STACKABLE, OVERRIDING:
				effectTypes = append(effectTypes, packet.PackEffect{
					EffectName: string(v[0].Type),
					Duration:   utility.RoundToDecimal(v[0].Duration, 2),
				})
			case ADDITIVE:
				effectTypes = append(effectTypes, packet.PackEffect{
					EffectName: string(v[0].Type),
					Duration:   float64(len(v)),
				})
			default:
				log.Errorf("GetPackEffects 有尚未定義實作的StackType: %v", v[0].MyStackType)
			}
		}
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
		if p1, ok := g.Owner.(*Player); ok {
			p1.SendPacketToPlayer_SkillFail()
		}
		log.Errorf("%s UseSkill錯誤: %v", g.ID, err)
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
func (g *Gladiator) createBaseKnockSkill() (float64, *Skill, error) {
	spellInit := g.GetInit()
	skill, err := NewBaseKnockSkill(g, g.Opponent)
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

	copy(jsonSkills[:], skillSlice)

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

// GetHandSkillsByActivationType 取得手牌中指定發動類型的技能
func (g *Gladiator) GetHandSkillsByActivationType(activationType gameJson.ActivationType) []gameJson.JsonSkill {
	var jsonSkills = make([]gameJson.JsonSkill, 0)
	for i, v := range g.HandSkills {
		if i < 3 {
			if v.Activation == activationType {
				jsonSkills = append(jsonSkills, v)
			}
		}
	}
	return jsonSkills
}

// GetAvaliableHandSkillsByActivationType 取得手牌中 體力足夠的指定發動類型的技能
func (g *Gladiator) GetAvaliableHandSkillsByActivationType(activationType gameJson.ActivationType) []gameJson.JsonSkill {
	var typeSkills = g.GetHandSkillsByActivationType(activationType)
	jsonSkils := make([]gameJson.JsonSkill, 0)
	for _, v := range typeSkills {
		if v.Activation == gameJson.Instant && v.Vigor > int(g.CurVigor) { // 立即技能要確認體力是否足夠
			continue
		}
		jsonSkils = append(jsonSkils, v)
	}
	return jsonSkils
}
