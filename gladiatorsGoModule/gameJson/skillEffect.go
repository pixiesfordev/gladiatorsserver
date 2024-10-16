package gameJson

import (
	"encoding/json"
	"fmt"
	// log "github.com/sirupsen/logrus"
)

type JsonSkillEffect struct {
	ID      string `json:"ID"`
	SkillID int    `json:"SkillID"`
	Target  string `json:"Target"`
	Effects []Effect
}

type Effect struct {
	Type  EffectType
	Value string
	Prob  float64
}

var SkillEffectDataDic = make(map[int][]JsonSkillEffect)

func (jsonData JsonSkillEffect) UnmarshalJSONData(jsonName string, jsonBytes []byte) (map[interface{}]interface{}, error) {
	var wrapper map[string][]map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &wrapper); err != nil {
		return nil, err
	}

	datas, ok := wrapper[jsonName]
	if !ok {
		return nil, fmt.Errorf("找不到key值: %s", jsonName)
	}

	items := make(map[interface{}]interface{})
	for _, rawEffect := range datas {
		json := JsonSkillEffect{}
		json.ID = rawEffect["ID"].(string)
		json.SkillID = int(rawEffect["SkillID"].(float64))
		json.Target = rawEffect["Target"].(string)

		// 處理Effect
		json.Effects = parseEffects(rawEffect)
		items[json.ID] = json
		AddToSkillEffectDataDic(json)
	}

	return items, nil
}

func parseEffects(data map[string]interface{}) []Effect {
	effects := []Effect{}
	for i := 1; ; i++ {
		typeKey := fmt.Sprintf("EffectType%d", i)
		valueKey := fmt.Sprintf("EffectValue%d", i)
		probKey := fmt.Sprintf("EffectProb%d", i)

		effectType, typeOk := data[typeKey]
		if !typeOk {
			break
		}

		effectValue := data[valueKey].(string)
		effectProb := data[probKey].(float64)

		sEffectType := effectType.(string)
		effects = append(effects, Effect{
			Type:  EffectType(sEffectType),
			Value: effectValue,
			Prob:  effectProb,
		})
	}
	return effects
}

func AddToSkillEffectDataDic(jsonSkillEffect JsonSkillEffect) {
	if len(SkillEffectDataDic[jsonSkillEffect.SkillID]) != 0 {
		SkillEffectDataDic[jsonSkillEffect.SkillID] = append(SkillEffectDataDic[jsonSkillEffect.SkillID], jsonSkillEffect)
	} else {
		SkillEffectDataDic[jsonSkillEffect.SkillID] = []JsonSkillEffect{jsonSkillEffect}
	}
}

type BufferType string

const (
	Buff    BufferType = "Buff"    // 正面效果
	Debuff             = "Debuff"  // 負面效果
	Neutral            = "Neutral" // 中性效果
)

// 效果類型
type EffectType string

const (
	PDmg                                      EffectType = "PDmg"                                      // 造成物理傷害,數值是造成多少傷害
	MDmg                                                 = "MDmg"                                      // 造成魔法傷害,數值是造成多少傷害
	RestoreHP                                            = "RestoreHP"                                 // 恢復生命,數值是恢復多少生命
	RestoreVigor                                         = "RestoreVigor"                              // 恢復體力,數值是恢復多少體力
	Rush                                                 = "Rush"                                      // 突擊,快速衝刺進行肉搏
	Pull                                                 = "Pull"                                      // 拉取,將敵人拉至前往自己
	Purge                                                = "Purge"                                     // 淨化,移除所有負面效果
	Shuffle                                              = "Shuffle"                                   // 洗牌,重洗目前的技能卡回牌庫並隨機抽回手牌
	Fortune                                              = "Fortune"                                   // 財富,獲得遊戲幣, 數值是增加值
	PermanentHp                                          = "PermanentHp"                               // 永久提升生命,數值填提昇值
	MeleeDmgReflect                                      = "MeleeDmgReflect"                           // [Buffer]肉搏傷害反射,受到肉搏攻擊會反彈受到的百分比傷害給攻擊者, 數值填0.5,3就是反彈50%傷害持續3秒
	Block                                                = "Block"                                     // [Buffer]格檔,格檔對手的攻擊，數值填3就是格檔3次
	Intuition                                            = "Intuition"                                 // [Buffer]直覺,顯示對手手牌技能
	Enraged                                              = "Enraged"                                   // [Buffer]激怒,強制進入突擊狀態, 層數是持續時間
	Dodge_RangeAttack                                    = "Dodge_RangeAttack"                         // [Buffer]迴避對手遠程攻擊, 層數是次數
	RegenHP                                              = "RegenHP"                                   // [Buffer]隨時間回復生命,數值填5,3代表每秒恢復5生命持續3秒(賦予時會立即觸發1次)
	RegenVigor                                           = "RegenVigor"                                // [Buffer]隨時間回復體力,數值填5,3代表每秒恢復5體力持續3秒(賦予時會立即觸發1次)
	Dizzy                                                = "Dizzy"                                     // [Buffer]暈眩,無法移動,無法施放立即技能, 層數是持續時間
	Poison                                               = "Poison"                                    // [Buffer]中毒,持續受到層數傷害直到戰鬥結束(賦予時會立即觸發1次)
	Bleeding                                             = "Bleeding"                                  // [Buffer]流血,受到攻擊額外受到層數傷害直到戰鬥結束
	Burning                                              = "Burning"                                   // [Buffer]著火,每數秒受到層數傷害直到層數歸0, 每次受到傷害後層數減半(賦予時會立即觸發1次)
	Fearing                                              = "Fearing"                                   // [Buffer]恐懼,往後逃跑, 層數是持續時間
	Vulnerable                                           = "Vulnerable"                                // [Buffer]無力,傷害減少30%, 層數是持續時間
	Weak                                                 = "Weak"                                      // [Buffer]虛弱,防禦減少30%, 層數是持續時間
	Fatigue                                              = "Fatigue"                                   // [Buffer]疲勞,體力回復減少50%, 層數是持續時間
	Protection                                           = "Protection"                                // [Buffer]加護,受到的傷害減少30%, 層數是持續時間
	MeleeSkillReflect                                    = "MeleeSkillReflect"                         // [Buffer]肉搏技能反擊,反彈下一次受到的肉搏技能, 數值是次數
	RangeSkillReflect                                    = "RangeSkillReflect"                         // [Buffer]遠程技能反擊,反彈下一次受到的遠程技能, 數值是次數
	PDefUp                                               = "PDefUp"                                    // [Buffer]物防提升,數值填5,3代表提升5持續3秒
	MDefUp                                               = "MDefUp"                                    // [Buffer]魔防提升,數值填5,3代表提升5持續3秒
	StrUp                                                = "StrUp"                                     // [Buffer]力量提升,數值填5,3代表提升5持續3秒
	KnockbackUp                                          = "KnockbackUp"                               // [Buffer]擊退提升,數值填5,3代表提升5持續3秒
	Barrier                                              = "Barrier"                                   // [Buffer]不動,無法被擊退, 層數是持續時間
	Poisoning                                            = "Poisoning"                                 // [Buffer]淬毒,攻擊會造成中毒, 層數是持續時間
	ComboAttack                                          = "ComboAttack"                               // [Buffer]連擊,額外技能攻擊次數, 層數是額外次數
	Vampire                                              = "Vampire"                                   // [Buffer]攻擊吸血,回復下次攻擊技能傷害的百分比生命, 數值填0.5,2代表回復造成傷害的50%生命, 下2次攻擊
	CriticalUp                                           = "CriticalUp"                                // [Buffer]爆擊提升,數值填0.5,3代表提升50%持續3秒
	InitUp                                               = "InitUp"                                    // [Buffer]先攻提升,數值填5,3代表提升5持續3秒
	Indomitable                                          = "Indomitable"                               // [Buffer]不屈,免疫行動控制類的負面效果, 層數是持續時間
	Berserk                                              = "Berserk"                                   // [Buffer]狂暴,進入無法控制的狂暴狀態, 力量上升50%, 數值是持續時間
	StrUpByHp                                            = "StrUpByHp"                                 // [Buffer]血怒,根據失去的生命百分比提升力量, 數值填最大提升值
	Chaos                                                = "Chaos"                                     // [Buffer]混沌,技能牌不會按照順序發牌, 數值填持續秒數
	SkillVigorUp                                         = "SkillVigorUp"                              // [Buffer]技能耗體增加,隨機一個技能本場戰鬥技能消耗體力增加, 填2,-1代表隨機2個技能體力需求-1
	StrBurst                                             = "StrBurst"                                  // [Buffer]力量爆發,填3就是下次的攻擊技能力量提升額外3倍
	TriggerEffect_BeAttack_StrUp                         = "TriggerEffect_BeAttack_StrUp"              // [Buffer]觸發效果-受到攻擊時, 力量上升,數值填力量上升值
	TriggerEffect_Time_Fortune                           = "TriggerEffect_Time_Fortune"                // [Buffer]觸發效果-每秒, 獲得遊戲幣,數值填獲得量
	TriggerEffect_WaitTime_RestoreVigor                  = "TriggerEffect_WaitTime_RestoreVigor"       // [Buffer]觸發效果-X秒後, 恢復體力,數值填5,10就是5秒後獲得10體力
	TriggerEffect_BattleResult_PermanentHp               = "TriggerEffect_BattleResult_PermanentHp"    // [Buffer]觸發效果-戰鬥結果觸發, 生命永久上升,數值填0,20代表戰鬥勝利時生命永久上升20,0為勝利,1為平手,2為戰敗
	TriggerEffect_SkillVigorBelow_ComboAttack            = "TriggerEffect_SkillVigorBelow_ComboAttack" // 觸發條件-下一技能耗體在X以下(包含X)觸發, 獲得ComboAttack數, 移除此Buffer,數值填3,2代表技能消耗在3以下獲得2層ComboAttack
	TriggerEffect_FirstAttack_Dodge                      = "TriggerEffect_FirstAttack_Dodge"           // 觸發條件-比對手先攻, 迴避對手攻擊, 移除此Buffer

)
