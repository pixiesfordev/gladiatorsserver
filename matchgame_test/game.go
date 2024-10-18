package main

import (
	"context"
	"fmt"
	"matchgame/game"
	"unicode"

	// "os"
	// "os/signal"
	// 	"syscall"
	"sync"

	"time"

	"github.com/eiannone/keyboard"
	log "github.com/sirupsen/logrus"
)

var sm *skillManager

func initKeyboard() {
	sm = newSkillManager()
	ctx, _ := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go startKeyboard(ctx, &wg)
}

func startKeyboard(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := keyboard.Open(); err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	log.Info("鍵盤監聽開始")
	for {
		select {
		case <-ctx.Done():
			log.Info("鍵盤監聽結束")
			return
		default:
			log.Info("等待輸入...")
			char, _, err := keyboard.GetKey()
			if err != nil {
				log.Fatal(err)
			}
			switch unicode.ToUpper(char) {
			case 'Q':
				sm.clickSkill(0)
			case 'W':
				sm.clickSkill(1)
			case 'E':
				sm.clickSkill(2)
			}

		}
	}
}

func changeGameState(gameState game.GameState) {
	switch gameState {
	case game.GAMESTATE_WAITINGPLAYERS:
		go startPingLoop()
	}
}

func startPingLoop() {
	log.Infof("開始PING LOOP")
	ticker := time.NewTicker(1 * time.Second) // 每秒觸發一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			send_Ping()
		}
	}
}

type skillManager struct {
	skillIDs  []int
	skillOnID int
}

func newSkillManager() *skillManager {
	return &skillManager{
		skillIDs:  make([]int, 0),
		skillOnID: 0,
	}
}

// updateSkills 更新技能列表和啟用中的技能
func (sm *skillManager) updateSkills(skillIDs []int, skillOnID int) {
	sm.skillIDs = skillIDs
	sm.skillOnID = skillOnID

	logStr := "手牌: "
	for i, id := range skillIDs {
		logStr += fmt.Sprintf("%d", id)
		if i != len(skillIDs)-1 {
			logStr += ", "
		}
	}
	logStr += fmt.Sprintf("  啟用中的技能ID: %d", skillOnID)
	log.Infof(logStr)
}

func (sm *skillManager) clickSkill(idx int) {
	if idx < 0 || idx > 2 {
		return
	}
	skillID := sm.skillIDs[idx]
	log.Infof("點技能%d", skillID)
	if sm.skillOnID != skillID {
		// 啟用新的技能
		sm.skillOnID = skillID
		log.Infof("啟用技能%d", skillID)
	} else {
		// 關閉當前技能
		sm.skillOnID = 0
		log.Infof("關閉技能%d", skillID)
	}
	send_UseSkill(skillID)
}
