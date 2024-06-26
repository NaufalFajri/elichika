package userdata

import (
	"elichika/client/response"
	"elichika/utils"

	"time"
)

func (session *Session) GetLoginResponse() response.LoginResponse {
	login := response.LoginResponse{}
	exists, err := session.Db.Table("u_login").Where("user_id = ?", session.UserId).Get(&login)
	utils.CheckErr(err)
	if !exists {
		login = response.LoginResponse{
			IsPlatformServiceLinked: true,
			LastTimestamp:           time.Now().UnixMilli(),
			CheckMaintenance:        true,
			FromEea:                 false,
		}
		login.ReproInfo.GroupNo = 1
		GenericDatabaseInsert(session, "u_login", login)
	}
	return login
}

func (session *Session) UpdateLoginData(login response.LoginResponse) {
	affected, err := session.Db.Table("u_login").Where("user_id = ?", session.UserId).AllCols().Update(&login)
	utils.CheckErr(err)
	if affected == 0 {
		GenericDatabaseInsert(session, "u_login", login)
	}
}

func (session *Session) Login() response.LoginResponse {
	// perform a login, load the relevant data into UserModel and load login data into login
	login := session.GetLoginResponse()
	login.UserModel = &session.UserModel

	if session.UserStatus.LastLoginAt < utils.BeginOfDay(session.Time).Unix() { // new day
		session.UserStatus.LoginDays++
	}
	session.UserStatus.LastLoginAt = session.Time.Unix()
	session.SessionType = SessionTypeLogin
	session.SendMissionDetail = true
	session.GenerateNewSessionKey()
	for _, populator := range populators {
		populator(session)
	}
	// only this part is necessary
	login.MemberLovePanels.Slice = session.MemberLovePanels
	return login
}
