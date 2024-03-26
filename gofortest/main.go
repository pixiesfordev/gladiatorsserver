package main

import (
	"fmt"
	"gladiatorsGoModule/gameJson"
)

func main() {
	err := gameJson.Init("Dev")
	if err != nil {
		fmt.Printf("初始化失敗: %v\n", err)
		return
	}
	// gladiator1, err := gameJson.GetGladiatorByID("1")
	// if err != nil {
	// 	fmt.Printf("取資料錯誤: %v", err)
	// }

}
