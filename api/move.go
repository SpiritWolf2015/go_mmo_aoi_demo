package api

import (
	"fmt"
	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
	"github.com/golang/protobuf/proto"
	"gzc_zinx01/core"
	"gzc_zinx01/pb"
)

type MoveApi struct {
	znet.BaseRouter
}

func (*MoveApi) Handle(request ziface.IRequest) {
	msg := &pb.Position{}
	err := proto.Unmarshal(request.GetData(), msg)
	if nil != err {
		fmt.Println("[move][Handle] unmarshal err:", err)
		return
	}

	pID, err := request.GetConnection().GetProperty("pID")
	if nil != err {
		fmt.Println("[move][Handle] error not found pid in connection:", err)
		request.GetConnection().Stop()
		return
	}

	player := core.WorldMgrObj.GetPlayerByPID(pID.(int32))
	if nil == player {
		fmt.Println("[move][Handle] error not found player pid:", pID)
		request.GetConnection().Stop()
		return
	}

	player.UpdatePos(msg)
}
