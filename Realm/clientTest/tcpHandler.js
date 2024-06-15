const net = require('net');
const { EventEmitter } = require('events');
const eventEmitter = new EventEmitter();


const serverHost = '34.81.218.96';
const serverPort = 7654;

// 创建一个 socket 连接到服务器
const client = net.createConnection({ host: serverHost, port: serverPort }, () => {
    console.log('已連線到Server');

    // 发送认证包
    const auth = {
        CMD: 'AUTH',
        Content: { Token: 'authRes.AccessToken' } // 替换为有效的 token
    };
    const packetBytes = JSON.stringify(auth);
    client.write(packetBytes + '\n'); // 发送数据到服务器
    console.log('封包已发送');
});

client.on('error', (err) => {
    console.error('连线Server错误: ', err);
    process.exit(1);
});

client.on('data', (data) => {
    console.log('收到数据: ', data.toString());
    processData(data);
});

client.on('end', () => {
    console.log('从服务器断开');
});

function processData(data) {
    try {
        const msg = data.toString().trim(); // 假设服务器发送的是 JSON 字符串 + 换行
        const pack = JSON.parse(msg);
        switch (pack.CMD) {
            case 'AUTH_TOCLIENT':
                if (pack.Content && pack.Content.IsAuth) {
                    console.log(`Authentication Status: ${pack.Content.IsAuth}, Token: ${pack.Content.ConnToken}`);
                } else {
                    console.error('Content转型失败: AUTH_TOCLIENT');
                }
                break;
            default:
                console.error('未定义的 CMD: ', pack.CMD);
        }
    } catch (err) {
        console.error('解析 Pack 错误: ', err);
    }
}