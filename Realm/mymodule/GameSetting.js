module.exports = {
    EnvDB: Object.freeze({
        Dev: "gladiators",// 開發版
        Release: "???",// 正式版
    }),
    // 遊戲資料Json表
    JsonName: Object.freeze({
        GameSetting: "GameSetting",
        Gladiator: "Gladiator",
    }),
    // DB集合
    ColName: Object.freeze({
        player: "player",
        playerCustom: "playerCustom",
        playerState: "playerState",
        playerHistory: "playerHistory",
        gameSetting: "gameSetting",
        gameLog: "gameLog",
        template: "template",
    }),
    // 環境版本對應Endpoint
    AppEndpoint: Object.freeze({
        Dev: "https://asia-south1.gcp.data.mongodb-api.com/app/gladiators-pirlo",// 開發版
        Release: "???",// 正式版
    }),
    // GroupID就是ProjectID(在atlas app service左上方有垂直三個點那點Project Settings)
    // 也可以在開啟Atlas Services時 網址會顯示ProjectID
    // 在https://services.cloud.mongodb.com/groups/65b4b62b344719089d82ca3a/apps/65b4c6435b1a5d26443841cc/dashboard中
    EnvGroupID: Object.freeze({
        Dev: "65b4b62b344719089d82ca3a",// 開發版
        Release: "???",// 正式版
    }),
    // 環境版本對應AppID(在Atlas service->App Setting中)
    EnvAppID: Object.freeze({
        Dev: "gladiators-pirlo",// 開發版
        Release: "???",// 正式版
    }),
    // 環境版本對應AppObjID
    // App ObjectID跟AppID不一樣, 開啟Atlas Services時 網址會顯示App ObjectID
    // https://services.cloud.mongodb.com/groups/65b4b62b344719089d82ca3a/apps/65b4c6435b1a5d26443841cc/dashboard
    EnvAppObjID: Object.freeze({
        Dev: "65b4c6435b1a5d26443841cc",// 開發版
        Release: "???",// 正式版
    }),
    // 註冊類型
    AuthType: Object.freeze({
        Guest: "Guest",// 訪客
        Official: "Official",// 官方註冊
        Unknown: "Unknown",// 未知錯誤
    }),
    // 在線狀態
    OnlineState: Object.freeze({
        Online: "Online",// 在線
        Offline: "Offline",// 離線
    }),
    // 帳戶腳色(playerCustom中的腳色)
    PlayerCustomRole: Object.freeze({
        Player: "Player",// 玩家
        Developer: "Developer",// 開發者, 有更進階的DB訪問權限
    }),
    // 在線狀態
    LogType: Object.freeze({
        OnUserCreation: "OnUserCreation",// 玩家創Realm帳戶時會寫入此Log
        InitPlayerData: "InitPlayerData",// 玩家初始化玩家資料時會寫入此Log
        Signin: "Signin",// 玩家登入時寫入此Log
        PlayerVerify: "PlayerVerify",// 玩家登入時寫入此Log
    }),
    // 這邊要填入ColName的Key值, 如果template集合中有定義對應表的模板資料就要加在這裡
    ColTemplate: new Set(['player', 'playerState', 'playerHistory']),
}