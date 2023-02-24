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
	timer                         *time.Timer
	socketClient                  model.SocketClient
	isConnectionHolden            bool // 是否进行断线重连等管理
	reconnectTimeDelay            int  // 延时连接时间
	reconnectTicket               int  // 重连等待时间刻度，毫秒
	reconnectIncrementCount       int  // 重连递增次数
	reconnectIncrementTicketCount int  // 重连每次递增刻度
	connectionFailedTimes         int  // 连接失败次数,不包括断开异常
	totalReconnectTimes           int
	lastTime                      int
	running                       bool
}

func NewReconnectManager(socketClient model.SocketClient) *ReconnectManager {
	return &ReconnectManager{
		socketClient:                  socketClient,
		isConnectionHolden:            true,
		reconnectTicket:               1000,
		reconnectIncrementCount:       5,
		reconnectIncrementTicketCount: 2,
	}
}

func (r *ReconnectManager) reset() {
	log.Info("重置重连")
	if r.timer != nil {
		r.timer.Stop()
	}
	r.reconnectTimeDelay = DefaultDelay
	r.connectionFailedTimes = 0
}

func (r *ReconnectManager) handleReconnect() {
	log.Info("handle reconnect")
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
			r.OnSocketDisconnection(err)
			return
		}
		log.Info("reconnect success!")
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
	r.reconnectTimeDelay = r.reconnectTimeDelay + r.reconnectIncrementTicketCount*r.reconnectTicket
	if r.reconnectTimeDelay >= DefaultDelay+r.reconnectIncrementCount*r.reconnectIncrementTicketCount*r.reconnectTicket {
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
