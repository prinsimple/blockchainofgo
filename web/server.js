const express = require('express');
const { exec } = require('child_process');
const path = require('path');
const fs = require('fs');
const app = express();
const port = 3000;

// 配置静态文件服务
app.use(express.static(path.join(__dirname, 'public')));
app.use(express.json());

// 设置安全头
app.use((req, res, next) => {
    res.setHeader('Content-Security-Policy', "default-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net");
    next();
});

// 根路径处理
app.get('/', (req, res) => {
    res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

// 工具函数：执行命令并返回Promise
function execCommand(command) {
    return new Promise((resolve, reject) => {
        exec(command, { cwd: path.join(__dirname, '..'), shell: 'powershell.exe' }, (error, stdout, stderr) => {
            if (error) {
                console.error(`执行错误: ${error}`);
                if (stderr) console.error(`标准错误输出: ${stderr}`);
                
                // 如果是数据库锁定错误，返回空数组而不是报错
                if (stderr && stderr.includes("Cannot create lock file") && command.includes("getpending")) {
                    resolve("[]");
                    return;
                }
                
                reject({ error, stderr });
                return;
            }
            console.log(`标准输出: ${stdout}`);
            resolve(stdout);
        });
    });
}

// 工具函数：删除目录
function removeDirectory(dirPath) {
    const fullPath = path.join(__dirname, '..', dirPath);
    if (fs.existsSync(fullPath)) {
        fs.rmdirSync(fullPath, { recursive: true });
    }
}

// PoW区块链API
app.post('/api/pow/init', async (req, res) => {
    try {
        const { address } = req.body;
        // 确保数据库目录不存在
        removeDirectory('pow/db/blocks');
        // 创建创世区块
        const output = await execCommand(`go run main.go init ${address}`);
        res.json({ message: '创世区块已创建', output });
    } catch (err) {
        res.status(500).json({ error: '创建创世区块失败', details: err });
    }
});

app.get('/api/pow/blockchain', async (req, res) => {
    try {
        const output = await execCommand('go run main.go getchain');
        res.json({ data: output });
    } catch (err) {
        res.status(500).json({ error: '获取区块链数据失败', details: err });
    }
});

app.post('/api/pow/transaction', async (req, res) => {
    try {
        const { from, to, amount } = req.body;
        const output = await execCommand(`go run main.go send ${from} ${to} ${amount}`);
        res.json({ message: '交易已发送', output });
    } catch (err) {
        res.status(500).json({ error: '发送交易失败', details: err });
    }
});

app.post('/api/pow/mine', async (req, res) => {
    try {
        const output = await execCommand('go run main.go mine');
        res.json({ message: '挖矿成功', output });
    } catch (err) {
        res.status(500).json({ error: '挖矿失败', details: err });
    }
});

app.get('/api/pow/balance/:address', async (req, res) => {
    try {
        const { address } = req.params;
        const output = await execCommand(`go run main.go balance ${address}`);
        res.json({ data: output });
    } catch (err) {
        res.status(500).json({ error: '查询余额失败', details: err });
    }
});

// 获取挖矿统计信息
app.get('/api/pow/mining-stats', async (req, res) => {
    try {
        const output = await execCommand('go run main.go difficulty');
        const difficulty = parseInt(output);
        res.json({ 
            difficulty: difficulty,
            success: true 
        });
    } catch (err) {
        res.status(500).json({ error: '获取挖矿统计信息失败', details: err });
    }
});

// 更新挖矿难度
app.post('/api/pow/update-difficulty', async (req, res) => {
    try {
        const { difficulty } = req.body;
        if (!difficulty || difficulty < 1) {
            return res.status(400).json({ error: '无效的难度值' });
        }
        const output = await execCommand(`go run main.go setdifficulty -value ${difficulty}`);
        res.json({ 
            success: true,
            message: '难度更新成功',
            newDifficulty: difficulty
        });
    } catch (err) {
        res.status(500).json({ error: '更新挖矿难度失败', details: err });
    }
});

// PoS区块链API
app.post('/api/pos/init', async (req, res) => {
    try {
        const { address, stakeAmount } = req.body;
        // 确保数据库目录不存在
        removeDirectory('pos/db/blocks');
        // 创建创世区块
        const output = await execCommand(`go run pos/main.go init ${address} ${stakeAmount}`);
        res.json({ message: 'PoS创世区块已创建', output });
    } catch (err) {
        res.status(500).json({ error: '创建PoS创世区块失败', details: err });
    }
});

app.get('/api/pos/blockchain', async (req, res) => {
    try {
        const output = await execCommand('go run pos/main.go getchain');
        res.json({ data: output });
    } catch (err) {
        res.status(500).json({ error: '获取PoS区块链数据失败', details: err });
    }
});

app.post('/api/pos/transaction', async (req, res) => {
    try {
        const { from, to, amount } = req.body;
        const output = await execCommand(`go run pos/main.go send ${from} ${to} ${amount}`);
        res.json({ message: '交易已发送', output });
    } catch (err) {
        res.status(500).json({ error: '发送交易失败', details: err });
    }
});

app.post('/api/pos/stake', async (req, res) => {
    try {
        const { address, amount } = req.body;
        const output = await execCommand(`go run pos/main.go stake ${address} ${amount}`);
        res.json({ message: '质押成功', output });
    } catch (err) {
        res.status(500).json({ error: '质押失败', details: err });
    }
});

app.get('/api/pos/balance/:address', async (req, res) => {
    try {
        const { address } = req.params;
        const output = await execCommand(`go run pos/main.go balance ${address}`);
        res.json({ data: output });
    } catch (err) {
        res.status(500).json({ error: '查询余额失败', details: err });
    }
});

// 获取待处理交易
app.get('/api/pow/pending-transactions', async (req, res) => {
    try {
        const output = await execCommand('go run main.go getpending');
        console.log('待处理交易池数据:', output);
        
        // 如果输出为空或者是空字符串，返回空数组
        if (!output || output.trim() === '') {
            return res.json({ transactions: [] });
        }

        // 尝试解析JSON
        try {
            const transactions = JSON.parse(output);
            res.json({ transactions: Array.isArray(transactions) ? transactions : [] });
        } catch (parseError) {
            console.error('解析待处理交易失败:', parseError);
            res.json({ transactions: [] });
        }
    } catch (err) {
        // 如果是数据库锁定错误，返回空数组
        if (err.stderr && err.stderr.includes("Cannot create lock file")) {
            return res.json({ transactions: [] });
        }
        console.error('获取待处理交易失败:', err);
        res.status(500).json({ error: '获取待处理交易失败', details: err });
    }
});

app.listen(port, () => {
    console.log(`服务器运行在 http://localhost:${port}`);
});