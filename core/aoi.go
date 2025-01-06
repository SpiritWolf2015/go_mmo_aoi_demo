package core

import "fmt"

const (
	AOI_MIN_X   int = 85
	AOI_MAX_X   int = 410
	AOI_COUNT_X int = 10
	AOI_MIN_Y   int = 75
	AOI_MAX_Y   int = 400
	AOI_COUNT_Y int = 20
)

type AOIManager struct {
	MinX   int
	MaxX   int
	CountX int
	MinY   int
	MaxY   int
	CountY int
	grids  map[int]*Grid
}

func NewAOIManager(minX, maxX, countX, minY, maxY, countY int) *AOIManager {
	aoiMgr := &AOIManager{
		MinX:   minX,
		MaxX:   maxX,
		CountX: countX,
		MinY:   minY,
		MaxY:   maxY,
		CountY: countY,
		grids:  make(map[int]*Grid),
	}

	for y := 0; y < countY; y++ {
		for x := 0; x < countX; x++ {
			// 计算格子索引ID，id= idY * nX + idX（利用格子坐标得到格子编号）
			gId := y*countX + x

			aoiMgr.grids[gId] = NewGrid(gId,
				aoiMgr.MinX+x*aoiMgr.gridWidth(),
				aoiMgr.MinX+(x+1)*aoiMgr.gridWidth(),
				aoiMgr.MinY+y*aoiMgr.gridLength(),
				aoiMgr.MinY+(y+1)*aoiMgr.gridLength())
		}
	}

	return aoiMgr
}

func (m *AOIManager) gridWidth() int {
	return (m.MaxX - m.MinX) / m.CountX
}

func (m *AOIManager) gridLength() int {
	return (m.MaxY - m.MinY) / m.CountY
}

func (m *AOIManager) String() string {
	s := fmt.Sprintf("AOIManagr:\nminX:%d, maxX:%d, countX:%d, minY:%d, maxY:%d, countY:%d\n Grids in AOI Manager:\n",
		m.MinX, m.MaxX, m.CountX, m.MinY, m.MaxY, m.CountY)
	for _, grID := range m.grids {
		s += fmt.Sprintln(grID)
	}

	return s
}

// 根据格子id得到当前周边的九宫格信息（包含id格子自身）
func (m *AOIManager) GetSurroundGridsByGID(gID int) (grids []*Grid) {
	if _, ok := m.grids[gID]; !ok {
		return
	}

	// 将当前格子加入到九宫格中去
	grids = append(grids, m.grids[gID])
	// 根据gID得到格子所在的坐标
	indexX, indexY := gID%m.CountX, gID/m.CountX
	// 新建一个临时存储周围格子的数组
	surroundGID := make([]int, 0)

	// 新建8个方向向量: 左上: (-1, -1), 左中: (-1, 0), 左下: (-1,1), 中上: (0,-1), 中下: (0,1), 右上:(1, -1)
	// 右中: (1, 0), 右下: (1, 1), 分别将这8个方向的方向向量按顺序写入x, y的分量数组
	dx := []int{-1, -1, -1, 0, 0, 1, 1, 1}
	dy := []int{-1, 0, 1, -1, 1, -1, 0, 1}

	// 根据8个方向向量, 得到周围点的相对坐标, 挑选出没有越界的坐标, 将坐标转换为gID
	for i := 0; i < 8; i++ {
		// 根据方向向量换算出来的下标，就是按照左上 左中 左下 中上 中下 右上 右中 右下的顺序来遍历
		newIndexX := indexX + dx[i]
		newIndexY := indexY + dy[i]

		// 再排除格子靠着上下左右边界的8种情况
		if newIndexX >= 0 && newIndexX < m.CountX && newIndexY >= 0 && newIndexY < m.CountY {
			surroundGID = append(surroundGID, newIndexY*m.CountX+newIndexX)
		}
	}

	for _, gID := range surroundGID {
		if _, ok := m.grids[gID]; !ok {
			continue
		}
		grids = append(grids, m.grids[gID])
	}

	return
}

// 通过横纵坐标获取对应的格子ID
func (m *AOIManager) GetGIDByPos(x, y float32) int {
	gx := (int(x) - m.MinX) / m.gridWidth()
	gy := (int(y) - m.MinY) / m.gridLength()

	return gy*m.CountX + gx
}

// 通过横纵坐标得到周边九宫格内的全部PlayerIDs
func (m *AOIManager) GetPIDsByPos(x, y float32) (playerIDs []int) {
	// 根据横纵坐标得到当前坐标属于哪个格子ID
	gID := m.GetGIDByPos(x, y)

	// 根据格子ID得到周边九宫格的信息
	grids := m.GetSurroundGridsByGID(gID)
	for _, v := range grids {
		playerIDs = append(playerIDs, v.GetPlayerIDs()...)
		//fmt.Printf("===> grID ID : %d, pIDs : %v  ====", v.GID, v.GetPlyerIDs())
	}

	return
}

// 通过GID获取当前格子的全部playerID
func (m *AOIManager) GetPIDsByGID(gID int) (playerIDs []int) {
	playerIDs = m.grids[gID].GetPlayerIDs()
	return
}

// 移除一个格子中的PlayerID
func (m *AOIManager) RemovePIDFromGrid(pID, gID int) {
	m.grids[gID].Remove(pID)
}

// 添加一个PlayerID到一个格子中
func (m *AOIManager) AddPIDToGrID(pID, gID int) {
	m.grids[gID].Add(pID)
}

// 通过横纵位置坐标添加一个Player到一个格子中
func (m *AOIManager) AddToGridByPos(pID int, x, y float32) {
	gID := m.GetGIDByPos(x, y)
	grid := m.grids[gID]
	if nil == grid {
		fmt.Printf("[aoi][AddToGridByPos]: grid is nil,x=%f,y=%f,pID=%d,gID=%d\n", x, y, pID, gID)
		return
	}
	grid.Add(pID)
}

// RemoveFromGrIDByPos Remove a Player from the corresponding grid based on horizontal and vertical coordinates
// 通过横纵位置坐标把一个Player从对应的格子中删除
func (m *AOIManager) RemoveFromGridByPos(pID int, x, y float32) {
	gID := m.GetGIDByPos(x, y)
	grid := m.grids[gID]
	grid.Remove(pID)
}
