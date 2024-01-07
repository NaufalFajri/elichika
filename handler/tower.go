package handler

import (
	"elichika/config"
	"elichika/enum"
	"elichika/gamedata"
	"elichika/model"
	"elichika/protocol/request"
	"elichika/protocol/response"
	"elichika/userdata"
	"elichika/utils"

	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	// "github.com/tidwall/sjson"
)

func FetchTowerSelect(ctx *gin.Context) {
	// there's no request body
	userID := ctx.GetInt("user_id")
	session := userdata.GetSession(ctx, userID)
	defer session.Close()

	// no need to return anything, the same use database for this
	respObj := response.FetchTowerSelectResponse{
		TowerIDs:      []int{},
		UserModelDiff: &session.UserModel,
	}

	respBytes, _ := json.Marshal(respObj)
	resp := SignResp(ctx, string(respBytes), config.SessionKey)
	ctx.Header("Content-Type", "application/json")
	ctx.String(http.StatusOK, resp)
}

func FetchTowerTop(ctx *gin.Context) {
	reqBody := gjson.Parse(ctx.GetString("reqBody")).Array()[0].String()
	req := request.FetchTowerTopRequest{}
	err := json.Unmarshal([]byte(reqBody), &req)
	utils.CheckErr(err)

	userID := ctx.GetInt("user_id")
	session := userdata.GetSession(ctx, userID)
	defer session.Close()
	gamedata := ctx.MustGet("gamedata").(*gamedata.Gamedata)

	respObj := response.FetchTowerTopResponse{
		TowerCardUsedCountRows: session.GetUserTowerCardUsedList(req.TowerID),
		UserModelDiff:          &session.UserModel,
		IsShowUnlockEffect:     false,
		// other fields are for DLP with voltage ranking
	}

	userTower := session.GetUserTower(req.TowerID)
	tower := gamedata.Tower[req.TowerID]
	if userTower.ClearedFloor == userTower.ReadFloor {
		tower := gamedata.Tower[req.TowerID]
		if userTower.ReadFloor < tower.FloorCount {
			userTower.ReadFloor += 1
			respObj.IsShowUnlockEffect = true
			// unlock all the bonus live at once
			for ; userTower.ReadFloor < tower.FloorCount; userTower.ReadFloor++ {
				if tower.Floor[userTower.ReadFloor].TowerCellType != enum.TowerCellTypeBonusLive {
					break
				}
			}
		}
	}
	session.UpdateUserTower(userTower)

	// if tower with voltage ranking, then we have to prepare that
	if tower.IsVoltageRanked {
		//
		// EachBonusLiveVoltage should be filled with zero for everything, then fill in the score

		respObj.EachBonusLiveVoltage = make([]int, tower.FloorCount)
		respObj.Order = new(int)
		*respObj.Order = 1
		// fetch the score
		scores := session.GetUserTowerVoltageRankingScores(req.TowerID)
		for _, score := range scores {
			respObj.EachBonusLiveVoltage[score.FloorNo-1] = score.Voltage
		}
	}

	session.Finalize("", "dummy")

	respBytes, _ := json.Marshal(respObj)
	resp := SignResp(ctx, string(respBytes), config.SessionKey)
	ctx.Header("Content-Type", "application/json")
	ctx.String(http.StatusOK, resp)
}

func ClearedTowerFloor(ctx *gin.Context) {
	reqBody := gjson.Parse(ctx.GetString("reqBody")).Array()[0].String()
	req := request.ClearedTowerFloorRequest{}
	err := json.Unmarshal([]byte(reqBody), &req)
	utils.CheckErr(err)

	userID := ctx.GetInt("user_id")
	session := userdata.GetSession(ctx, userID)
	defer session.Close()

	respObj := response.ClearedTowerFloorResponse{
		UserModelDiff:      &session.UserModel,
		IsShowUnlockEffect: false,
	}

	userTower := session.GetUserTower(req.TowerID)
	if userTower.ClearedFloor < req.FloorNo {
		userTower.ClearedFloor = req.FloorNo
		session.UpdateUserTower(userTower)
	}
	session.UserStatus.IsAutoMode = req.IsAutoMode
	session.Finalize("", "dummy")

	respBytes, _ := json.Marshal(respObj)
	resp := SignResp(ctx, string(respBytes), config.SessionKey)
	ctx.Header("Content-Type", "application/json")
	ctx.String(http.StatusOK, resp)
}

func RecoveryTowerCardUsed(ctx *gin.Context) {
	reqBody := gjson.Parse(ctx.GetString("reqBody")).Array()[0].String()
	req := request.RecoveryTowerCardUsedRequest{}
	err := json.Unmarshal([]byte(reqBody), &req)
	utils.CheckErr(err)

	userID := ctx.GetInt("user_id")
	session := userdata.GetSession(ctx, userID)
	defer session.Close()
	gamedata := ctx.MustGet("gamedata").(*gamedata.Gamedata)

	tower := gamedata.Tower[req.TowerID]

	for _, cardMasterID := range req.CardMasterIDs {
		cardUsedCount := session.GetUserTowerCardUsed(req.TowerID, cardMasterID)
		cardUsedCount.UsedCount--
		cardUsedCount.RecoveredCount++
		session.UpdateUserTowerCardUsed(cardUsedCount)
	}
	// remove the item
	has := session.GetUserResource(enum.ContentTypeRecoveryTowerCardUsedCount, 24001).Content.ContentAmount
	if has >= int64(len(req.CardMasterIDs)) {
		session.RemoveResource(model.Content{
			ContentType:   enum.ContentTypeRecoveryTowerCardUsedCount,
			ContentID:     24001,
			ContentAmount: int64(len(req.CardMasterIDs)),
		})
	} else {
		session.RemoveResource(model.Content{
			ContentType:   enum.ContentTypeRecoveryTowerCardUsedCount,
			ContentID:     24001,
			ContentAmount: has,
		})
		session.RemoveSnsCoin((int64(len(req.CardMasterIDs)) - has) * int64(tower.RecoverCostBySnsCoin))

	}
	session.Finalize("", "dummy")
	respObj := response.RecoveryTowerCardUsedResponse{
		TowerCardUsedCountRows: session.GetUserTowerCardUsedList(req.TowerID),
		UserModelDiff:          &session.UserModel,
	}

	respBytes, _ := json.Marshal(respObj)
	resp := SignResp(ctx, string(respBytes), config.SessionKey)
	ctx.Header("Content-Type", "application/json")
	ctx.String(http.StatusOK, resp)
}

func RecoveryTowerCardUsedAll(ctx *gin.Context) {
	reqBody := gjson.Parse(ctx.GetString("reqBody")).Array()[0].String()
	req := request.RecoveryTowerCardUsedAllRequest{}
	err := json.Unmarshal([]byte(reqBody), &req)
	utils.CheckErr(err)

	userID := ctx.GetInt("user_id")
	session := userdata.GetSession(ctx, userID)
	defer session.Close()

	respObj := response.RecoveryTowerCardUsedResponse{
		TowerCardUsedCountRows: session.GetUserTowerCardUsedList(req.TowerID),
		UserModelDiff:          &session.UserModel,
	}
	for i := range respObj.TowerCardUsedCountRows {
		respObj.TowerCardUsedCountRows[i].UsedCount = 0
		respObj.TowerCardUsedCountRows[i].RecoveredCount = 0
		session.UpdateUserTowerCardUsed(respObj.TowerCardUsedCountRows[i])
	}
	userTower := session.GetUserTower(req.TowerID)
	session.UpdateUserTower(userTower)

	session.Finalize("", "dummy")
	respBytes, _ := json.Marshal(respObj)
	resp := SignResp(ctx, string(respBytes), config.SessionKey)
	ctx.Header("Content-Type", "application/json")
	ctx.String(http.StatusOK, resp)
}

func FetchTowerRanking(ctx *gin.Context) {
	reqBody := gjson.Parse(ctx.GetString("reqBody")).Array()[0].String()
	req := request.FetchTowerRankingRequest{}
	err := json.Unmarshal([]byte(reqBody), &req)
	utils.CheckErr(err)

	userID := ctx.GetInt("user_id")
	session := userdata.GetSession(ctx, userID)
	defer session.Close()

	// TODO(multiplayer ranking): return actual data for this
	respObj := response.FetchTowerRankingResponse{
		MyOrder: 1,
	}
	towerRankingCell := session.GetTowerRankingCell(req.TowerID)
	respObj.TopRankingCells = append(respObj.TopRankingCells, towerRankingCell)
	respObj.MyRankingCells = append(respObj.MyRankingCells, towerRankingCell)
	respObj.FriendRankingCells = append(respObj.FriendRankingCells, towerRankingCell)
	respObj.RankingBorderInfo = append(respObj.RankingBorderInfo,
		response.TowerRankingBorderInfo{
			RankingBorderVoltage: 0,
			RankingBorderMasterRow: response.TowerRankingBorderMasterRow{
				RankingType:  enum.EventCommonRankingTypeAll,
				UpperRank:    1,
				LowerRank:    1,
				DisplayOrder: 1,
			}})
	respObj.RankingBorderInfo = append(respObj.RankingBorderInfo,
		response.TowerRankingBorderInfo{
			RankingBorderVoltage: 0,
			RankingBorderMasterRow: response.TowerRankingBorderMasterRow{
				RankingType:  enum.EventCommonRankingTypeFriend,
				UpperRank:    1,
				LowerRank:    1,
				DisplayOrder: 1,
			}})
	respBytes, _ := json.Marshal(respObj)
	resp := SignResp(ctx, string(respBytes), config.SessionKey)
	ctx.Header("Content-Type", "application/json")
	ctx.String(http.StatusOK, resp)
}