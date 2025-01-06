package core

import "sync"

type WorldManager struct {
	AoiMgr *AOIManager
	// 所有在线玩家
	Players map[int32]*Player
	// 保护在线玩家集合的锁
	pLock sync.RWMutex
}

var WorldMgrObj *WorldManager

func init() {
	WorldMgrObj = &WorldManager{
		Players: make(map[int32]*Player),
		AoiMgr:  NewAOIManager(AOI_MIN_X, AOI_MAX_X, AOI_COUNT_X, AOI_MIN_Y, AOI_MAX_Y, AOI_COUNT_Y),
	}
}

func (wm *WorldManager) AddPlayer(player *Player) {
	if nil == player {
		return
	}

	wm.pLock.Lock()
	wm.Players[player.PID] = player
	wm.pLock.Unlock()

	// 注意这里给函数的y轴传参是player的z轴，将unity的顶视图映射成服务器的平面二维视图
	wm.AoiMgr.AddToGridByPos(int(player.PID), player.X, player.Z)
}

func (wm *WorldManager) RemovePlayerByPID(pID int32) {
	wm.pLock.Lock()
	delete(wm.Players, pID)
	wm.pLock.Unlock()
}

func (wm *WorldManager) GetPlayerByPID(pID int32) *Player {
	wm.pLock.RLock()
	defer wm.pLock.RUnlock()

	return wm.Players[pID]
}

// 获取所有在线的玩家
func (wm *WorldManager) GetAllPlayers() []*Player {
	wm.pLock.RLock()
	defer wm.pLock.RUnlock()

	players := make([]*Player, 0, len(wm.Players))

	for _, player := range wm.Players {
		players = append(players, player)
	}

	return players
}

// 获取指定网格的所有玩家
func (wm *WorldManager) GetPlayersByGID(gID int) []*Player {
	pIDs := wm.AoiMgr.grids[gID].GetPlayerIDs()
	players := make([]*Player, 0, len(pIDs))

	wm.pLock.RLock()
	for _, pID := range pIDs {
		players = append(players, wm.Players[int32(pID)])
	}
	wm.pLock.RUnlock()

	return players
}
