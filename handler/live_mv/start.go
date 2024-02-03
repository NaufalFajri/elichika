package live_mv

import (
	"elichika/client/response"
	"elichika/handler/common"
	"elichika/router"
	"elichika/userdata"

	"github.com/gin-gonic/gin"
)

func start(ctx *gin.Context) {
	// we don't really need the request
	// maybe it's once needed or it's only used for gathering data
	// reqBody := gjson.Parse(ctx.GetString("reqBody")).Array()[0].String()
	// req := request.StartLiveMvRequest{}
	// err := json.Unmarshal([]byte(reqBody), &req)
	// utils.CheckErr(err)

	userId := int32(ctx.GetInt("user_id"))
	session := userdata.GetSession(ctx, userId)
	defer session.Close()

	common.JsonResponse(ctx, &response.StartLiveMvResponse{
		UniqId:        session.Time.UnixNano(),
		UserModelDiff: &session.UserModel,
	})
}

func init() {
	router.AddHandler("/liveMv/start", start)
}
