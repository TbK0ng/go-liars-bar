// lobby.js - 处理游戏大厅相关功能

const Lobby = {
    // 创建新游戏
    createGame: async function() {
        try {
            const response = await fetch('/api/games', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${Auth.token}`
                }
            });
            
            const data = await response.json();
            
            if (data.gameId) {
                return { success: true, gameId: data.gameId };
            } else {
                return { success: false, message: '创建游戏失败' };
            }
        } catch (error) {
            console.error('创建游戏错误:', error);
            return { success: false, message: '创建游戏过程中发生错误' };
        }
    },
    
    // 加入游戏
    joinGame: async function(gameId, playerName) {
        try {
            const response = await fetch('/api/games/join', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${Auth.token}`
                },
                body: JSON.stringify({ gameId, playerName })
            });
            
            const data = await response.json();
            
            if (data.playerId) {
                return { success: true, playerId: data.playerId };
            } else {
                return { success: false, message: data.message || '加入游戏失败' };
            }
        } catch (error) {
            console.error('加入游戏错误:', error);
            return { success: false, message: '加入游戏过程中发生错误' };
        }
    },
    
    // 获取可用游戏列表（在实际应用中应该实现）
    getAvailableGames: async function() {
        // 这里应该调用API获取游戏列表
        // 简化版：返回模拟数据
        return [];
    }
};

// 页面加载时初始化大厅功能
document.addEventListener('DOMContentLoaded', function() {
    // 创建游戏按钮事件
    document.getElementById('create-game-btn').addEventListener('click', async function() {
        const result = await Lobby.createGame();
        
        if (result.success) {
            // 保存游戏ID
            localStorage.setItem('currentGameId', result.gameId);
            
            // 加入自己创建的游戏
            const joinResult = await Lobby.joinGame(result.gameId, Auth.username);
            
            if (joinResult.success) {
                // 保存玩家ID
                localStorage.setItem('currentPlayerId', joinResult.playerId);
                
                // 切换到游戏界面
                document.getElementById('lobby-screen').classList.add('hidden');
                document.getElementById('game-screen').classList.remove('hidden');
                
                // 初始化游戏
                Game.init(result.gameId, joinResult.playerId);
            } else {
                alert(joinResult.message);
            }
        } else {
            alert(result.message);
        }
    });
    
    // 加入游戏按钮事件
    document.getElementById('join-game-btn').addEventListener('click', async function() {
        const gameId = document.getElementById('game-id-input').value;
        
        if (!gameId) {
            alert('请输入游戏ID');
            return;
        }
        
        // 加入游戏
        const result = await Lobby.joinGame(gameId, Auth.username);
        
        if (result.success) {
            // 保存游戏ID和玩家ID
            localStorage.setItem('currentGameId', gameId);
            localStorage.setItem('currentPlayerId', result.playerId);
            
            // 切换到游戏界面
            document.getElementById('lobby-screen').classList.add('hidden');
            document.getElementById('game-screen').classList.remove('hidden');
            
            // 初始化游戏
            Game.init(gameId, result.playerId);
        } else {
            alert(result.message);
        }
    });
    
    // 返回大厅按钮事件（游戏结束界面）
    document.getElementById('back-to-lobby-btn').addEventListener('click', function() {
        document.getElementById('game-over-screen').classList.add('hidden');
        document.getElementById('lobby-screen').classList.remove('hidden');
        
        // 清除当前游戏信息
        localStorage.removeItem('currentGameId');
        localStorage.removeItem('currentPlayerId');
    });
    
    // 离开游戏按钮事件
    document.getElementById('leave-game-btn').addEventListener('click', function() {
        // 关闭WebSocket连接
        Game.disconnect();
        
        // 切换回大厅界面
        document.getElementById('game-screen').classList.add('hidden');
        document.getElementById('lobby-screen').classList.remove('hidden');
        
        // 清除当前游戏信息
        localStorage.removeItem('currentGameId');
        localStorage.removeItem('currentPlayerId');
    });
    
    // 加载可用游戏列表
    async function loadAvailableGames() {
        const games = await Lobby.getAvailableGames();
        const gamesContainer = document.getElementById('games-container');
        
        // 清空容器
        gamesContainer.innerHTML = '';
        
        if (games.length === 0) {
            gamesContainer.innerHTML = '<p>当前没有可用的游戏</p>';
            return;
        }
        
        // 添加游戏列表
        games.forEach(game => {
            const gameElement = document.createElement('div');
            gameElement.className = 'game-item';
            gameElement.innerHTML = `
                <div class="game-info">
                    <span>ID: ${game.id}</span>
                    <span>玩家: ${game.playerCount}/4</span>
                </div>
                <button class="btn secondary join-btn" data-game-id="${game.id}">加入</button>
            `;
            
            gamesContainer.appendChild(gameElement);
        });
        
        // 添加加入按钮事件
        document.querySelectorAll('.join-btn').forEach(btn => {
            btn.addEventListener('click', async function() {
                const gameId = this.getAttribute('data-game-id');
                const result = await Lobby.joinGame(gameId, Auth.username);
                
                if (result.success) {
                    // 保存游戏ID和玩家ID
                    localStorage.setItem('currentGameId', gameId);
                    localStorage.setItem('currentPlayerId', result.playerId);
                    
                    // 切换到游戏界面
                    document.getElementById('lobby-screen').classList.add('hidden');
                    document.getElementById('game-screen').classList.remove('hidden');
                    
                    // 初始化游戏
                    Game.init(gameId, result.playerId);
                } else {
                    alert(result.message);
                }
            });
        });
    }
    
    // 定期刷新游戏列表
    loadAvailableGames();
    setInterval(loadAvailableGames, 10000); // 每10秒刷新一次
});