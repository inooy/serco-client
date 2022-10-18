package abilities

import (
	"github.com/inooy/serco-client/pkg/log"
	"github.com/inooy/serco-client/pkg/socket/model"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type HeartbeatOpts struct {

	/**
	 * 心跳间隔频率，单位：毫秒
	 */
	frequency int //30000L

	/**
	 * 心跳丢失次数<br>
	 * 大于或等于丢失次数时将断开该通道的连接<br>
	 * 抛出{@see DogDeadException}<br>
	 * 默认是5次 丢失心跳ACK的次数,例如5,当丢失3次时,自动断开.
	 */
	feedLoseTimes int // 3L
}

type HeartbeatType int

const (
	SEND HeartbeatType = 1
	FEED HeartbeatType = 2
)

type HeartbeatManager struct {
	loseTimes        int
	currentFrequency int
	totalPulseTimes  int
	totalFeedTimes   int
	isDead           bool
	timer            *time.Timer
	options          HeartbeatOpts
	socketClient     model.SocketClient
}

func NewHeartbeatManager(socketClient model.SocketClient) *HeartbeatManager {
	return &HeartbeatManager{
		socketClient: socketClient,
		loseTimes:    -1,
		options: HeartbeatOpts{
			frequency:     30000,
			feedLoseTimes: 3,
		},
	}
}

func (h *HeartbeatManager) dead() {
	if h.timer != nil {
		h.timer.Stop()
	}
}

func (h *HeartbeatManager) Dead() {
	log.Info("stop heartbeat")
	h.loseTimes = 0
	h.isDead = true
	h.dead()
}

// Pulse 启动心跳
func (h *HeartbeatManager) Pulse() {
	h.dead()
	if h.isDead {
		h.isDead = false
	}
	h.pulse()
}

func (h *HeartbeatManager) pulse() {
	h.dead()
	if h.isDead {
		return
	}
	h.currentFrequency = h.options.frequency
	if h.currentFrequency < 1000 {
		h.currentFrequency = 1000
	}
	h.timer = time.AfterFunc(time.Duration(h.currentFrequency)*time.Millisecond, func() {
		h.pulseHandler(SEND)
	})
}

// Feed 收到socket心跳回应，就调用喂狗
func (h *HeartbeatManager) Feed() {
	h.pulseHandler(FEED)
}

func (h *HeartbeatManager) pulseHandler(pulseType HeartbeatType) {
	if h.isDead {
		return
	}
	if pulseType == SEND {
		h.loseTimes++
		if h.options.feedLoseTimes != -1 && h.loseTimes >= h.options.feedLoseTimes {
			log.Warn("lose heartbeat..")
			_ = h.socketClient.Close(errors.WithMessage(model.DogDeadErr, strconv.Itoa(h.loseTimes)))
		} else {
			h.totalPulseTimes++
			err := h.socketClient.SendHeartbeat()
			if err != nil {
				log.Warn("send heartbeat fail!")
				_ = h.socketClient.Close(errors.WithMessage(model.DogDeadErr, strconv.Itoa(h.loseTimes)))
				return
			}
			h.pulse()
		}
	} else {
		h.totalFeedTimes++
		h.loseTimes = -1
	}
}
