package card

import (
	"elichika/client/request"
	"elichika/client/response"
	"elichika/handler/common"
	"elichika/router"
	"elichika/subsystem/user_card"
	"elichika/userdata"
	"elichika/utils"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

func changeIsAwakeningImage(ctx *gin.Context) {
	reqBody := gjson.Parse(ctx.GetString("reqBody")).Array()[0].String()
	req := request.ChangeIsAwakeningImageRequest{}
	err := json.Unmarshal([]byte(reqBody), &req)
	utils.CheckErr(err)

	userId := int32(ctx.GetInt("user_id"))
	session := userdata.GetSession(ctx, userId)
	defer session.Close()

	userCard := user_card.GetUserCard(session, req.CardMasterId)
	userCard.IsAwakeningImage = req.IsAwakeningImage
	user_card.UpdateUserCard(session, userCard)

	session.Finalize()
	common.JsonResponse(ctx, response.ChangeIsAwakeningImageResponse{
		UserModelDiff: &session.UserModel,
	})
}

func init() {
	router.AddHandler("/card/changeIsAwakeningImage", changeIsAwakeningImage)
}
