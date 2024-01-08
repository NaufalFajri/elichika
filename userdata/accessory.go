package userdata

import (
	"elichika/generic"
	"elichika/client"
	"elichika/utils"
)

func (session *Session) GetAllUserAccessories() []client.UserAccessory {
	accessories := []client.UserAccessory{}
	err := session.Db.Table("u_accessory").Where("user_id = ?", session.UserId).
		Find(&accessories)
	utils.CheckErr(err)
	return accessories
}

func (session *Session) GetUserAccessory(userAccessoryId int64) client.UserAccessory {
	// if exist then reuse
	pos, exist := session.UserAccessoryMapping.SetList(&session.UserModel.UserAccessoryByUserAccessoryId).Map[userAccessoryId]
	if exist {
		return session.UserModel.UserAccessoryByUserAccessoryId.Objects[pos]
	}

	// if not look in db
	accessory := client.UserAccessory{}
	exist, err := session.Db.Table("u_accessory").
		Where("user_id = ? AND user_accessory_id = ?", session.UserId, userAccessoryId).Get(&accessory)
	utils.CheckErr(err)
	if !exist {
		// if not exist, create new one
		accessory = client.UserAccessory{
			// UserId:             session.UserId,
			UserAccessoryId:    userAccessoryId,
			Level:              1,
			PassiveSkill1Level: generic.NewNullable(int32(1)),
			PassiveSkill2Level: generic.NewNullable(int32(1)),
			IsNew:              true,
			AcquiredAt:         session.Time.Unix(),
		}
	}
	return accessory
}

func (session *Session) UpdateUserAccessory(accessory client.UserAccessory) {
	session.UserAccessoryMapping.SetList(&session.UserModel.UserAccessoryByUserAccessoryId).Update(accessory)
}

func accessoryFinalizer(session *Session) {
	for _, accessory := range session.UserModel.UserAccessoryByUserAccessoryId.Objects {
		if accessory.IsNull {
			affected, err := session.Db.Table("u_accessory").
				Where("user_id = ? AND user_accessory_id = ?", session.UserId, accessory.UserAccessoryId).
				Delete(&accessory)
			utils.CheckErr(err)
			if affected != 1 {
				panic("accessory doesn't exist")
			}
		} else {
			affected, err := session.Db.Table("u_accessory").
				Where("user_id = ? AND user_accessory_id = ?", session.UserId, accessory.UserAccessoryId).
				AllCols().Update(accessory)
			utils.CheckErr(err)
			if affected == 0 {
				genericDatabaseInsert(session, "u_accessory", accessory)
			}
		}
	}

}

func init() {
	addFinalizer(accessoryFinalizer)
	addGenericTableFieldPopulator("u_accessory", "UserAccessoryByUserAccessoryId")
}
