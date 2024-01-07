package userdata

import (
	"elichika/utils"
)

func userEventFinalizer(session *Session) {
	for _, userEvent := range session.UserModel.UserEventMarathonByEventMasterId.Objects {
		affected, err := session.Db.Table("u_event_marathon").Where("user_id = ? AND event_master_id = ?",
			session.UserId, userEvent.EventMasterId).AllCols().Update(userEvent)
		utils.CheckErr(err)
		if affected == 0 {
			genericDatabaseInsert(session, "u_event_marathon", userEvent)
		}
	}
	for _, userEvent := range session.UserModel.UserEventMiningByEventMasterId.Objects {
		affected, err := session.Db.Table("u_event_mining").Where("user_id = ? AND event_master_id = ?",
			session.UserId, userEvent.EventMasterId).AllCols().Update(userEvent)
		utils.CheckErr(err)
		if affected == 0 {
			genericDatabaseInsert(session, "u_event_mining", userEvent)
		}
	}
	for _, userEvent := range session.UserModel.UserEventCoopByEventMasterId.Objects {
		affected, err := session.Db.Table("u_event_coop").Where("user_id = ? AND event_master_id = ?",
			session.UserId, userEvent.EventMasterId).AllCols().Update(userEvent)
		utils.CheckErr(err)
		if affected == 0 {
			genericDatabaseInsert(session, "u_event_coop", userEvent)
		}
	}
}
func init() {
	addFinalizer(userEventFinalizer)

	addGenericTableFieldPopulator("u_event_marathon", "UserEventMarathonByEventMasterId")
	addGenericTableFieldPopulator("u_event_mining", "UserEventMiningByEventMasterId")
	addGenericTableFieldPopulator("u_event_coop", "UserEventCoopByEventMasterId")
}
