package main

import (
	"fmt"
	"github.com/aceld/zinx/zdecoder"
	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
	"github.com/aceld/zinx/zpack"
	"gzc_zinx01/api"
	"gzc_zinx01/core"
)

func main() {
	s := znet.NewServer()

	// 注册链接的回调函数
	s.SetOnConnStart(OnConnectionAdd)
	s.SetOnConnStop(OnConnectionLost)

	// 注册业务消息处理器
	s.AddRouter(2, &api.WorldChatApi{})
	s.AddRouter(3, &api.MoveApi{})

	// 注册编结码器
	s.SetPacket(zpack.NewDataPackLtv())
	s.SetDecoder(zdecoder.NewLTV_Little_Decoder())

	// 启动服务器
	s.Serve()
}

func OnConnectionAdd(conn ziface.IConnection) {
	fmt.Println("[server][OnConnectionAdd] client addr:", conn.GetConnection().RemoteAddr().String())

	player := core.NewPlayer(conn)

	// 通知玩家id给客户端
	player.SyncPID()
	// 广播玩家出生位置
	player.SyncStartPosition()

	core.WorldMgrObj.AddPlayer(player)

	// 设置链接属性，玩家id
	conn.SetProperty("pID", player.PID)

	// 通知九宫格内玩家位置信息
	player.SyncSurroundingPos()
}

func OnConnectionLost(conn ziface.IConnection) {
	fmt.Println("[server][OnConnectionLost] client addr:", conn.GetConnection().RemoteAddr().String())

	pID, _ := conn.GetProperty("pID")
	var playerId int32
	if nil != pID {
		playerId = pID.(int32)
	}

	player := core.WorldMgrObj.GetPlayerByPID(playerId)
	if nil != player {
		player.LostConnection()
	}

	// todo 高并发情况下，会出现客户端断开链接，但服务器这却没有走到断开链接回调的情况
	fmt.Println("====> PlayerId: ", pID, " left =====")
}
