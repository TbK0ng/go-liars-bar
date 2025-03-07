package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"server/game"
	"syscall"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8080", "服务地址")

// 配置websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有CORS请求
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	flag.Parse()

	// 创建游戏管理器
	gameManager := game.NewGameManager()

	// 设置HTTP路由
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, gameManager)
	})

	// 设置API路由
	http.HandleFunc("/api/games", gameManager.HandleCreateGame)
	http.HandleFunc("/api/games/join", gameManager.HandleJoinGame)

	// 设置静态文件服务
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// 启动HTTP服务器
	server := &http.Server{Addr: *addr}

	// 优雅关闭
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		// 收到中断信号，关闭服务器
		log.Println("关闭服务器...")
		server.Close()
	}()

	// 启动服务器
	log.Printf("服务器启动在 %s", *addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 处理WebSocket连接
func handleWebSocket(w http.ResponseWriter, r *http.Request, gameManager *game.GameManager) {
	// 从查询参数获取游戏ID和玩家ID
	gameID := r.URL.Query().Get("gameId")
	playerID := r.URL.Query().Get("playerId")

	if gameID == "" || playerID == "" {
		http.Error(w, "缺少gameId或playerId参数", http.StatusBadRequest)
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	// 将玩家添加到游戏
	game := gameManager.GetGame(gameID)
	if game == nil {
		conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"message": "游戏不存在",
		})
		conn.Close()
		return
	}

	// 将玩家连接到游戏
	game.ConnectPlayer(playerID, conn)
}
