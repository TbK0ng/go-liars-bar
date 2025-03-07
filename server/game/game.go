package game

import (
	_ "encoding/json"
	_ "fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// 游戏状态常量
const (
	GameStateWaiting  = "waiting"  // 等待玩家加入
	GameStateStarting = "starting" // 游戏开始中
	GameStatePlaying  = "playing"  // 游戏进行中
	GameStateFinished = "finished" // 游戏已结束
)

// 卡牌类型
const (
	CardQ     = "Q"
	CardK     = "K"
	CardA     = "A"
	CardJoker = "Joker"
)

// Game 表示一个游戏实例
type Game struct {
	ID               string                     `json:"id"`
	State            string                     `json:"state"`
	Players          map[string]*Player         `json:"players"`
	PlayerOrder      []string                   `json:"playerOrder"` // 玩家顺序
	Connections      map[string]*websocket.Conn `json:"connections,omitempty"`
	Deck             []string                   `json:"deck"`
	TargetCard       string                     `json:"targetCard"`
	CurrentPlayerIdx int                        `json:"currentPlayerIdx"`
	LastShooterID    string                     `json:"lastShooterId"`
	RoundCount       int                        `json:"roundCount"`
	GameOver         bool                       `json:"gameOver"`
	mutex            sync.RWMutex
}

// Player 表示一个玩家
type Player struct {
	ID                    string            `json:"id"`
	Name                  string            `json:"name"`
	Hand                  []string          `json:"hand,omitempty"` // 对其他玩家隐藏
	Alive                 bool              `json:"alive"`
	BulletPosition        int               `json:"bulletPosition,omitempty"`        // 对其他玩家隐藏
	CurrentBulletPosition int               `json:"currentBulletPosition,omitempty"` // 对其他玩家隐藏
	Opinions              map[string]string `json:"opinions"`                        // 对其他玩家的看法
}

// PlayerInitialState 记录玩家初始状态
type PlayerInitialState struct {
	PlayerID           string   `json:"playerId"`
	PlayerName         string   `json:"playerName"`
	BulletPosition     int      `json:"bulletPosition"`
	CurrentGunPosition int      `json:"currentGunPosition"`
	InitialHand        []string `json:"initialHand"`
}

// PlayAction 记录一次出牌行为
type PlayAction struct {
	PlayerID        string   `json:"playerId"`
	PlayerName      string   `json:"playerName"`
	PlayedCards     []string `json:"playedCards"`
	RemainingCards  []string `json:"remainingCards"`
	PlayReason      string   `json:"playReason"`
	Behavior        string   `json:"behavior"`
	NextPlayerID    string   `json:"nextPlayerId"`
	NextPlayerName  string   `json:"nextPlayerName"`
	WasChallenged   bool     `json:"wasChallenged"`
	ChallengeReason string   `json:"challengeReason,omitempty"`
	ChallengeResult *bool    `json:"challengeResult,omitempty"`
}

// ShootingResult 记录一次开枪结果
type ShootingResult struct {
	ShooterID   string `json:"shooterId"`
	ShooterName string `json:"shooterName"`
	BulletHit   bool   `json:"bulletHit"`
}

// RoundRecord 记录一轮游戏
type RoundRecord struct {
	RoundID             int                          `json:"roundId"`
	TargetCard          string                       `json:"targetCard"`
	RoundPlayers        []string                     `json:"roundPlayers"`
	StartingPlayerID    string                       `json:"startingPlayerId"`
	StartingPlayerName  string                       `json:"startingPlayerName"`
	PlayerInitialStates []PlayerInitialState         `json:"playerInitialStates"`
	PlayerOpinions      map[string]map[string]string `json:"playerOpinions"`
	PlayHistory         []PlayAction                 `json:"playHistory"`
	RoundResult         *ShootingResult              `json:"roundResult,omitempty"`
}

// GameRecord 完整游戏记录
type GameRecord struct {
	GameID      string        `json:"gameId"`
	PlayerNames []string      `json:"playerNames"`
	Rounds      []RoundRecord `json:"rounds"`
	Winner      string        `json:"winner,omitempty"`
}

// NewGame 创建一个新的游戏实例
func NewGame(id string) *Game {
	rand.Seed(time.Now().UnixNano())
	return &Game{
		ID:          id,
		State:       GameStateWaiting,
		Players:     make(map[string]*Player),
		PlayerOrder: make([]string, 0),
		Connections: make(map[string]*websocket.Conn),
		Deck:        make([]string, 0),
		GameOver:    false,
		RoundCount:  0,
	}
}

// IsFull 检查游戏是否已满（最多4名玩家）
func (g *Game) IsFull() bool {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return len(g.Players) >= 4
}

// AddPlayer 添加一个新玩家到游戏
func (g *Game) AddPlayer(name string) string {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// 生成玩家ID
	playerID := uuid.New().String()

	// 创建玩家
	player := &Player{
		ID:                    playerID,
		Name:                  name,
		Hand:                  make([]string, 0),
		Alive:                 true,
		BulletPosition:        rand.Intn(6),
		CurrentBulletPosition: 0,
		Opinions:              make(map[string]string),
	}

	// 添加到游戏
	g.Players[playerID] = player
	g.PlayerOrder = append(g.PlayerOrder, playerID)

	// 广播玩家加入消息
	g.broadcastGameState()

	return playerID
}

// ConnectPlayer 将玩家的WebSocket连接添加到游戏
func (g *Game) ConnectPlayer(playerID string, conn *websocket.Conn) {
	g.mutex.Lock()

	// 添加连接
	g.Connections[playerID] = conn

	// 设置关闭处理函数
	conn.SetCloseHandler(func(code int, text string) error {
		g.mutex.Lock()
		delete(g.Connections, playerID)
		g.mutex.Unlock()
		return nil
	})

	// 发送当前游戏状态给新连接的玩家
	g.sendGameStateToPlayer(playerID)

	g.mutex.Unlock()

	// 启动消息处理循环
	go g.handlePlayerMessages(playerID, conn)
}

// handlePlayerMessages 处理来自玩家的WebSocket消息
func (g *Game) handlePlayerMessages(playerID string, conn *websocket.Conn) {
	for {
		// 读取消息
		var message map[string]interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Printf("读取消息错误: %v", err)
			break
		}

		// 处理消息
		g.handleMessage(playerID, message)
	}
}

// handleMessage 处理玩家发送的消息
func (g *Game) handleMessage(playerID string, message map[string]interface{}) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// 获取消息类型
	msgType, ok := message["type"].(string)
	if !ok {
		log.Printf("无效的消息格式: 缺少type字段")
		return
	}

	// 根据消息类型处理
	switch msgType {
	case "start_game":
		// 只有当游戏处于等待状态且有足够的玩家时才能开始
		if g.State == GameStateWaiting && len(g.Players) >= 2 {
			g.startGame()
		}

	case "play_cards":
		// 处理出牌
		if g.State == GameStatePlaying && g.getCurrentPlayerID() == playerID {
			cardsData, ok := message["cards"].([]interface{})
			if !ok {
				log.Printf("无效的出牌格式")
				return
			}

			// 转换卡牌数据
			cards := make([]string, len(cardsData))
			for i, card := range cardsData {
				cards[i] = card.(string)
			}

			// 处理出牌逻辑
			g.handlePlayCards(playerID, cards)
		}

	case "challenge":
		// 处理质疑
		if g.State == GameStatePlaying {
			challenge, ok := message["challenge"].(bool)
			if !ok {
				log.Printf("无效的质疑格式")
				return
			}

			reason, _ := message["reason"].(string)
			g.handleChallenge(playerID, challenge, reason)
		}
	}
}

// broadcastGameState 向所有连接的玩家广播游戏状态
func (g *Game) broadcastGameState() {
	for playerID := range g.Connections {
		g.sendGameStateToPlayer(playerID)
	}
}

// sendGameStateToPlayer 向特定玩家发送游戏状态
func (g *Game) sendGameStateToPlayer(playerID string) {
	conn, ok := g.Connections[playerID]
	if !ok {
		return
	}

	// 创建针对该玩家的游戏状态视图
	gameState := g.createGameStateForPlayer(playerID)

	// 发送游戏状态
	conn.WriteJSON(map[string]interface{}{
		"type":  "game_state",
		"state": gameState,
	})
}

// createGameStateForPlayer 创建针对特定玩家的游戏状态视图
func (g *Game) createGameStateForPlayer(playerID string) map[string]interface{} {
	// 基本游戏信息
	gameState := map[string]interface{}{
		"id":               g.ID,
		"state":            g.State,
		"currentPlayerIdx": g.CurrentPlayerIdx,
		"roundCount":       g.RoundCount,
		"gameOver":         g.GameOver,
		"targetCard":       g.TargetCard,
	}

	// 玩家信息（隐藏其他玩家的手牌和子弹位置）
	players := make(map[string]interface{})
	for id, player := range g.Players {
		playerView := map[string]interface{}{
			"id":    player.ID,
			"name":  player.Name,
			"alive": player.Alive,
		}

		// 只向当前玩家展示自己的手牌和子弹位置
		if id == playerID {
			playerView["hand"] = player.Hand
			playerView["bulletPosition"] = player.BulletPosition
			playerView["currentBulletPosition"] = player.CurrentBulletPosition
		} else {
			// 对其他玩家只显示手牌数量
			playerView["handCount"] = len(player.Hand)
		}

		// 添加到玩家列表
		players[id] = playerView
	}

	// 添加玩家信息到游戏状态
	gameState["players"] = players

	// 添加玩家顺序
	gameState["playerOrder"] = g.PlayerOrder

	// 如果游戏已结束，添加胜利者信息
	if g.GameOver {
		for _, player := range g.Players {
			if player.Alive {
				gameState["winner"] = player.Name
				break
			}
		}
	}

	return gameState
}

// broadcastPlayAction 广播出牌行为
func (g *Game) broadcastPlayAction(action PlayAction) {
	// 创建广播消息
	message := map[string]interface{}{
		"type":       "play_action",
		"playerName": action.PlayerName,
		"cardCount":  len(action.PlayedCards),
		"targetCard": g.TargetCard,
		"nextPlayer": action.NextPlayerName,
	}

	// 广播给所有玩家
	for playerID, conn := range g.Connections {
		// 对出牌玩家显示实际打出的牌
		if playerID == action.PlayerID {
			message["playedCards"] = action.PlayedCards
		}
		conn.WriteJSON(message)
	}
}

// waitForChallenge 等待下一个玩家决定是否质疑
func (g *Game) waitForChallenge(nextPlayerID string, playAction PlayAction) {
	// 通知下一个玩家需要决定是否质疑
	conn, ok := g.Connections[nextPlayerID]
	if !ok {
		return
	}

	// 发送质疑请求
	conn.WriteJSON(map[string]interface{}{
		"type":       "challenge_request",
		"playerName": playAction.PlayerName,
		"cardCount":  len(playAction.PlayedCards),
		"targetCard": g.TargetCard,
	})
}

// handleChallenge 处理玩家质疑
func (g *Game) handleChallenge(playerID string, challenge bool, reason string) {
	// 获取当前玩家和上一个玩家
	currentPlayerIdx := -1
	for i, id := range g.PlayerOrder {
		if id == playerID {
			currentPlayerIdx = i
			break
		}
	}

	if currentPlayerIdx == -1 {
		return
	}

	// 找到上一个玩家（出牌者）
	prevPlayerIdx := (currentPlayerIdx - 1 + len(g.PlayerOrder)) % len(g.PlayerOrder)
	for !g.Players[g.PlayerOrder[prevPlayerIdx]].Alive || len(g.Players[g.PlayerOrder[prevPlayerIdx]].Hand) == 0 {
		prevPlayerIdx = (prevPlayerIdx - 1 + len(g.PlayerOrder)) % len(g.PlayerOrder)
		if prevPlayerIdx == currentPlayerIdx {
			return // 没有找到有效的上一个玩家
		}
	}

	prevPlayerID := g.PlayerOrder[prevPlayerIdx]
	//prevPlayer := g.Players[prevPlayerID]

	// 如果玩家选择不质疑
	if !challenge {
		// 广播不质疑的消息
		g.broadcastChallengeResult(playerID, false, false, reason)

		// 切换到下一个玩家
		g.CurrentPlayerIdx = currentPlayerIdx
		g.moveToNextPlayer()
		return
	}

	// 玩家选择质疑，验证上一个玩家的出牌是否合法
	// 这里需要获取上一次出牌的信息
	// 简化处理，假设我们能从某处获取到上一次出牌信息
	lastPlayedCards := make([]string, 0) // 这里应该从游戏记录中获取

	// 验证出牌是否合法
	isValid := g.isValidPlay(lastPlayedCards)

	// 广播质疑结果
	g.broadcastChallengeResult(playerID, true, isValid, reason)

	// 根据质疑结果确定受罚玩家
	penaltyPlayerID := ""
	if isValid {
		// 质疑失败，质疑者受罚
		penaltyPlayerID = playerID
	} else {
		// 质疑成功，出牌者受罚
		penaltyPlayerID = prevPlayerID
	}

	// 记录最后射击者
	g.LastShooterID = penaltyPlayerID

	// 执行惩罚
	g.performPenalty(penaltyPlayerID)
}

// isValidPlay 判断出牌是否符合规则
func (g *Game) isValidPlay(cards []string) bool {
	for _, card := range cards {
		if card != g.TargetCard && card != CardJoker {
			return false
		}
	}
	return true
}

// broadcastChallengeResult 广播质疑结果
func (g *Game) broadcastChallengeResult(challengerID string, challenged bool, challengeSuccess bool, reason string) {
	challenger := g.Players[challengerID]

	// 创建广播消息
	message := map[string]interface{}{
		"type":            "challenge_result",
		"challengerName":  challenger.Name,
		"wasChallenged":   challenged,
		"challengeReason": reason,
	}

	if challenged {
		message["challengeSuccess"] = challengeSuccess
	}

	// 广播给所有玩家
	for _, conn := range g.Connections {
		conn.WriteJSON(message)
	}
}

// moveToNextPlayer 切换到下一个玩家
func (g *Game) moveToNextPlayer() {
	// 找到下一个有手牌的玩家
	g.CurrentPlayerIdx = g.findNextPlayerWithCards(g.CurrentPlayerIdx)

	// 广播游戏状态
	g.broadcastGameState()

	// 通知当前玩家轮到他出牌
	g.notifyCurrentPlayer()
}

// findNextPlayerWithCards 找到下一个有手牌的玩家
func (g *Game) findNextPlayerWithCards(startIdx int) int {
	idx := startIdx
	for i := 0; i < len(g.PlayerOrder); i++ {
		idx = (idx + 1) % len(g.PlayerOrder)
		playerID := g.PlayerOrder[idx]
		player := g.Players[playerID]
		if player.Alive && len(player.Hand) > 0 {
			return idx
		}
	}
	return startIdx // 如果没有其他玩家有手牌，返回当前玩家
}

// performPenalty 执行惩罚
func (g *Game) performPenalty(playerID string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// 获取玩家
	player := g.Players[playerID]
	if player == nil {
		return
	}

	// 执行射击
	log.Printf("玩家 %s 开枪！", player.Name)

	// 增加当前子弹位置
	player.CurrentBulletPosition = (player.CurrentBulletPosition + 1) % 6

	// 检查是否命中
	bulletHit := player.CurrentBulletPosition == player.BulletPosition

	// 记录射击结果
	shootingResult := ShootingResult{
		ShooterID:   playerID,
		ShooterName: player.Name,
		BulletHit:   bulletHit,
	}

	// 广播射击结果
	g.broadcastShootingResult(shootingResult)

	// 如果子弹命中，玩家死亡
	if bulletHit {
		player.Alive = false
		log.Printf("%s 已死亡！", player.Name)

		// 检查胜利条件
		if !g.checkVictory() {
			// 重置回合
			g.resetRound(true)
		}
	} else {
		// 子弹未命中，重置回合
		g.resetRound(true)
	}
}

// broadcastShootingResult 广播射击结果
func (g *Game) broadcastShootingResult(result ShootingResult) {
	// 创建广播消息
	message := map[string]interface{}{
		"type":        "shooting_result",
		"shooterName": result.ShooterName,
		"bulletHit":   result.BulletHit,
	}

	// 广播给所有玩家
	for _, conn := range g.Connections {
		conn.WriteJSON(message)
	}

	// 给玩家一些时间查看结果
	time.Sleep(2 * time.Second)
}

// resetRound 重置回合
func (g *Game) resetRound(recordShooter bool) {
	log.Println("小局游戏重置，开始新的一局！")

	// 重新发牌
	g.dealCards()
	g.chooseTargetCard()

	if recordShooter && g.LastShooterID != "" {
		// 从上一个射击者开始
		shooterIdx := -1
		for i, id := range g.PlayerOrder {
			if id == g.LastShooterID {
				shooterIdx = i
				break
			}
		}

		if shooterIdx >= 0 && g.Players[g.LastShooterID].Alive {
			g.CurrentPlayerIdx = shooterIdx
		} else {
			// 射击者已死亡，顺延至下一个存活且有手牌的玩家
			g.CurrentPlayerIdx = g.findNextPlayerWithCards(shooterIdx)
		}
	} else {
		// 随机选择一个存活的玩家
		alivePlayers := make([]string, 0)
		for _, id := range g.PlayerOrder {
			if g.Players[id].Alive {
				alivePlayers = append(alivePlayers, id)
			}
		}

		if len(alivePlayers) > 0 {
			randomIdx := rand.Intn(len(alivePlayers))
			playerID := alivePlayers[randomIdx]

			// 找到这个玩家在PlayerOrder中的索引
			for i, id := range g.PlayerOrder {
				if id == playerID {
					g.CurrentPlayerIdx = i
					break
				}
			}
		}
	}

	// 开始新回合
	g.startRound()

	// 广播游戏状态
	g.broadcastGameState()

	// 通知当前玩家轮到他出牌
	g.notifyCurrentPlayer()
}

// checkVictory 检查胜利条件
func (g *Game) checkVictory() bool {
	// 统计存活玩家
	alivePlayers := make([]string, 0)
	for _, id := range g.PlayerOrder {
		if g.Players[id].Alive {
			alivePlayers = append(alivePlayers, id)
		}
	}

	// 如果只剩一名玩家，游戏结束
	if len(alivePlayers) == 1 {
		winnerID := alivePlayers[0]
		winner := g.Players[winnerID]

		log.Printf("\n%s 获胜！", winner.Name)

		// 更新游戏状态
		g.State = GameStateFinished
		g.GameOver = true

		// 广播胜利消息
		g.broadcastVictory(winnerID)

		return true
	}

	return false
}

// broadcastVictory 广播胜利消息
func (g *Game) broadcastVictory(winnerID string) {
	winner := g.Players[winnerID]

	// 创建广播消息
	message := map[string]interface{}{
		"type":       "game_over",
		"winnerName": winner.Name,
	}

	// 广播给所有玩家
	for _, conn := range g.Connections {
		conn.WriteJSON(message)
	}
}

// handleSystemChallenge 处理系统自动质疑
func (g *Game) handleSystemChallenge(playerID string, cards []string) {
	log.Printf("系统自动质疑 %s 的手牌！", g.Players[playerID].Name)

	// 验证出牌是否合法
	isValid := g.isValidPlay(cards)

	// 广播系统质疑结果
	g.broadcastSystemChallengeResult(playerID, isValid, cards)

	if isValid {
		log.Printf("系统质疑失败！%s 的手牌符合规则。", g.Players[playerID].Name)
		// 重置回合
		g.resetRound(false)
	} else {
		log.Printf("系统质疑成功！%s 的手牌违规，将执行射击惩罚。", g.Players[playerID].Name)
		// 记录最后射击者
		g.LastShooterID = playerID
		// 执行惩罚
		g.performPenalty(playerID)
	}
}

// broadcastSystemChallengeResult 广播系统质疑结果
func (g *Game) broadcastSystemChallengeResult(playerID string, isValid bool, cards []string) {
	// 创建广播消息
	message := map[string]interface{}{
		"type":           "system_challenge",
		"playerName":     g.Players[playerID].Name,
		"challengeValid": !isValid, // 质疑成功意味着出牌不合法
		"playedCards":    cards,
	}

	// 广播给所有玩家
	for _, conn := range g.Connections {
		conn.WriteJSON(message)
	}

	// 给玩家一些时间查看结果
	time.Sleep(2 * time.Second)
}

// checkOtherPlayersNoCards 检查是否所有其他存活玩家都没有手牌
func (g *Game) checkOtherPlayersNoCards(playerID string) bool {
	for id, player := range g.Players {
		if id != playerID && player.Alive && len(player.Hand) > 0 {
			return false
		}
	}
	return true
}
