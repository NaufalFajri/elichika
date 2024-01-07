package login_bonus

import (
	"elichika/enum"
	"elichika/gamedata"
	"elichika/model"
	"elichika/userdata"

	"math/rand"
	"time"
)

func birthdayLoginBonusHandler(mode string, session *userdata.Session, loginBonus *gamedata.LoginBonus, target *model.BootstrapLoginBonus) {
	if loginBonus.LoginBonusType != enum.LoginBonusTypeBirthday {
		panic("wrong handler used")
	}
	year, month, day := session.Time.Date()
	mmdd := int(month)*100 + int(day)
	list, exists := session.Gamedata.MemberByBirthday[mmdd]
	if !exists { // no one with this birthday
		return
	}
	userLoginBonus := session.GetUserLoginBonus(loginBonus.LoginBonusId)
	lastUnlocked := time.Date(year, month, day, 0, 0, 0, 0, session.Time.Location())
	if userLoginBonus.LastReceivedAt >= lastUnlocked.Unix() { // already got it
		return
	}
	userLoginBonus.LastReceivedAt = session.Time.Unix()

	for _, member := range list {
		// the present is as follow:
		// - 50 gems
		// - 2 memento
		// - 50 memorial
		// - additional 25 memorial for channel member
		naviLoginBonus := loginBonus.NaviLoginBonus()
		naviLoginBonus.LoginBonusRewards = append(naviLoginBonus.LoginBonusRewards,
			model.LoginBonusRewards{
				Day:          1,
				Status:       enum.LoginBonusReceiveStatusReceiving,
				ContentGrade: enum.LoginBonusContentGradeRare,
				LoginBonusContents: []model.Content{
					model.Content{
						ContentType:   enum.ContentTypeTrainingMaterial,
						ContentID:     18000 + member.ID,
						ContentAmount: 2,
					},
					model.Content{
						ContentType:   enum.ContentTypeTrainingMaterial,
						ContentID:     8000 + member.ID,
						ContentAmount: 50,
					},
					model.Content{
						ContentType:   enum.ContentTypeSnsCoin,
						ContentID:     0,
						ContentAmount: 50,
					},
				},
			},
		)
		if session.UserStatus.MemberGuildMemberMasterID == member.ID {
			naviLoginBonus.LoginBonusRewards[0].LoginBonusContents[1].ContentAmount += 25
		}

		for _, content := range naviLoginBonus.LoginBonusRewards[0].LoginBonusContents {
			// TODO(present_box): This correctly has to go to the present box, but we just do it here
			session.AddResource(content)
		}

		// choose the background and the costume
		memberLoginBonusBirthday := member.MemberLoginBonusBirthdays[0]
		switch mode {
		case "random":
			memberLoginBonusBirthday = member.MemberLoginBonusBirthdays[rand.Intn(len(member.MemberLoginBonusBirthdays))]
		case "latest":
		default:
			panic("not supported")
		}
		target.BirthdayMember = append(target.BirthdayMember, model.LoginBonusBirthDayMember{
			MemberMasterId: member.ID,
			SuitMasterId:   memberLoginBonusBirthday.SuitMasterId,
		})
		naviLoginBonus.BackgroundId = memberLoginBonusBirthday.Id
		target.BirthdayLoginBonuses = append(target.BirthdayLoginBonuses, naviLoginBonus)
	}
	session.UpdateUserLoginBonus(userLoginBonus)
}