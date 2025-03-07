// game.js - 处理游戏逻辑和WebSocket通信

const Game = {
    socket: null,
    gameId: null,
    playerId: null,
    gameState: null,
    selectedCards: [],
    
    // 初始化游戏
    init: function(gameId, playerId) {
        this.gameId = gameId;
        this.playerId = playerId;
        this.selectedCards = [];
        
        // 更新游戏状态显示
        document.getElementById('game-status-text').textContent = '正在连接服务器...';
        
        // 连接WebSocket
        this.connect();
        
        // 初始化事件监听器
        this.initEventListeners();
    },
    
    // 连接WebSocket
    connect: function() {
        // 创建WebSocket连接
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?gameId=${this.gameId}&playerId=${this.playerId}`;
        
        this.socket = new WebSocket(wsUrl);
        
        // 设置WebSocket事件处理函数
        this.socket.onopen = this.handleSocketOpen.bind(this);
        this.socket.onmessage = this.handleSocketMessage.bind(this);
        this.socket.onclose = this.handleSocketClose.bind(this);
        this.socket.onerror = this.handleSocketError.bind(this);
    },
    
    // 断开WebSocket连接
    disconnect: function() {
        if (this.socket) {
            this.socket.close();
            this.socket = null;
        }
    },
    
    // 初始化事件监听器
    initEventListeners: function() {
        // 出牌按钮事件
        document.getElementById('play-cards-btn').addEventListener('click', () => {
            if (this.selectedCards.length > 0) {
                this.playCards(this.selectedCards);
                this.selectedCards = [];
                document.getElementById('selected-cards').innerHTML = '';
                document.getElementById('play-cards-btn').disabled = true;
            }
        });
        
        // 质疑按钮事件
        document.getElementById('challenge-yes-btn').addEventListener('click', () => {
            const reason = document.getElementById('challenge-reason').value;
            this.sendChallenge(true, reason);
            document.getElementById('challenge-container').classList.add('hidden');
        });
        
        // 不质疑按钮事件
        document.getElementById('challenge-no-btn').addEventListener('click', () => {
            const reason = document.getElementById('challenge-reason').value || '不质疑';
            this.sendChallenge(false, reason);
            document.getElementById('challenge-container').classList.add('hidden');
        });
    },
    
    // WebSocket连接打开时的处理函数
    handleSocketOpen: function() {
        console.log('WebSocket连接已建立');
        document.getElementById('game-status-text').textContent = '已连接到游戏服务器';
        
        // 发送加入游戏消息
        this.socket.send(JSON.stringify({
            type: 'join_game',
            gameId: this.gameId,
            playerId: this.playerId
        }));
    },
    
    // 处理WebSocket消息
    handleSocketMessage: function(event) {
        const message = JSON.parse(event.data);
        console.log('收到消息:', message);
        
        // 根据消息类型处理
        switch (message.type) {
            case 'game_state':
                this.updateGameState(message.state);
                break;
                
            case 'your_turn':
                this.handleYourTurn(message);
                break;
                
            case 'play_action':
                this.handlePlayAction(message);
                break;
                
            case 'challenge_request':
                this.handleChallengeRequest(message);
                break;
                
            case 'challenge_result':
                this.handleChallengeResult(message);
                break;
                
            case 'shooting_result':
                this.handleShootingResult(message);
                break;
                
            case 'system_challenge':
                this.handleSystemChallenge(message);
                break;
                
            case 'game_over':
                this.handleGameOver(message);
                break;
                
            case 'error':
                this.handleError(message);
                break;
        }
    },
    
    // WebSocket连接关闭时的处理函数
    handleSocketClose: function() {
        console.log('WebSocket连接已关闭');
        document.getElementById('game-status-text').textContent = '与服务器的连接已断开';
    },
    
    // WebSocket错误处理函数
    handleSocketError: function(error) {
        console.error('WebSocket错误:', error);
        document.getElementById('game-status-text').textContent = '连接错误';
    },
    
    // 更新游戏状态
    updateGameState: function(state) {
        this.gameState = state;
        
        // 更新游戏状态显示
        let statusText = '';
        switch (state.state) {
            case 'waiting':
                statusText = '等待玩家加入...';
                break;
            case 'starting':
                statusText = '游戏即将开始...';
                break;
            case 'playing':
                const currentPlayerName = this.getCurrentPlayerName();
                statusText = `游戏进行中 - 轮到 ${currentPlayerName} 出牌`;
                break;
            case 'finished':
                statusText = '游戏已结束';
                break;
        }
        document.getElementById('game-status-text').textContent = statusText;
        
        // 更新目标牌
        if (state.targetCard) {
            document.getElementById('target-card').textContent = state.targetCard;
        } else {
            document.getElementById('target-card').textContent = '?';
        }
        
        // 更新玩家手牌
        this.updatePlayerHand();
        
        // 更新其他玩家信息
        this.updateOtherPlayers();
    },
    
    // 获取当前玩家名称
    getCurrentPlayerName: function() {
        if (!this.gameState || !this.gameState.players) return '';
        
        const currentPlayerIdx = this.gameState.currentPlayerIdx;
        if (currentPlayerIdx >= 0 && currentPlayerIdx < this.gameState.playerOrder.length) {
            const currentPlayerId = this.gameState.playerOrder[currentPlayerIdx];
            return this.gameState.players[currentPlayerId].name;
        }
        return '';
    },
    
    // 更新玩家手牌显示
    updatePlayerHand: function() {
        if (!this.gameState || !this.gameState.players) return;
        
        const playerHandElement = document.getElementById('player-hand');
        playerHandElement.innerHTML = '';
        
        const player = this.gameState.players[this.playerId];
        if (!player || !player.hand) return;
        
        // 显示玩家手牌
        player.hand.forEach(card => {
            const cardElement = document.createElement('div');
            cardElement.className = 'card';
            cardElement.textContent = card;
            cardElement.addEventListener('click', () => this.toggleCardSelection(cardElement, card));
            playerHandElement.appendChild(cardElement);
        });
    },
    
    // 更新其他玩家信息
    updateOtherPlayers: function() {
        if (!this.gameState || !this.gameState.players) return;
        
        const otherPlayersElement = document.getElementById('other-players');
        otherPlayersElement.innerHTML = '';
        
        // 创建其他玩家信息
        for (const playerId in this.gameState.players) {
            if (playerId === this.playerId) continue; // 跳过当前玩家
            
            const player = this.gameState.players[playerId];
            const playerElement = document.createElement('div');
            playerElement.className = 'other-player';
            
            // 设置玩家状态样式
            if (!player.alive) {
                playerElement.classList.add('dead');
            }
            
            // 当前回合玩家高亮显示
            if (this.gameState.playerOrder[this.gameState.currentPlayerIdx] === playerId) {
                playerElement.classList.add('current-turn');
            }
            
            playerElement.innerHTML = `
                <div class="player-name">${player.name} ${player.alive ? '' : '(已死亡)'}</div>
                <div class="player-cards">手牌数量: ${player.handCount || 0}</div>
            `;
            
            otherPlayersElement.appendChild(playerElement);
        }
    },
    
    // 切换卡牌选择状态
    toggleCardSelection: function(cardElement, card) {
        // 检查是否是当前玩家的回合
        if (!this.isCurrentPlayerTurn()) return;
        
        // 切换卡牌选择状态
        if (cardElement.classList.contains('selected')) {
            // 取消选择
            cardElement.classList.remove('selected');
            this.selectedCards = this.selectedCards.filter(c => c !== card);
        } else {
            // 选择卡牌
            cardElement.classList.add('selected');
            this.selectedCards.push(card);
        }
        
        // 更新选中的卡牌显示
        this.updateSelectedCards();
        
        // 更新出牌按钮状态
        document.getElementById('play-cards-btn').disabled = this.selectedCards.length === 0;
    },
    
    // 更新选中的卡牌显示
    updateSelectedCards: function() {
        const selectedCardsElement = document.getElementById('selected-cards');
        selectedCardsElement.innerHTML = '';
        
        this.selectedCards.forEach(card => {
            const cardElement = document.createElement('div');
            cardElement.className = 'card';
            cardElement.textContent = card;
            selectedCardsElement.appendChild(cardElement);
        });
    },
    
    // 检查是否是当前玩家的回合
    isCurrentPlayerTurn: function() {
        if (!this.gameState) return false;
        
        const currentPlayerIdx = this.gameState.currentPlayerIdx;
        if (currentPlayerIdx >= 0 && currentPlayerIdx < this.gameState.playerOrder.length) {
            return this.gameState.playerOrder[currentPlayerIdx] === this.playerId;
        }
        return false;
    },
    
    // 处理轮到当前玩家出牌
    handleYourTurn: function(message) {
        // 显示出牌区域
        document.getElementById('play-cards-container').classList.remove('hidden');
        
        // 添加日志
        this.addLogEntry(`轮到你出牌了`);
    },
    
    // 处理玩家出牌行为
    handlePlayAction: function(message) {
        // 隐藏出牌区域
        document.getElementById('play-cards-container').classList.add('hidden');
        
        // 添加日志
        let logText = `玩家 ${message.playerName} 出了 ${message.cardCount} 张牌`;
        if (message.playedCards) {
            logText += ` (${message.playedCards.join(', ')})`;
        }
        this.addLogEntry(logText);
    },
    
    // 处理质疑请求
    handleChallengeRequest: function(message) {
        // 显示质疑区域
        document.getElementById('challenge-container').classList.remove('hidden');
        
        // 设置质疑信息
        document.getElementById('challenging-player').textContent = message.playerName;
        document.getElementById('card-count').textContent = message.cardCount;
        document.getElementById('challenge-reason').value = '';
        
        // 添加日志
        this.addLogEntry(`玩家 ${message.playerName} 出了 ${message.cardCount} 张牌，你可以选择是否质疑`);
    },
    
    // 处理质疑结果
    handleChallengeResult: function(message) {
        let logText = '';
        
        if (message.wasChallenged) {
            if (message.challengeSuccess) {
                logText = `玩家 ${message.challengerName} 质疑成功！`;
            } else {
                logText = `玩家 ${message.challengerName} 质疑失败！`;
            }
            
            if (message.challengeReason) {
                logText += ` 理由: ${message.challengeReason}`;
            }
        } else {
            logText = `玩家 ${message.challengerName} 选择不质疑`;
            if (message.challengeReason) {
                logText += ` (${message.challengeReason})`;
            }
        }
        
        this.addLogEntry(logText);
    },
    
    // 处理射击结果
    handleShootingResult: function(message) {
        let logText = `玩家 ${message.shooterName} 开枪！`;
        
        if (message.bulletHit) {
            logText += ` 子弹命中，${message.shooterName} 已死亡！`;
        } else {
            logText += ` 子弹未命中，${message.shooterName} 幸存。`;
        }
        
        this.addLogEntry(logText);
    },
    
    // 处理系统自动质疑
    handleSystemChallenge: function(message) {
        let logText = `系统自动质疑 ${message.playerName} 的手牌！`