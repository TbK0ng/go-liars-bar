package game

import (
	"log"
	"math/rand"
	"time"
)

// startGame 开始游戏
func (g *Game) startGame() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// 更新游戏状态
	g.State = GameStatePlaying
	g.GameOver = false
	g.RoundCount = 0

	// 随机选择起始玩家
	g.CurrentPlayerIdx = rand.Intn(len(g.PlayerOrder))

	// 发牌并选择目标牌
	g.dealCards()
	g.chooseTargetCard()

	// 开始第一轮
	g.startRound()

	// 广播游戏状态
	g.broadcastGameState()

	// 通知当前玩家轮到他出牌
	g.notifyCurrentPlayer()
}

// dealCards 发牌
func (g *Game) dealCards() {
	// 创建并洗牌
	g.Deck = g.createDeck()

	// 清空所有玩家手牌
	for _, player := range g.Players {
		if player.Alive {
			player.Hand = make([]string, 0)
		}
	}

	// 每位玩家发5张牌
	for i := 0; i < 5; i++ {
		for _, playerID := range g.PlayerOrder {
			player := g.Players[playerID]
			if player.Alive && len(g.Deck) > 0 {
				// 从牌堆顶部抽一张牌
				card := g.Deck[0]
				g.Deck = g.Deck[1:]
				player.Hand = append(player.Hand, card)
			}
		}
	}
}

// createDeck 创建并洗牌牌组
func (g *Game) createDeck() []string {
	deck := make([]string, 0)

	// 添加牌
	for i := 0; i < 6; i++ {
		deck = append(deck, CardQ)
	}
	for i := 0; i < 6; i++ {
		deck = append(deck, CardK)
	}
	for i := 0; i < 6; i++ {
		deck = append(deck, CardA)
	}
	for i := 0; i < 2; i++ {
		deck = append(deck, CardJoker)
	}

	// 洗牌
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	return deck
}

// chooseTargetCard 随机选择目标牌
func (g *Game) chooseTargetCard() {
	targetCards := []string{CardQ, CardK, CardA}
	g.TargetCard = targetCards[rand.Intn(len(targetCards))]
	log.Printf("目标牌是: %s", g.TargetCard)
}

// startRound 开始新的回合
func (g *Game) startRound() {
	g.RoundCount++
	log.Printf("开始第 %d 轮", g.RoundCount)

	// 获取当前玩家
	currentPlayerID := g.PlayerOrder[g.CurrentPlayerIdx]
	currentPlayer := g.Players[currentPlayerID]

	// 记录回合开始信息
	log.Printf("从玩家 %s 开始", currentPlayer.Name)
}

// getCurrentPlayerID 获取当前玩家ID
func (g *Game) getCurrentPlayerID() string {
	if g.CurrentPlayerIdx >= 0 && g.CurrentPlayerIdx < len(g.PlayerOrder) {
		return g.PlayerOrder[g.CurrentPlayerIdx]
	}
	return ""
}

// notifyCurrentPlayer 通知当前玩家轮到他出牌
func (g *Game) notifyCurrentPlayer() {
	currentPlayerID := g.getCurrentPlayerID()
	if currentPlayerID == "" {
		return
	}

	conn, ok := g.Connections[currentPlayerID]
	if !ok {
		return
	}

	// 发送通知
	conn.WriteJSON(map[string]interface{}{
		"type":    "your_turn",
		"message": "轮到你出牌了",
	})
}

// handlePlayCards 处理玩家出牌
func (g *Game) handlePlayCards(playerID string, cards []string) {
	// 验证是否是当前玩家
	if g.getCurrentPlayerID() != playerID {
		return
	}

	// 获取玩家
	player := g.Players[playerID]

	// 验证玩家手牌中是否有这些牌
	for _, card := range cards {
		found := false
		for i, handCard := range player.Hand {
			if handCard == card {
				// 从手牌中移除
				player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			// 玩家没有这张牌，发送错误消息
			conn, ok := g.Connections[playerID]
			if ok {
				conn.WriteJSON(map[string]interface{}{
					"type":    "error",
					"message": "你没有这张牌",
				})
			}
			return
		}
	}
}
