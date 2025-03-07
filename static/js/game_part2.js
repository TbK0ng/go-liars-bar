// game_part2.js - 游戏逻辑和WebSocket通信的后半部分

// 继续Game对象的方法定义

// 处理系统自动质疑
Game.handleSystemChallenge = function(message) {
    let logText = `系统自动质疑 ${message.playerName} 的手牌！`;
    
    if (message.challengeValid) {
        logText += ` 质疑成功！${message.playerName} 的手牌违规，将执行射击惩罚。`;
    } else {
        logText += ` 质疑失败！${message.playerName} 的手牌符合规则。`;
    }
    
    this.addLogEntry(logText);
};

// 处理游戏结束
Game.handleGameOver = function(message) {
    // 更新游戏状态
    document.getElementById('game-status-text').textContent = '游戏已结束';
    
    // 显示胜利者信息
    document.getElementById('winner-name').textContent = message.winnerName;
    
    // 切换到游戏结束界面
    document.getElementById('game-screen').classList.add('hidden');
    document.getElementById('game-over-screen').classList.remove('hidden');
    
    // 添加日志
    this.addLogEntry(`游戏结束！${message.winnerName} 获胜！`);
    
    // 断开WebSocket连接
    this.disconnect();
};

// 处理错误消息
Game.handleError = function(message) {
    // 显示错误消息
    alert(message.message || '发生错误');
    
    // 添加日志
    this.addLogEntry(`错误: ${message.message}`);
};

// 添加日志条目
Game.addLogEntry = function(text) {
    const logContainer = document.getElementById('log-container');
    const logEntry = document.createElement('div');
    logEntry.className = 'log-entry';
    
    // 添加时间戳
    const timestamp = new Date().toLocaleTimeString();
    logEntry.textContent = `[${timestamp}] ${text}`;
    
    // 添加到日志容器
    logContainer.appendChild(logEntry);
    
    // 滚动到底部
    logContainer.scrollTop = logContainer.scrollHeight;
};

// 发送出牌消息
Game.playCards = function(cards) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
        this.addLogEntry('无法出牌：与服务器的连接已断开');
        return;
    }
    
    // 发送出牌消息
    this.socket.send(JSON.stringify({
        type: 'play_cards',
        cards: cards
    }));
    
    // 隐藏出牌区域
    document.getElementById('play-cards-container').classList.add('hidden');
};

// 发送质疑消息
Game.sendChallenge = function(challenge, reason) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
        this.addLogEntry('无法发送质疑：与服务器的连接已断开');
        return;
    }
    
    // 发送质疑消息
    this.socket.send(JSON.stringify({
        type: 'challenge',
        challenge: challenge,
        reason: reason
    }));
};

// 发送开始游戏消息
Game.startGame = function() {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
        this.addLogEntry('无法开始游戏：与服务器的连接已断开');
        return;
    }
    
    // 发送开始游戏消息
    this.socket.send(JSON.stringify({
        type: 'start_game'
    }));
};

// 检查是否可以开始游戏（至少2名玩家）
Game.canStartGame = function() {
    if (!this.gameState || !this.gameState.players) return false;
    return Object.keys(this.gameState.players).length >= 2;
};

// 自动重连
Game.reconnect = function() {
    // 如果已经连接，先断开
    if (this.socket) {
        this.disconnect();
    }
    
    // 重新连接
    this.connect();
};

// 尝试从localStorage恢复游戏
Game.tryResumeGame = function() {
    const gameId = localStorage.getItem('currentGameId');
    const playerId = localStorage.getItem('currentPlayerId');
    
    if (gameId && playerId) {
        // 恢复游戏
        this.init(gameId, playerId);
        
        // 显示游戏界面
        document.getElementById('auth-screen').classList.add('hidden');
        document.getElementById('lobby-screen').classList.add('hidden');
        document.getElementById('game-screen').classList.remove('hidden');
        
        return true;
    }
    
    return false;
};