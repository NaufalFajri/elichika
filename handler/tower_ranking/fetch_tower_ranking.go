package tower_ranking

import (
	"elichika/client"
	"elichika/client/request"
	"elichika/client/response"
	"elichika/enum"
	"elichika/generic"
	"elichika/handler/common"
	"elichika/router"
	"elichika/subsystem/user_tower"
	"elichika/userdata"
	"elichika/utils"

	"encoding/json"

	"github.com/gin-gonic/gin"
)

func fetchTowerRanking(ctx *gin.Context) {
	req := request.FetchTowerRankingRequest{}
	err := json.Unmarshal(*ctx.MustGet("reqBody").(*json.RawMessage), &req)
	utils.CheckErr(err)

	session := ctx.MustGet("session").(*userdata.Session)

	// TODO(ranking): return actual data for this
	resp := response.FetchTowerRankingResponse{}
	resp.TopRankingCells.Append(user_tower.GetTowerRankingCell(session, req.TowerId))
	resp.MyRankingCells.Append(user_tower.GetTowerRankingCell(session, req.TowerId))
	resp.FriendRankingCells.Append(user_tower.GetTowerRankingCell(session, req.TowerId))
	resp.RankingBorderInfo.Append(client.TowerRankingBorderInfo{
		RankingBorderVoltage: 0,
		RankingBorderMasterRow: client.TowerRankingBorderMasterRow{
			RankingType:  enum.EventCommonRankingTypeAll,
			UpperRank:    1,
			DisplayOrder: 1,
		}})
	resp.RankingBorderInfo.Append(client.TowerRankingBorderInfo{
		RankingBorderVoltage: 0,
		RankingBorderMasterRow: client.TowerRankingBorderMasterRow{
			RankingType:  enum.EventCommonRankingTypeFriend,
			UpperRank:    1,
			DisplayOrder: 1,
		}})
	resp.MyOrder = generic.NewNullable(int32(1))

	common.JsonResponse(ctx, &resp)
}

func init() {
	router.AddHandler("/towerRanking/fetchTowerRanking", fetchTowerRanking)
}
