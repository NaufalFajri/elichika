package story

import (
	"elichika/client"
	"elichika/client/request"
	"elichika/client/response"
	"elichika/enum"
	"elichika/generic"
	"elichika/handler/common"
	"elichika/item"
	"elichika/router"
	"elichika/subsystem/user_present"
	"elichika/userdata"
	"elichika/utils"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

func finishUserStoryMain(ctx *gin.Context) {
	reqBody := gjson.Parse(ctx.GetString("reqBody")).Array()[0].String()
	req := request.StoryMainRequest{}
	err := json.Unmarshal([]byte(reqBody), &req)
	utils.CheckErr(err)

	userId := int32(ctx.GetInt("user_id"))
	session := userdata.GetSession(ctx, userId)
	defer session.Close()

	if req.IsAutoMode.HasValue {
		session.UserStatus.IsAutoMode = req.IsAutoMode.Value
	}
	resp := response.StoryMainResponse{
		UserModelDiff: &session.UserModel,
	}

	if session.InsertUserStoryMain(req.CellId) { // newly inserted story, award some gem
		resp.FirstClearReward.Append(item.StarGem.Amount(10))
		user_present.AddPresent(session, client.PresentItem{
			Content:          item.StarGem.Amount(10),
			PresentRouteType: enum.PresentRouteTypeStoryMain,
			PresentRouteId:   generic.NewNullable(req.CellId),
		})
	}
	if req.MemberId.HasValue { // has a member -> select member thingy
		session.UpdateUserStoryMainSelected(req.CellId, req.MemberId.Value)
	}

	session.Finalize()
	common.JsonResponse(ctx, &resp)
}

func init() {
	router.AddHandler("/story/finishUserStoryMain", finishUserStoryMain)
}