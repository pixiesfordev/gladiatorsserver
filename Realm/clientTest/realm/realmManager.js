const Realm = require("realm");

const appConfig = require("../setting/realm.js");
const app = new Realm.App({ id: appConfig.appID });

module.exports = {
    AnonymousLogin,
    CallAtlasFunc,
};

function AnonymousLogin() {
    const credentials = Realm.Credentials.anonymous();
    return app.logIn(credentials).then(user => {
        console.log("匿名註冊成功 玩家ID: " + user.id);
        return user;
    }).catch(err => {
        console.error("匿名註冊失敗: " + err);
        throw err; // 保持错误向外传递，以便调用者可以处理
    });
}

async function CallAtlasFunc(user, funcName, args) {
    try {
        const result = await user.functions[funcName](args);
        // console.log(`Function ${funcName} 回應: ${result}`);
        return result;
    } catch (err) {
        console.error(`Call Function ${funcName} 錯誤: ${err}`);
        throw err;
    }
}

