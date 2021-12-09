package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"qubert/uuid"
)

type wsClient struct {
	token   string
	srcAddr string
	conn    *websocket.Conn

	mx *sync.Mutex

	module string
	args   []string
}

func (c *wsClient) setLocation(module string, args []string) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.module = module
	c.args = args
}

type Session struct {
	id        uuid.UUID
	token     string
	userName  string
	ValidFrom time.Time
	ValidTo   time.Time
	ctx       context.Context
	cancel    func()

	wsClientsMx sync.Mutex
	wsClients   []*wsClient
}

func (s *Session) newClient(token string, addr string, conn *websocket.Conn) *wsClient {
	s.wsClientsMx.Lock()
	defer s.wsClientsMx.Unlock()

	client := &wsClient{
		token:   token,
		srcAddr: addr,
		conn:    conn,
		mx:      &s.wsClientsMx,
	}

	s.wsClients = append(s.wsClients, client)

	return client
}

func (s *Session) removeClient(conn *websocket.Conn) {
	s.wsClientsMx.Lock()
	defer s.wsClientsMx.Unlock()

	for i, c := range s.wsClients {
		if c.conn == conn {
			s.wsClients = append(s.wsClients[:i], s.wsClients[i+1:]...)
		}
	}
}

type sessionManager struct {
	ctx           context.Context
	mx            sync.Mutex
	tokenLifetime time.Duration
	sessions      map[string]*Session
}

func newSessionManager(ctx context.Context) *sessionManager {
	return &sessionManager{
		ctx:           ctx,
		sessions:      make(map[string]*Session),
		tokenLifetime: 3 * time.Hour,
	}
}

func generateToken() string {
	token := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, token); err != nil {
		panic(err.Error())
	}

	result := make([]byte, 64)
	hex.Encode(result, token)

	return string(result)
}

func (sm *sessionManager) sessionByToken(token string) *Session {
	sm.mx.Lock()
	defer sm.mx.Unlock()

	if sn, ok := sm.sessions[token]; ok {
		now := time.Now()

		if sn.ValidFrom.Before(now) && sn.ValidTo.After(now) {
			return sn
		}

		delete(sm.sessions, token)

		return nil
	}

	return nil
}

func (sm *sessionManager) newSession(userName string) *Session {
	now := time.Now()
	validTo := now.Add(sm.tokenLifetime)

	sessionCTX, cancelFunc := context.WithDeadline(sm.ctx, validTo)

	sn := &Session{
		id:        uuid.New(),
		token:     generateToken(),
		userName:  userName,
		ValidFrom: now,
		ValidTo:   validTo,
		ctx:       sessionCTX,
		cancel:    cancelFunc,
	}

	sm.mx.Lock()
	defer sm.mx.Unlock()

	sm.sessions[sn.token] = sn

	return sn
}

func (sm *sessionManager) send(data interface{}, module string, args []string) bool {
	sm.mx.Lock()
	defer sm.mx.Unlock()

	var (
		wasSent bool
		wg      sync.WaitGroup
	)

	for _, s := range sm.sessions {
		for _, c := range s.wsClients {
			if module == "" || module == c.module {
				wg.Add(1)
				go func(conn *websocket.Conn, wg *sync.WaitGroup) {
					defer wg.Done()
					_ = c.conn.WriteJSON(data)
				}(c.conn, &wg)

				wasSent = true
			}
		}
	}

	wg.Wait()

	return wasSent
}
