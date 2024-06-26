package unlock_scene

import (
	"elichika/client/request"
	"elichika/client/response"
	"elichika/enum"
	"elichika/handler/common"
	"elichika/router"
	"elichika/subsystem/user_unlock_scene"
	"elichika/userdata"
	"elichika/utils"

	"encoding/json"

	"github.com/gin-gonic/gin"
)

func saveUnlockedScene(ctx *gin.Context) {
	req := request.SaveUnlockedSceneRequest1{}
	err := json.Unmarshal(*ctx.MustGet("reqBody").(*json.RawMessage), &req)
	utils.CheckErr(err)

	session := ctx.MustGet("session").(*userdata.Session)

	for _, sceneType := range req.UnlockSceneTypes.Slice {
		user_unlock_scene.UnlockScene(session, sceneType, enum.UnlockSceneStatusOpened)
	}

	common.JsonResponse(ctx, response.UserModelResponse{
		UserModel: &session.UserModel,
	})
}

func init() {
	router.AddHandler("/unlockScene/saveUnlockedScene", saveUnlockedScene)
}
