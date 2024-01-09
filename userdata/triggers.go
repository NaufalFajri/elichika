package userdata

import (
	"elichika/client"
	"elichika/utils"
)

func (session *Session) UpdateTriggerCardGradeUp(trigger client.UserInfoTriggerCardGradeUp) {
	session.UserTriggerCardGradeUpMapping.SetList(&session.UserModel.UserInfoTriggerCardGradeUpByTriggerId).Update(trigger)
}

// card grade up trigger is responsible for showing the pop-up animation when openning a card after getting a new copy
// or right after performing a limit break using items
// Getting a new trigger also destroy old trigger, and we might have to update it
func (session *Session) AddTriggerCardGradeUp(trigger client.UserInfoTriggerCardGradeUp) {
	if trigger.TriggerId == 0 {
		trigger.TriggerId = session.Time.UnixNano() + session.UniqueCount
		session.UniqueCount++
	}
	// update the trigger
	session.UpdateTriggerCardGradeUp(trigger)
}

func triggerCardGradeUpFinalizer(session *Session) {
	for _, trigger := range session.UserModel.UserInfoTriggerCardGradeUpByTriggerId.Objects {
		if !trigger.IsNull {
			dbTrigger := client.UserInfoTriggerCardGradeUp{}
			exist, err := session.Db.Table("u_trigger_card_grade_up").
				Where("user_id = ? AND card_master_id = ?", session.UserId, trigger.CardMasterId).Get(&dbTrigger)
			utils.CheckErr(err)
			if exist { // if the card has a trigger, we have to remove it first
				dbTrigger.IsNull = true
				session.UpdateTriggerCardGradeUp(dbTrigger)
				session.Db.Table("u_trigger_card_grade_up").
					Where("user_id = ? AND card_master_id = ?", session.UserId, trigger.CardMasterId).Delete(&dbTrigger)
			}
			trigger.BeforeLoveLevelLimit = trigger.AfterLoveLevelLimit
			// db trigger when login have BeforeLoveLevelLimit = AfterLoveLevelLimit
			// if the 2 numbers are equal the level up isn't shown when we open the card.
			genericDatabaseInsert(session, "u_trigger_card_grade_up", trigger)
		} else {
			// remove from db
			// this is only caused by infoTrigger/read
			_, err := session.Db.Table("u_trigger_card_grade_up").Where("trigger_id = ?", trigger.TriggerId).Delete(
				&client.UserInfoTriggerCardGradeUp{})
			utils.CheckErr(err)
		}
	}
}

func (session *Session) UpdateTriggerBasic(trigger client.UserInfoTriggerBasic) {
	session.UserTriggerBasicMapping.SetList(&session.UserModel.UserInfoTriggerBasicByTriggerId).Update(trigger)
}

func (session *Session) AddTriggerBasic(trigger client.UserInfoTriggerBasic) {
	if trigger.TriggerId == 0 {
		trigger.TriggerId = session.Time.UnixNano() + session.UniqueCount
		session.UniqueCount++
	}
	session.UpdateTriggerBasic(trigger)
}

func triggerBasicFinalizer(session *Session) {
	for _, trigger := range session.UserModel.UserInfoTriggerBasicByTriggerId.Objects {
		if trigger.IsNull { // delete
			_, err := session.Db.Table("u_trigger_basic").Where("trigger_id = ?", trigger.TriggerId).Delete(
				&client.UserInfoTriggerBasic{})
			utils.CheckErr(err)
		} else { // add
			genericDatabaseInsert(session, "u_trigger_basic", trigger)
		}
	}
}

func triggerMemberGuildSupportItemExpiredFinalizer(session *Session) {
	for _, trigger := range session.UserModel.UserInfoTriggerMemberGuildSupportItemExpiredByTriggerId.Objects {
		if trigger.IsNull { // delete
			_, err := session.Db.Table("u_trigger_member_guild_support_item_expired").Where("trigger_id = ?", trigger.TriggerId).Delete(
				&client.UserInfoTriggerMemberGuildSupportItemExpired{})
			utils.CheckErr(err)
		} else { // add
			genericDatabaseInsert(session, "u_trigger_member_guild_support_item_expired", trigger)
		}
	}
}

func (session *Session) ReadMemberGuildSupportItemExpired() {
	err := session.Db.Table("u_trigger_member_guild_support_item_expired").
		Where("user_id = ?", session.UserId).
		Find(&session.UserModel.UserInfoTriggerMemberGuildSupportItemExpiredByTriggerId.Objects)
	utils.CheckErr(err)
	for i := range session.UserModel.UserInfoTriggerMemberGuildSupportItemExpiredByTriggerId.Objects {
		session.UserModel.UserInfoTriggerMemberGuildSupportItemExpiredByTriggerId.Objects[i].IsNull = true
	}
	// already marked as removed, the finalizer will take care of things
	// there's also no need to remove the item, the client won't show them if they're expired
}

// TODO: Trigger member love level up isn't really that persistent, so it's probably better to only keep it in ram
// This could be done by keeping a full user model in ram too.

func (session *Session) UpdateTriggerMemberLoveLevelUp(trigger client.UserInfoTriggerMemberLoveLevelUp) {
	session.UserTriggerMemberLoveLevelUpMapping.SetList(&session.UserModel.UserInfoTriggerMemberLoveLevelUpByTriggerId).Update(trigger)
}

func (session *Session) AddTriggerMemberLoveLevelUp(trigger client.UserInfoTriggerMemberLoveLevelUp) {
	if trigger.TriggerId == 0 {
		trigger.TriggerId = session.Time.UnixNano() + session.UniqueCount
		session.UniqueCount++
	}
	session.UpdateTriggerMemberLoveLevelUp(trigger)
	if !trigger.IsNull {
		genericDatabaseInsert(session, "u_trigger_member_love_level_up", trigger)
	} else {
		_, err := session.Db.Table("u_trigger_member_love_level_up").Where("trigger_id = ?", trigger.TriggerId).Delete(
			&client.UserInfoTriggerMemberLoveLevelUp{})
		utils.CheckErr(err)
	}
}

func (session *Session) ReadAllMemberLoveLevelUpTriggers() {

	err := session.Db.Table("u_trigger_member_love_level_up").
		Where("user_id = ?", session.UserId).Find(&session.UserModel.UserInfoTriggerMemberLoveLevelUpByTriggerId.Objects)
	utils.CheckErr(err)
	for i := range session.UserModel.UserInfoTriggerMemberLoveLevelUpByTriggerId.Objects {
		session.UserModel.UserInfoTriggerMemberLoveLevelUpByTriggerId.Objects[i].IsNull = true
	}
	_, err = session.Db.Table("u_trigger_member_love_level_up").Where("user_id = ?", session.UserId).Delete(
		&client.UserInfoTriggerMemberLoveLevelUp{})
	utils.CheckErr(err)
}

func init() {
	addFinalizer(triggerCardGradeUpFinalizer)
	addFinalizer(triggerBasicFinalizer)
	addFinalizer(triggerMemberGuildSupportItemExpiredFinalizer)
	addGenericTableFieldPopulator("u_trigger_basic", "UserInfoTriggerBasicByTriggerId")
	addGenericTableFieldPopulator("u_trigger_card_grade_up", "UserInfoTriggerCardGradeUpByTriggerId")
	addGenericTableFieldPopulator("u_trigger_member_love_level_up", "UserInfoTriggerMemberLoveLevelUpByTriggerId")
	addGenericTableFieldPopulator("u_trigger_member_guild_support_item_expired", "UserInfoTriggerMemberGuildSupportItemExpiredByTriggerId")
}
