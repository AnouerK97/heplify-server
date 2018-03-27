package input

import (
	"bytes"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/database"
	"github.com/negbie/heplify-server/logp"
	"github.com/negbie/heplify-server/metric"
	"github.com/negbie/heplify-server/queue"
)

type HEPInput struct {
	addr    string
	pool    chan chan struct{}
	stats   HEPStats
	stop    bool
	workers int
}

type HEPStats struct {
	DupCount uint64
	ErrCount uint64
	HEPCount uint64
	PktCount uint64
}

var (
	inCh  = make(chan []byte, 10000)
	dbCh  = make(chan *decoder.HEP, 10000)
	mqCh  = make(chan []byte, 10000)
	mCh   = make(chan *decoder.HEP, 10000)
	dbCnt int
	mqCnt int
	mCnt  int

	hepBuffer = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 8192)
		},
	}
)

func NewHEP() *HEPInput {
	return &HEPInput{
		addr:    config.Setting.HEPAddr,
		workers: config.Setting.HEPWorkers,
		pool:    make(chan chan struct{}, runtime.NumCPU()*1e4),
	}
}

func (h *HEPInput) Run() {
	udpAddr, err := net.ResolveUDPAddr("udp", h.addr)
	if err != nil {
		logp.Critical("%v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logp.Critical("%v", err)
	}

	for n := 0; n < h.workers; n++ {
		go func() {
			shut := make(chan struct{})
			h.pool <- shut
			h.hepWorker(shut)
		}()
	}

	if config.Setting.DBAddr != "" {
		go func() {
			d := database.New(config.Setting.DBDriver)
			d.Chan = dbCh

			if err := d.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if config.Setting.MQAddr != "" && config.Setting.MQName != "" {
		go func() {
			q := queue.New(config.Setting.MQName)
			q.Topic = config.Setting.MQTopic
			q.Chan = mqCh

			if err := q.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	if config.Setting.PromAddr != "" {
		go func() {
			m := metric.New("prometheus")
			m.Chan = mCh

			if err := m.Run(); err != nil {
				logp.Err("%v", err)
			}
		}()
	}

	logp.Info("hep input address: %s, workders: %d\n", h.addr, h.workers)
	go h.logStats()

	for !h.stop {
		buf := hepBuffer.Get().([]byte)
		conn.SetReadDeadline(time.Now().Add(1e9))
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			continue
		} else if n > 8192 {
			logp.Warn("received to big packet with %d bytes", n)
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}
		atomic.AddUint64(&h.stats.PktCount, 1)
		inCh <- buf[:n]
	}
}

func (h *HEPInput) End() {
	h.stop = true
	time.Sleep(2 * time.Second)
	logp.Info("heplify-server has been stopped")
	close(inCh)
}

func (h *HEPInput) hepWorker(shut chan struct{}) {
	var (
		hepPkt *decoder.HEP
		msg    = hepBuffer.Get().([]byte)
		buf    = new(bytes.Buffer)
		err    error
		ok     bool
	)

GO:
	for {

		hepBuffer.Put(msg[:8192])
		buf.Reset()

		select {
		case <-shut:
			break GO
		case msg, ok = <-inCh:
			if !ok {
				break GO
			}
		}

		hepPkt, err = decoder.DecodeHEP(msg)
		if err != nil || hepPkt == nil {
			atomic.AddUint64(&h.stats.ErrCount, 1)
			continue
		}

		atomic.AddUint64(&h.stats.HEPCount, 1)

		if config.Setting.DBAddr != "" {
			select {
			case dbCh <- hepPkt:
			default:
				dbCnt++
				if dbCnt%128 == 0 {
					dbCnt = 0
					logp.Warn("overflowing db channel by 128 packets")
				}
			}
		}

		if config.Setting.PromAddr != "" {
			select {
			case mCh <- hepPkt:
			default:
				mCnt++
				if mCnt%128 == 0 {
					mCnt = 0
					logp.Warn("overflowing metric channel by 128 packets")
				}
			}
		}

		if config.Setting.MQAddr != "" {
			select {
			case mqCh <- msg:
			default:
				mqCnt++
				if mqCnt%128 == 0 {
					mqCnt = 0
					logp.Warn("overflowing queue channel by 128 packets")
				}
			}
		}
	}
}

func (h *HEPInput) logStats() {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ticker.C:
			logp.Info("Packet stats since last minute Total: %d, HEP: %d, Duplicate: %d, Error: %d",
				atomic.LoadUint64(&h.stats.PktCount),
				atomic.LoadUint64(&h.stats.HEPCount),
				atomic.LoadUint64(&h.stats.DupCount),
				atomic.LoadUint64(&h.stats.ErrCount),
			)
			atomic.StoreUint64(&h.stats.PktCount, 0)
			atomic.StoreUint64(&h.stats.HEPCount, 0)
			atomic.StoreUint64(&h.stats.DupCount, 0)
			atomic.StoreUint64(&h.stats.ErrCount, 0)

			logp.Info("Current channel queue length Input: %d, Database: %d, Metric: %d",
				len(inCh),
				len(dbCh),
				len(mCh),
			)
		}
	}
}
