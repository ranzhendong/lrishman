package errorhandle

import (
	"github.com/thinkeridea/go-extend/exstrings"
	"strconv"
	"time"
)

type MyError struct {
	Error        string
	Message      string
	Code         int
	TimeStamp    time.Time
	ExecutorTime string
}

var (
	mux       = make(map[int]string)
	muxS      = make(map[int]string)
	randSlice = make([]int, 3)
)

//registered
/*
000 successful

1-9 method

001-030 system error

140-150 system status

101 - 200 etcd error


*/
func init() {
	muxS[0] = "ServeHTTP: "
	muxS[1] = "Upstream GET: "
	muxS[2] = "Upstream PUT: "
	muxS[3] = "Upstream POST: "
	muxS[4] = "Upstream PATCH: "
	muxS[5] = "Upstream DELETE: "
	muxS[6] = "Viper Watcher: "

	mux[000] = "Successful"
	mux[001] = "Upstream: "
	mux[002] = "INIT: Loading Body Failed"
	mux[003] = "JudgeValidator Error"
	mux[004] = "Json: Marshal Error"
	mux[005] = "Json: UNMarshal Error"
	mux[006] = "WriteString Error"
	mux[007] = "Not Support Method Error"
	mux[010] = "Url Not Exist"
	mux[011] = "HTTP Server Init Error"

	mux[140] = "Config Change Reloading"
	mux[141] = "IrishMan Is Running With Execute Path"
	mux[142] = "Config Read Error"

	mux[101] = "Etcd Put: Put Key Error"
	mux[102] = "Etcd Get: Key Not Exist Error"
	mux[103] = "Etcd Get: Repeat Key Error"
	mux[104] = "Etcd GetALL: No Key Error"
	mux[105] = "Etcd Delete: Error"
	mux[106] = "Etcd Delete: Etcd Key's Pool Has One ServerList At Least, Delete Canceled !"
	mux[107] = "Etcd Delete: Etcd Key's Pool Has One ServerList At Least, Can Not Delete Them ALL !"
}

//register error to message
func (self *MyError) Messages() {
	defer func() {
		_ = recover()
		if self.Message == "" {
			self.Message = "No Error Match"
		} else if self.Error == "" {
			self.Error = self.Message
		} else if self.Error == "" && self.Message == "" {
			self.Error = "No Error Match"
			self.Message = "No Error Match"
		}
	}()
	self.Message = muxS[self.Code/1000%10] + mux[Code(self.Code)]
}

//error log handler
func ErrorLog(code int, content ...string) string {
	if content == nil {
		return muxS[code/1000%10] + mux[Code(code)]
	}
	return muxS[code/1000%10] + mux[Code(code)] + content[0]
}

//timer clock
func (self *MyError) Clock() {
	//if TimeStamp is none
	if len(time.Since(self.TimeStamp).String()) > 20 {
		self.ExecutorTime = time.Since(time.Now()).String()
		return
	}
	self.ExecutorTime = time.Since(self.TimeStamp).String()
}

//code cut out
func Code(e int) (a int) {
	randSlice[0] = e / 100 % 10
	randSlice[1] = e / 10 % 10
	randSlice[2] = e / 1 % 10
	a, _ = strconv.Atoi(exstrings.JoinInts(randSlice, ""))
	return
}
