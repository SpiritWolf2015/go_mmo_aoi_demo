package main

import (
	"fmt"
	"github.com/aceld/zinx/zdecoder"
	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/zlog"
	"github.com/aceld/zinx/znet"
	"github.com/aceld/zinx/zpack"
	"gzc_zinx01/core"
	"gzc_zinx01/robot"
	"os"
	"os/signal"
	"time"
)

func main() {
	num := 5
	allRobots := make([]*Robot, 0, num)

	for i := 0; i < num; i++ {
		aiRobot := NewRobot()
		allRobots = append(allRobots, aiRobot)
		go aiRobot.start()
	}

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c

	fmt.Println("===exit===", sig)

	for _, r := range allRobots {
		r.stop()
	}
	time.Sleep(time.Second * 2)
}

type Robot struct {
	cli ziface.IClient
}

func NewRobot() *Robot {
	client := znet.NewClient("127.0.0.1", 8999)

	client.SetOnConnStart(DoClientConnectedBegin)
	client.SetOnConnStop(DoClientConnectedLost)

	// 注册编结码器
	client.SetPacket(zpack.NewDataPackLtv())
	client.SetDecoder(zdecoder.NewLTV_Little_Decoder())

	// 注册业务消息处理器
	client.AddRouter(1, &robot.SyncPidApi{})
	d := &robot.DefaultApi{}
	client.AddRouter(200, d)
	client.AddRouter(201, d)
	client.AddRouter(202, d)

	return &Robot{
		cli: client,
	}
}

func (r *Robot) start() {
	r.cli.Start()
}

func (r *Robot) stop() {
	r.cli.Stop()
}

func DoClientConnectedBegin(conn ziface.IConnection) {
	zlog.Debug("DoConnectionBegin is Called ... ")
}

func DoClientConnectedLost(conn ziface.IConnection) {
	// (在连接销毁之前，查询conn的Name，Home属性)
	zlog.Debug("DoClientConnectedLost is Called ... ")

	pID, err := conn.GetProperty("pID")
	if err != nil {
		zlog.Error("conn.GetProperty() error(%v)", err)
		return
	}

	zlog.Debug("Conn Property pid= ", pID)
	core.WorldMgrObj.RemovePlayerByPID(pID.(int32))
}
