const Realm = require("realm");

const realmManager = require("./realm/realmManager");

const serverHost = '34.81.218.96';
const serverPort = 7654;

const net = require('net');

async function main() {

    // 帳號匿名註冊
    const user = await realmManager.AnonymousLogin()
    if (!user) {
        return;
    }

    // 呼叫AtlasFunction進行玩家資料初始化
    let result = await realmManager.CallAtlasFunc(user, "InitPlayerData", { "AuthType": "Guest" })
    console.log(`回應: ${result}`);


    // 建立socket連線
    const client = net.createConnection({ host: serverHost, port: serverPort }, () => {
        console.log('已連線到Server');

        const auth = {
            CMD: 'AUTH',
            Content: { Token: user.accessToken }
        };
        const packetBytes = JSON.stringify(auth);
        client.write(packetBytes + '\n'); // 送server
    });

    client.on('error', (err) => {
        console.error('連線Server錯誤: ', err);
        process.exit(1);
    });

    client.on('data', (data) => {
        console.log('接收: ', data.toString());
        processData(data);
    });

    client.on('end', () => {
        console.log('已斷線');
    });


}

function processData(data) {
    try {
        const msg = data.toString().trim();
        const pack = JSON.parse(msg);
        switch (pack.CMD) {
            case 'AUTH_TOCLIENT':
                if (pack.Content && pack.Content.IsAuth) {
                    console.log(`Authentication Status: ${pack.Content.IsAuth}, Token: ${pack.Content.ConnToken}`);
                } else {
                    console.error('Content轉型失败: AUTH_TOCLIENT');
                }
                break;
            default:
                console.error('尚未定義的 CMD: ', pack.CMD);
        }
    } catch (err) {
        console.error('解析Pack錯誤: ', err);
    }
}

main();
