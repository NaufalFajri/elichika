package user_profile

import (
	"elichika/client/request"
	"elichika/client/response"
	"elichika/handler/common"
	"elichika/router"
	"elichika/subsystem/user_profile"
	"elichika/userdata"
	"elichika/utils"

	"encoding/json"

	"github.com/gin-gonic/gin"
)

func setCommboLive(ctx *gin.Context) {
	req := request.SetLiveRequest{}
	err := json.Unmarshal(*ctx.MustGet("reqBody").(*json.RawMessage), &req)
	utils.CheckErr(err)

	session := ctx.MustGet("session").(*userdata.Session)

	user_profile.SetCommboLive(session, req.LiveDifficultyMasterId)

	common.JsonResponse(ctx, response.SetLiveResponse{
		LiveDifficultyMasterId: req.LiveDifficultyMasterId,
	})
}

func init() {
	router.AddHandler("/userProfile/setCommboLive", setCommboLive)
}
