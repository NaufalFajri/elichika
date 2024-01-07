package userdata

import (
	"elichika/model"
	"elichika/protocol/response"
)

func (session *Session) GetRankingUser() response.RankingUser {
	card := session.GetUserCard(session.UserStatus.RecommendCardMasterID)
	return response.RankingUser{
		UserID: session.UserStatus.UserID,
		UserName: model.LocalizedText{
			DotUnderText: session.UserStatus.Name.DotUnderText,
		},
		UserRank:               session.UserStatus.Rank,
		CardMasterID:           card.CardMasterID,
		Level:                  card.Level,
		IsAwakening:            card.IsAwakening,
		IsAllTrainingActivated: card.IsAllTrainingActivated,
		EmblemMasterID:         session.UserStatus.EmblemID,
	}
}