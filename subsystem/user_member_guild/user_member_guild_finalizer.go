package user_member_guild

import (
	"elichika/userdata"
	"elichika/utils"
)

func userMemberGuildFinalizer(session *userdata.Session) {
	for _, userMemberGuild := range session.UserModel.UserMemberGuildById.Map {
		affected, err := session.Db.Table("u_member_guild").Where("user_id = ? AND member_guild_id = ?",
			session.UserId, userMemberGuild.MemberGuildId).AllCols().Update(*userMemberGuild)
		utils.CheckErr(err)
		if affected == 0 {
			userdata.GenericDatabaseInsert(session, "u_member_guild", *userMemberGuild)
		}
	}
	for _, userMemberGuildSupportItem := range session.UserModel.UserMemberGuildSupportItemById.Map {
		affected, err := session.Db.Table("u_member_guild_support_item").Where("user_id = ? AND support_item_id = ?",
			session.UserId, userMemberGuildSupportItem.SupportItemId).AllCols().Update(*userMemberGuildSupportItem)
		utils.CheckErr(err)
		if affected == 0 {
			userdata.GenericDatabaseInsert(session, "u_member_guild_support_item", *userMemberGuildSupportItem)
		}
	}
}

func init() {
	userdata.AddFinalizer(userMemberGuildFinalizer)
}
