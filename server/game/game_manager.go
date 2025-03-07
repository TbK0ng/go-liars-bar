package game

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// GameManager 管理所有游戏实例
type GameManager struct {
	games      map[string]*Game
	gamesMutex sync.RWMutex
}

// NewGameManager 创建新的游戏管理器
func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*Game),
	}
}

// CreateGame 创建新游戏
func (gm *GameManager) CreateGame() string {
	gameID := uuid.New().String()

	gm.gamesMutex.Lock()
	gm.games[gameID] = NewGame(gameID)
	gm.gamesMutex.Unlock()

	return gameID
}

// GetGame 获取游戏实例
func (gm *GameManager) GetGame(gameID string) *Game {
	gm.gamesMutex.RLock()
	defer gm.gamesMutex.RUnlock()

	return gm.games[gameID]
}

// RemoveGame 移除游戏
func (gm *GameManager) RemoveGame(gameID string) {
	gm.gamesMutex.Lock()
	delete(gm.games, gameID)
	gm.gamesMutex.Unlock()
}

// handleCreateGame 处理创建游戏的HTTP请求
func (gm *GameManager) HandleCreateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 创建新游戏
	gameID := gm.CreateGame()

	// 返回游戏ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"gameId": gameID,
	})
}

// handleJoinGame 处理加入游戏的HTTP请求
func (gm *GameManager) HandleJoinGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 解析请求
	var request struct {
		GameID     string `json:"gameId"`
		PlayerName string `json:"playerName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	// 获取游戏
	game := gm.GetGame(request.GameID)
	if game == nil {
		http.Error(w, "游戏不存在", http.StatusNotFound)
		return
	}

	// 检查游戏是否已满
	if game.IsFull() {
		http.Error(w, "游戏已满", http.StatusBadRequest)
		return
	}

	// 添加玩家
	playerID := game.AddPlayer(request.PlayerName)

	// 返回玩家ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"playerId": playerID,
	})
}
