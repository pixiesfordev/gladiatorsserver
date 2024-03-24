let db = db.getSiblingDB('gladiators');

// 刪除整個template collection
db.template.deleteMany({});

let nowDate = new Date();
// 開始插入模板
let playerDoc = db.template.insertMany([
  // 模板-玩家資料
  {
    _id: "player",
    createdAt: nowDate,
    authType: "Guest",
    point: 1000000,//NumberLong("1")
    onlineState: "Offline",
    lastSigninAt_nowDate: null,
    lastSignoutAt_nowDate: null,
    ban: false,
    deviceUID: "",
    leftGameAt_nowDate: null,
    inMatchgameID: "",
    redisSync: true,
    heroExp: 0,
    spellCharges: [0, 0, 0],
    drops: [0, 0, 0],
  },
  // 模板-玩家狀態
  {
    _id: "playerState",
    createdAt: nowDate,
    lastUpdateAt_nowDate: null,
  },
  // 模板-玩家歷程
  {
    _id: "playerHistory",
    createdAt: nowDate,
  }
]);

printjson(playerDoc);