// main.js - 主要的JavaScript文件，用于初始化和协调客户端的各个组件

document.addEventListener('DOMContentLoaded', function() {
    // 尝试从localStorage加载认证信息
    const isAuthenticated = Auth.loadFromStorage();
    
    // 尝试恢复游戏会话
    const gameResumed = Game.tryResumeGame();
    
    // 如果已认证但没有恢复游戏，显示大厅
    if (isAuthenticated && !gameResumed) {
        document.getElementById('auth-screen').classList.add('hidden');
        document.getElementById('lobby-screen').classList.remove('hidden');
        document.getElementById('user-name').textContent = Auth.username;
    }
    
    // 添加开始游戏按钮事件（当有足够玩家时显示）
    const startGameBtn = document.createElement('button');
    startGameBtn.id = 'start-game-btn';
    startGameBtn.className = 'btn primary';
    startGameBtn.textContent = '开始游戏';
    startGameBtn.style.display = 'none'; // 初始隐藏
    
    // 将按钮添加到游戏信息区域
    document.querySelector('.game-info').appendChild(startGameBtn);
    
    // 添加开始游戏按钮事件
    startGameBtn.addEventListener('click', function() {
        Game.startGame();
        this.style.display = 'none';
    });
    
    // 定期检查是否可以开始游戏
    setInterval(function() {
        if (Game.canStartGame() && Game.gameState && Game.gameState.state === 'waiting') {
            startGameBtn.style.display = 'block';
        } else {
            startGameBtn.style.display = 'none';
        }
    }, 1000);
    
    // 添加WebSocket自动重连功能
    setInterval(function() {
        if (Game.socket && Game.socket.readyState === WebSocket.CLOSED) {
            console.log('尝试重新连接...');
            Game.reconnect();
        }
    }, 5000);
});