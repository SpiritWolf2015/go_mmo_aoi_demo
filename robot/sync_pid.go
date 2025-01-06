package robot

import (
	"fmt"
	"github.com/aceld/zinx/ziface"
	"github.com/aceld/zinx/znet"
	"github.com/golang/protobuf/proto"
	"gzc_zinx01/core"
	"gzc_zinx01/pb"
	"math/rand"
	"time"
)

type SyncPidApi struct {
	znet.BaseRouter
}

func (*SyncPidApi) Handle(request ziface.IRequest) {
	msg := &pb.SyncPID{}
	err := proto.Unmarshal(request.GetData(), msg)
	if nil != err {
		fmt.Println("[sync_pid][Handle] unmarshal err:", err)
		return
	}

	player := core.NewPlayer(request.GetConnection())
	player.PID = msg.PID
	core.WorldMgrObj.AddPlayer(player)

	// 设置链接属性，玩家id
	request.GetConnection().SetProperty("pID", player.PID)

	go func() {
		fmt.Println("[sync_pid][Handle] start robotAI pid:", player.PID)
		for {
			robotAI(player)
			time.Sleep(time.Second)
		}
	}()
}

// 聊天或者移动
func robotAI(player *core.Player) {
	//随机获得动作
	tp := rand.Intn(1000)
	if tp < 500 {
		chatArr := make([]string, 0, 5)
		chatArr = append(chatArr, "good luck")
		chatArr = append(chatArr, "hello world")
		chatArr = append(chatArr, "go go go")
		chatArr = append(chatArr, "java java java")
		chatArr = append(chatArr, "rust rust rust")

		idx := rand.Intn(5)

		content := fmt.Sprintf("我是player:%d, %s", player.PID, chatArr[idx])
		msg := &pb.Talk{
			Content: content,
		}
		player.SendMsg(2, msg)
	} else {
		//移动
		x := player.X
		z := player.Z

		randPos := rand.Intn(1000)
		distance := rand.Intn(5) + 10
		if randPos > 500 {
			x -= float32(rand.Intn(distance))
			z -= float32(rand.Intn(distance))
		} else {
			x += float32(rand.Intn(distance))
			z += float32(rand.Intn(distance))
		}

		//纠正坐标位置
		if x > float32(core.AOI_MAX_X) {
			x = float32(core.AOI_MAX_X)
		} else if x < float32(core.AOI_MIN_X) {
			x = float32(core.AOI_MIN_X)
		}

		if z > float32(core.AOI_MAX_Y) {
			z = float32(core.AOI_MAX_Y)
		} else if z < float32(core.AOI_MIN_Y) {
			z = float32(core.AOI_MIN_Y)
		}

		//移动方向角度
		randV := rand.Intn(1000)
		v := player.V
		if randV > 500 {
			v = 25
		} else {
			v = 335
		}
		//封装Position消息
		msg := &pb.Position{
			X: x,
			Y: player.Y,
			Z: z,
			V: v,
		}

		fmt.Println(fmt.Sprintf("player ID: %d Walking.. at(%f,%f,%f,%f)", player.PID, player.X, player.Y, player.Z, player.V))
		//发送移动MsgID:3的指令
		player.SendMsg(3, msg)
	}
}
