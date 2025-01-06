package core

import (
	"fmt"
	"testing"
)

func TestNewAOIManager(t *testing.T) {
	aoiMgr := NewAOIManager(100, 300, 4, 200, 450, 5)
	fmt.Println(aoiMgr)
}

func TestAOIManagerSurroundGrIDsByGID(t *testing.T) {
	aoiMgr := NewAOIManager(0, 250, 5, 0, 250, 5)

	for k := range aoiMgr.grids {
		// (得到当前格子周边的九宫格)
		grIDs := aoiMgr.GetSurroundGridsByGID(k)

		// (得到九宫格所有的IDs)
		fmt.Println("gID : ", k, " grIDs len = ", len(grIDs))

		gIDs := make([]int, 0, len(grIDs))
		for _, grID := range grIDs {
			gIDs = append(gIDs, grID.GID)
		}
		fmt.Printf("grid ID: %d, surrounding grid IDs are %v\n", k, gIDs)
	}
}

func TestAOIManager_GetGIDByPos(t *testing.T) {
	aoiMgr := NewAOIManager(100, 300, 4, 200, 450, 5)

	gID := aoiMgr.GetGIDByPos(100, 200)
	if gID != 0 {
		t.Errorf("not first grid id:%d\n", gID)
	}

	gID = aoiMgr.GetGIDByPos(300, 450)
	if gID != 24 {
		t.Errorf("not last grid id:%d\n", gID)
	}
}

func TestAOIManager_GetPIDsByPos(t *testing.T) {
	aoiMgr := NewAOIManager(100, 300, 4, 200, 450, 5)
	aoiMgr.AddPIDToGrID(1001, 0)
	aoiMgr.AddPIDToGrID(1002, 0)

	pIDs := aoiMgr.GetPIDsByPos(100, 200)
	if pIDs[0] != 1001 || pIDs[1] != 1002 {
		t.Errorf("GetPIDsByPos error, gId:%d\n", 0)
	}
}

func TestAOIManager_GetPIDsByGID(t *testing.T) {
	aoiMgr := NewAOIManager(100, 300, 4, 200, 450, 5)
	aoiMgr.AddPIDToGrID(1001, 0)

	pIDs := aoiMgr.GetPIDsByGID(0)
	if pIDs[0] != 1001 {
		t.Errorf("GetPIDsByGID error, gId:%d\n", 0)
	}
}

func TestAOIManager_RemovePIDFromGrid(t *testing.T) {
	aoiMgr := NewAOIManager(100, 300, 4, 200, 450, 5)
	aoiMgr.AddPIDToGrID(1001, 0)

	pIDs := aoiMgr.GetPIDsByGID(0)
	if pIDs[0] != 1001 {
		t.Errorf("AddPIDToGrID error, gId:%d\n", 0)
	}

	aoiMgr.RemovePIDFromGrid(1001, 0)
	pIDs = aoiMgr.GetPIDsByGID(0)
	if len(pIDs) > 0 {
		t.Errorf("RemovePIDFromGrid error, gId:%d\n", 0)
	}
}

func TestAOIManager_AddPIDToGrID(t *testing.T) {
	aoiMgr := NewAOIManager(100, 300, 4, 200, 450, 5)
	aoiMgr.AddPIDToGrID(1001, 0)

	pIDs := aoiMgr.GetPIDsByGID(0)
	if pIDs[0] != 1001 {
		t.Errorf("AddPIDToGrID error, gId:%d\n", 0)
	}
}

func TestAOIManager_AddToGridByPos(t *testing.T) {
	aoiMgr := NewAOIManager(100, 300, 4, 200, 450, 5)
	aoiMgr.AddToGridByPos(1001, 100, 200)

	pIDs := aoiMgr.GetPIDsByGID(0)
	if pIDs[0] != 1001 {
		t.Errorf("AddToGridByPos error, gId:%d\n", 0)
	}
}

func TestAOIManager_RemoveFromGrid(t *testing.T) {
	aoiMgr := NewAOIManager(100, 300, 4, 200, 450, 5)
	aoiMgr.AddPIDToGrID(1001, 0)

	aoiMgr.RemoveFromGridByPos(1001, 100, 200)
	pIDs := aoiMgr.GetPIDsByGID(0)
	if len(pIDs) > 0 {
		t.Errorf("RemoveFromGrid error, gId:%d\n", 0)
	}
}
