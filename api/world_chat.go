package api

import (
	"fmt"
	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
	"github.com/golang/protobuf/proto"
	"gzc_zinx01/core"
	"gzc_zinx01/pb"
)

type WorldChatApi struct {
	znet.BaseRouter
}

func (*WorldChatApi) Handle(request ziface.IRequest) {
	msg := &pb.Talk{}
	err := proto.Unmarshal(request.GetData(), msg)
	if nil != err {
		fmt.Println("[world_chat][Handle] error unmarshal:", err)
		return
	}

	pID, err := request.GetConnection().GetProperty("pID")
	if nil != err {
		fmt.Println("[world_chat][Handle] error not found connection property pid:", pID)
		request.GetConnection().Stop()
		return
	}

	player := core.WorldMgrObj.GetPlayerByPID(pID.(int32))
	if nil == player {
		fmt.Println("[world_chat][Handle] error world not found player pid:", pID)
		request.GetConnection().Stop()
		return
	}

	player.Talk(msg.Content)
}
