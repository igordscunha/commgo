package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"github.com/igordscunha/commgo/mypkg"
)

const (
	serviceName = "_meuchat._tcp"
	domain      = "local."
)

type Peer struct {
	conn *websocket.Conn
}

var (
	peers      = make(map[string]*Peer)
	peersMutex = &sync.RWMutex{}

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	localUsername string
	teaProgram    *tea.Program
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro ao fazer upgrade para WebSocket: %v", err)
		return
	}

	defer conn.Close()
	log.Printf("Peer conectado de: %s", conn.RemoteAddr())

	addr := conn.RemoteAddr().String()
	defer func() {
		peersMutex.Lock()
		delete(peers, addr)
		peersMutex.Unlock()
		log.Printf("Conexão com %s fechada e removida.", addr)
	}()

	for {
		_, mensagemBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Peer desconectado: %s, Erro: %v", conn.RemoteAddr(), err)
			break
		}

		var msg mypkg.Mensagem
		if err := json.Unmarshal(mensagemBytes, &msg); err != nil {
			log.Printf("Erro ao decodificar mensagem JSON: %v", err)
			continue
		}

		if teaProgram != nil {
			teaProgram.Send(mypkg.IncomingMsg(msg))
		}
	}
}

func broadcastMensagem(texto string) {
	peersMutex.RLock()
	defer peersMutex.RUnlock()

	msg := mypkg.Mensagem{
		ID:        uuid.New().String(),
		Username:  localUsername,
		Texto:     texto,
		Timestamp: time.Now(),
	}

	mensagemBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Erro ao codificar mensagem para JSON: %v", err)
		return
	}

	for addr, peer := range peers {
		if err := peer.conn.WriteMessage(websocket.TextMessage, mensagemBytes); err != nil {
			log.Printf("Erro ao enviar mensagem para %s: %v. Removendo peer.", addr, err)
			peer.conn.Close()
		}
	}
}

func discoverPeers(ctx context.Context) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalf("Falha ao inicializar o resolver mDNS: %v", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(resultados <-chan *zeroconf.ServiceEntry) {
		for entry := range resultados {
			if entry.Instance == localUsername {
				continue
			}

			peerAddr := fmt.Sprintf("%s:%d", entry.AddrIPv4[0], entry.Port)

			peersMutex.RLock()
			_, existe := peers[peerAddr]
			peersMutex.RUnlock()
			if existe {
				continue
			}

			log.Printf("Peer descoberto: %s em %s", entry.Instance, peerAddr)
			wsURL := "ws://" + peerAddr + "/ws"

			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				log.Printf("Falha ao conectar ao peer %s: %v", peerAddr, err)
				continue
			}

			log.Printf("Conexão WebSocket estabelecida com %s", peerAddr)
			novoPeer := &Peer{conn: conn}
			peersMutex.Lock()
			peers[conn.RemoteAddr().String()] = novoPeer
			peersMutex.Unlock()
		}
	}(entries)

	log.Println("Buscando por outros chats na rede...")
	if err := resolver.Browse(ctx, serviceName, domain, entries); err != nil {
		log.Fatalf("Falha ao buscar serviços mDNS: %v", err)
	}
	<-ctx.Done()
}

// ##### MAIN #####

func main() {

	logFile, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer logFile.Close()

	log.Println("Chat iniciado...")
	fmt.Print("Digite seu nome de usuário: ")
	scanner := bufio.NewScanner(os.Stdin)

	if scanner.Scan() {
		localUsername = scanner.Text()
	}
	if localUsername == "" {
		log.Fatal("Nome do usuário deve conter algum valor.")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Falha ao iniciar listener: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	go func() {
		if err := http.Serve(listener, mux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Servidor HTTP falhou: %v", err)
		}
	}()

	log.Printf("Servidor iniciado. Escutando em http://localhost:%d", port)

	server, err := zeroconf.Register(localUsername, serviceName, domain, port, []string{"txtv=0", "lo=1", "la=2"}, nil)
	if err != nil {
		log.Fatalf("Falha ao registrar serviço mDNS: %v", err)
	}
	defer server.Shutdown()
	log.Printf("Serviço '%s' anunciado na porta %d", localUsername, port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go discoverPeers(ctx)

	initialModel := mypkg.InitialModel(localUsername, broadcastMensagem)
	teaProgram = tea.NewProgram(initialModel, tea.WithAltScreen())

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Encerrando a aplicação...")
		cancel()
		server.Shutdown()
		teaProgram.Quit()
	}()

	if _, err := teaProgram.Run(); err != nil {
		log.Fatalf("Erro ao executar o design Bubble Tea: %v", err)
	}
}
