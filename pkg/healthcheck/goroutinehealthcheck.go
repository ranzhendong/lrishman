package healthcheck

import (
	"context"
	"github.com/ranzhendong/irishman/pkg/datastruck"
	"github.com/ranzhendong/irishman/pkg/kvnuts"
	"log"
	"time"
)

//healthCheck : template for goroutines
type healthCheck struct {
	CheckProtocol string   `json:"checkProtocol"`
	CheckPath     string   `json:"checkPath"`
	Health        health   `json:"health"`
	UnHealth      unHealth `json:"unhealth"`
}

type health struct {
	Interval       int   `json:"interval"`
	SuccessTime    int   `json:"successTime"`
	SuccessTimeout int   `json:"successTimeout"`
	SuccessStatus  []int `json:"successStatus"`
}

//template and put UnHealth
type unHealth struct {
	Interval        int   `json:"interval"`
	FailuresTime    int   `json:"failuresTime"`
	FailuresTimeout int   `json:"failuresTimeout"`
	FailuresStatus  []int `json:"failuresStatus"`
}

type ctxUpstreamList struct {
	upstreamList [][]byte
	ctx          context.Context
	cancel       context.CancelFunc
}

type ctxStart struct {
	upstreamList [][]byte
	ctx          context.Context
}

var (
	c                   datastruck.Config
	upstreamListChan    = make(chan [][]byte)
	ctxCancelChan       = make(chan context.CancelFunc)
	ctxStartChan        = make(chan ctxStart)
	ctxUpstreamListChan = make(chan ctxUpstreamList)
)

//HC : new health check
func HC() {

	// set bit, tell hc controller need to be updated
	//rootCtx = context.Background()
	//
	////Derive a context with cancel

	for {
		select {
		case Cancels := <-ctxCancelChan:
			Cancels()
		case Start := <-ctxStartChan:
			go upList(Start.upstreamList, Start.ctx)
		case cu := <-ctxUpstreamListChan:
			go upList(cu.upstreamList, cu.ctx)
		default:
			go FalgHC()
		}
	}
}

func FalgHC() {
	var (
		upstreamList [][]byte
		cul          ctxUpstreamList
		st           ctxStart
	)

	//first hc
	_ = kvnuts.Del("FalgHC", "FalgHC")
	upstreamList, _ = kvnuts.SMem(c.NutsDB.Tag.UpstreamList, c.NutsDB.Tag.UpstreamList)
	ctx, cancel := context.WithCancel(context.Background())
	cul.upstreamList = upstreamList
	cul.ctx = ctx
	cul.cancel = cancel
	ctxUpstreamListChan <- cul

	for {
		time.Sleep(1 * time.Second)
		upstreamList, _ = kvnuts.SMem(c.NutsDB.Tag.UpstreamList, c.NutsDB.Tag.UpstreamList)
		if _, _, err := kvnuts.Get("FalgHC", "FalgHC", "i"); err != nil {
			_ = kvnuts.Del("FalgHC", "FalgHC")
			ctxCancelChan <- cul.cancel
			st.ctx = cul.ctx
			st.upstreamList = upstreamList
			ctxStartChan <- st
		}
	}
}

func upList(upstreamList [][]byte, ctx context.Context) {
	for _, k := range upstreamList {
		log.Println("my string", string(k))
		//list has eight data, so index[0-7]
		log.Println(kvnuts.LIndex(string(k), k, 0, 7))
		if item, _ := kvnuts.LIndex(string(k), k, 0, 7); len(item) != 0 {
			hp := string(item[0])
			hps := string(item[1])
			hi, _ := kvnuts.BytesToInt(item[2], true)
			ht, _ := kvnuts.BytesToInt(item[3], true)
			hto, _ := kvnuts.BytesToInt(item[4], true)
			hfi, _ := kvnuts.BytesToInt(item[5], true)
			hft, _ := kvnuts.BytesToInt(item[6], true)
			hfto, _ := kvnuts.BytesToInt(item[7], true)
			go UpOneStart(ctx, string(k), hp, hps, hi, ht, hto, hfi, hft, hfto)
			go DownOneStart(ctx, string(k), hp, hps, hi, ht, hto, hfi, hft, hfto)
			//go UpOneStart(ctx, string(k), hp, hps, hi, ht, hto, hfi, hft, hfto)
			//go DownOneStart(ctx, string(k), hp, hps, hi, ht, hto, hfi, hft, hfto)
			go test(k)
		}
	}
}

func test(v []byte) {
	var l [][]byte
	for {
		time.Sleep(2 * time.Second)
		l, _ = kvnuts.SMem(c.NutsDB.Tag.Up, v)
		for _, s := range l {
			log.Println(string(v), "Success:", string(s))
		}
		l, _ = kvnuts.SMem(c.NutsDB.Tag.Down, v)
		for _, s := range l {
			log.Println(string(v), "Failure:", string(s))
		}
	}
}

//UpOneStart : up status health check driver
func UpOneStart(ctx context.Context, upstreamName, protocal, path string, sInterval, sTimes, sTimeout, fInterval, fTimes, fTimeout int) {
	for {
		time.Sleep(time.Duration(sInterval) * time.Millisecond)
		UpHC(upstreamName, protocal, path, fTimes, fTimeout)
	}
}

//DownOneStart : down status health check driver
func DownOneStart(ctx context.Context, upstreamName, protocal, path string, sInterval, sTimes, sTimeout, fInterval, fTimes, fTimeout int) {
	for {
		time.Sleep(time.Duration(fInterval) * time.Millisecond)
		DownHC(upstreamName, protocal, path, sTimes, sTimeout)
	}
}

//UpHC : up status ip&port check
func UpHC(upstreamName, protocal, path string, times, timeout int) {
	// get the upstream up list
	ipPort, _ := kvnuts.SMem(c.NutsDB.Tag.Up, upstreamName)
	if len(ipPort) == 0 {
		return
	}

	//check every ip port
	for i := 0; i < len(ipPort); i++ {
		ip := ipPort[i]
		if protocal == "http" {
			statusCode, _ := HTTP(string(ip)+path, timeout)
			log.Println(upstreamName, string(ip), statusCode)

			//the status code can not be in failure, and must be in success code.
			if !kvnuts.SIsMem(c.NutsDB.Tag.FailureCode+upstreamName, upstreamName, statusCode) &&
				kvnuts.SIsMem(c.NutsDB.Tag.SuccessCode+upstreamName, upstreamName, statusCode) {
				continue
			}
		} else {
			if TCP(string(ip), timeout) {
				continue
			}
		}

		if CodeCount(upstreamName+string(ip), "f", times) {
			_ = kvnuts.SRem(c.NutsDB.Tag.Up, upstreamName, ip)
			_ = kvnuts.SAdd(c.NutsDB.Tag.Down, upstreamName, ip)
		}
	}
}

//DownHC : down status ip&port check
func DownHC(upstreamName, protocal, path string, times, timeout int) {
	// get the upstream down list
	ipPort, _ := kvnuts.SMem(c.NutsDB.Tag.Down, upstreamName)
	if len(ipPort) == 0 {
		return
	}

	//check every ip port
	for i := 0; i < len(ipPort); i++ {
		ip := ipPort[i]
		log.Println(string(ip))
		if protocal == "http" {
			statusCode, _ := HTTP(string(ip)+path, timeout)
			log.Println(upstreamName, string(ip), statusCode)

			//the status code must be in success
			if !kvnuts.SIsMem(c.NutsDB.Tag.SuccessCode+upstreamName, upstreamName, statusCode) {
				continue
			}
		} else {
			if !TCP(string(ip), timeout) {
				continue
			}
		}

		if CodeCount(upstreamName+string(ip), "s", times) {
			_ = kvnuts.SRem(c.NutsDB.Tag.Down, upstreamName, ip)
			_ = kvnuts.SAdd(c.NutsDB.Tag.Up, upstreamName, ip)
		}
	}
}

//CodeCount : success && failed counter
func CodeCount(n, key string, times int) bool {
	log.Println(kvnuts.Get(n, key, "i"))
	_, nTime, err := kvnuts.Get(n, key, "i")

	//first be counted
	if err != nil {
		_ = kvnuts.Put(n, key, 1)
		return false
	}

	//counted times less than healthCheck items
	if nTime < times {
		_ = kvnuts.Put(n, key, nTime+1)
		return false
	}

	_ = kvnuts.Del(n, key)
	return true
}
