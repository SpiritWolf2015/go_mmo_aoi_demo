package core

import (
	"fmt"
	"github.com/aceld/zinx/ziface"
	"github.com/golang/protobuf/proto"
	"gzc_zinx01/pb"
	"math/rand"
	"sync"
	"time"
)

type Player struct {
	PID  int32
	Conn ziface.IConnection
	X    float32 // 注意，服务器是映射的客户端的顶视图为平面坐标，所以x是横坐标(还是x轴不变)，z是纵坐标(y轴)
	Z    float32
	Y    float32 // 3d世界的高度值
	V    float32 // 旋转角度，0-360度
}

// 玩家id生成器
var PIDGen int32 = 1

// 保护玩家id生成器的锁
var IDLock sync.Mutex

func NewPlayer(conn ziface.IConnection) *Player {
	IDLock.Lock()
	id := PIDGen
	PIDGen++
	IDLock.Unlock()

	p := &Player{
		PID:  id,
		Conn: conn,
		X:    float32(AOI_MIN_X + 50 + rand.Intn(10)),
		Y:    0,
		Z:    float32(AOI_MIN_Y + 50 + rand.Intn(10)),
		V:    0,
	}
	return p
}

// 通知玩家id给客户端
func (p *Player) SyncPID() {
	data := &pb.SyncPID{
		PID: p.PID,
	}

	p.SendMsg(1, data)
}

// 发送消息给客户端，将protobuff数据序列化后再发送
func (p *Player) SendMsg(msgID uint32, data proto.Message) {
	if nil == p.Conn {
		fmt.Println("[player][SendMsg]01 network connection closed in playerId:", p.PID)
		return
	}

	msg, err := proto.Marshal(data)
	if nil != err {
		fmt.Println("[player][SendMsg]02 marshal pb msg error:", err)
		return
	}

	// todo 判活一直是false
	//if !p.Conn.IsAlive() {
	//	fmt.Printf("[player][SendMsg]03 pid:%d\n", p.PID)
	//	return
	//}

	// todo 高并发情况下会出现往已经断开的链接发消息的报错
	if err := p.Conn.SendMsg(msgID, msg); nil != err {
		fmt.Printf("[player][SendMsg]04 error:%v,pid:%d\n", err, p.PID)
		p.Conn.Stop()
		p.Conn = nil
		return
	}
}

// 通知玩家的出生位置
func (p *Player) SyncStartPosition() {
	msg := &pb.BroadCast{
		PID: p.PID,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	p.SendMsg(200, msg)
}

// 广播聊天消息
func (p *Player) Talk(content string) {
	msg := &pb.BroadCast{
		PID: p.PID,
		Tp:  1,
		Data: &pb.BroadCast_Content{
			Content: content,
		},
	}

	players := WorldMgrObj.GetAllPlayers()

	for _, player := range players {
		player.SendMsg(200, msg)
	}
}

// 1.广播通知九宫格其他玩家"我"的位置消息
// 2.将九宫格内其他玩家的位置通知给“我”
func (p *Player) SyncSurroundingPos() {
	// 通过玩家位置，获取周边九宫格内所有玩家id
	pIds := WorldMgrObj.AoiMgr.GetPIDsByPos(p.X, p.Z)

	players := make([]*Player, 0, len(pIds))
	for _, pID := range pIds {
		players = append(players, WorldMgrObj.GetPlayerByPID(int32(pID)))
	}

	msg := &pb.BroadCast{
		PID: p.PID,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	// 1.广播通知九宫格其他玩家"我"的位置消息
	for _, player := range players {
		player.SendMsg(200, msg)
	}

	// 2.将九宫格内其他玩家的位置通知给“我”
	otherPlayersData := make([]*pb.Player, 0, len(players))
	for _, player := range players {
		p := &pb.Player{
			PID: player.PID,
			P: &pb.Position{
				X: player.X,
				Y: player.Y,
				Z: player.Z,
				V: player.V,
			},
		}
		otherPlayersData = append(otherPlayersData, p)
	}

	SyncOtherPlayersMsg := &pb.SyncPlayers{
		Ps: otherPlayersData[0:],
	}
	p.SendMsg(202, SyncOtherPlayersMsg)
}

func (p *Player) UpdatePos(msg *pb.Position) {
	oldGID := WorldMgrObj.AoiMgr.GetGIDByPos(p.X, p.Z)
	newGID := WorldMgrObj.AoiMgr.GetGIDByPos(msg.X, msg.Z)

	p.X = msg.X
	p.Z = msg.Z
	p.Y = msg.Y
	p.V = msg.V

	// 位置变化导致移动到新的格子里去了
	if oldGID != newGID {
		// 将玩家从旧格子中删除，并添加到新格子中去
		WorldMgrObj.AoiMgr.RemovePIDFromGrid(int(p.PID), oldGID)
		WorldMgrObj.AoiMgr.AddPIDToGrID(int(p.PID), newGID)

		p.OnAoiGridChanged(oldGID, newGID)
	}

	// 玩家所在格子未变化，广播新位置
	posMsg := &pb.BroadCast{
		PID: p.PID,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	// 获取九宫格内所有玩家
	players := p.GetSurroundingPlayers()
	for _, player := range players {
		player.SendMsg(200, posMsg)
	}
}

func (p *Player) OnAoiGridChanged(oldGridId, newGridId int) {
	oldGrids := WorldMgrObj.AoiMgr.GetSurroundGridsByGID(oldGridId)
	// 旧九宫格有哪些格子
	oldGridIDsMap := make(map[int]bool, len(oldGrids))
	for _, grid := range oldGrids {
		oldGridIDsMap[grid.GID] = true
	}

	newGrids := WorldMgrObj.AoiMgr.GetSurroundGridsByGID(newGridId)
	// 新九宫格有哪些格子
	newGridIDsMap := make(map[int]bool, len(newGrids))
	for _, grid := range newGrids {
		newGridIDsMap[grid.GID] = true
	}

	//----------玩家从视野消息处理-------------

	// 玩家消失消息
	offlineMsg := &pb.SyncPID{
		PID: p.PID,
	}

	// 找出在旧九宫格中出现，但不在新九宫格中出现的格子
	leavingGrids := make([]*Grid, 0)
	for _, grid := range oldGrids {
		if _, ok := newGridIDsMap[grid.GID]; !ok {
			leavingGrids = append(leavingGrids, grid)
		}
	}

	for _, grid := range leavingGrids {
		// 需要消失格子的所有玩家
		players := WorldMgrObj.GetPlayersByGID(grid.GID)
		for _, player := range players {
			// 让“我”从其他玩家视野中消失
			player.SendMsg(201, offlineMsg)

			// 让其他玩家从“我”的视野中消失
			otherOfflineMsg := &pb.SyncPID{
				PID: player.PID,
			}

			p.SendMsg(201, otherOfflineMsg)
			// todo 这里如果不加sleep会出现bug，走进新格子后显示格子里其他玩家不全，但是控制台打印的玩家id列表是对的
			time.Sleep(200 * time.Millisecond)
		}
	}

	//------------------玩家从视野出现的处理---------------------

	// 玩家出现消息
	onlineMsg := &pb.BroadCast{
		PID: p.PID,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	// 找出在新九宫格中出现，但不在旧九宫格中出现的格子
	enteringGrids := make([]*Grid, 0)
	for _, grid := range newGrids {
		if _, ok := oldGridIDsMap[grid.GID]; !ok {
			enteringGrids = append(enteringGrids, grid)
		}
	}

	for _, grid := range enteringGrids {
		// 需要新出现格子的所有玩家
		players := WorldMgrObj.GetPlayersByGID(grid.GID)
		fmt.Printf("[player][OnAoiGridChanged]------------------appear players start,pid:%d,gid:%d\n", p.PID, grid.GID)
		for _, player := range players {
			// 让“我”从其他玩家视野中出现
			player.SendMsg(200, onlineMsg)

			// 让其他玩家从“我”的视野中出现
			otherOnlineMsg := &pb.BroadCast{
				PID: player.PID,
				Tp:  2,
				Data: &pb.BroadCast_P{
					P: &pb.Position{
						X: player.X,
						Y: player.Y,
						Z: player.Z,
						V: player.V,
					},
				},
			}
			fmt.Printf("[player][OnAoiGridChanged] pid:%d,gid:%d, appear other playerId:%d\n ", p.PID, grid.GID, player.PID)
			p.SendMsg(200, otherOnlineMsg)
			time.Sleep(200 * time.Millisecond)
		}
		fmt.Printf("[player][OnAoiGridChanged]------------------appear players end,pid:%d,gid:%d\n", p.PID, grid.GID)
	}
}

// 获取玩家所在位置九宫格内所有玩家(包括玩家本人)
func (p *Player) GetSurroundingPlayers() []*Player {
	pIDs := WorldMgrObj.AoiMgr.GetPIDsByPos(p.X, p.Z)

	players := make([]*Player, 0, len(pIDs))
	for _, pID := range pIDs {
		player := WorldMgrObj.GetPlayerByPID(int32(pID))
		if player == nil {
			continue
		}
		players = append(players, player)
	}

	return players
}

// 网络链接断开，玩家下线相关的处理
func (p *Player) LostConnection() {
	players := p.GetSurroundingPlayers()

	msg := &pb.SyncPID{
		PID: p.PID,
	}

	// 广播通知九宫格玩家“我”下线
	for _, player := range players {
		if p.PID == player.PID {
			continue
		}
		player.SendMsg(201, msg)
	}

	WorldMgrObj.AoiMgr.RemoveFromGridByPos(int(p.PID), p.X, p.Z)
	WorldMgrObj.RemovePlayerByPID(p.PID)
}
