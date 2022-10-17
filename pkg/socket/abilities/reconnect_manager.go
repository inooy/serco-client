package abilities

import (
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/socket/model"
	"time"
)

const (
	DefaultDelay = 1000
)

type ReconnectManager struct {
	timer        *time.Timer
	socketClient model.SocketClient
	// 是否进行断线重连等管理
	isConnectionHolden bool
	// 延时连接时间
	reconnectTimeDelay int
	// 连接失败次数,不包括断开异常
	connectionFailedTimes int
	totalReconnectTimes   int
	lastTime              int
	running               bool
}

func NewReconnectManager(socketClient model.SocketClient) *ReconnectManager {
	return &ReconnectManager{
		socketClient: socketClient,
	}
}

func (r *ReconnectManager) reset() {
	if r.timer != nil {
		r.timer.Stop()
	}
	r.reconnectTimeDelay = DefaultDelay
	r.connectionFailedTimes = 0
}

func (r *ReconnectManager) handleReconnect() {
	if !r.running {
		r.reset()
		return
	}
	if !r.isConnectionHolden {
		r.reset()
		return
	}
	if !r.socketClient.IsConnect() {
		r.totalReconnectTimes++
		r.lastTime = time.Now().Nanosecond()
		err := r.socketClient.Connect()
		if err != nil {
			log.Error("reconnect fail", err.Error())
			return
		}
	}
}

func (r *ReconnectManager) reconnectDelay() {
	if r.timer != nil {
		r.timer.Stop()
	}
	r.timer = time.AfterFunc(time.Duration(r.reconnectTimeDelay)*time.Millisecond, func() {
		r.handleReconnect()
	})
	log.Info("Reconnect after mills ", r.reconnectTimeDelay)
	// 5+10+20+40 = 75 4次   1 + 2*5 11 21 31 1 10 20 40 80
	r.reconnectTimeDelay = r.reconnectTimeDelay + 5*2*1000
	if r.reconnectTimeDelay >= DefaultDelay+5*10*1000 {
		// DEFAULT * 10 = 50  10
		r.reconnectTimeDelay = DefaultDelay
	}
}

func (r *ReconnectManager) OnSocketDisconnection(err error) {
	if !r.running {
		return
	}
	if err == nil {
		r.reset()
		return
	}
	r.reconnectDelay()
}

func (r *ReconnectManager) OnSocketConnectionSuccess() {
	r.reset()
}

func (r *ReconnectManager) Start() {
	r.running = true
	r.reset()
}

func (r *ReconnectManager) Shutdown() {
	r.running = false
	r.reset()
}
