package serverdb

import (
	"elichika/model"

	"encoding/json"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

func (session *Session) GetLiveDifficultyRecord(liveDifficultyID int) model.LifeDifficultyRecord {
	record := model.LifeDifficultyRecord{}
	exists, err := Engine.Table("s_user_live_record").
		Where("user_id = ? AND live_difficulty_id = ?", session.UserStatus.UserID, liveDifficultyID).
		Get(&record)
	if err != nil {
		panic(err)
	}
	if !exists {
		record.UserID = UserID
		record.LiveDifficultyID = liveDifficultyID
		record.EnableAutoplay = true
		record.IsNew = true
	}
	return record
}


func (session *Session) InsertOrUpdateLiveDifficultyRecord(record model.LifeDifficultyRecord)  {
	inserted, err := Engine.Table("s_user_live_record").
		Where("user_id = ? AND live_difficulty_id = ?", record.UserID, record.LiveDifficultyID).
		AllCols().Update(&record)
	if err != nil {
		panic(err)
	}

	if inserted == 0 { // need to insert
		inserted, err = Engine.Table("s_user_live_record").Insert(record)
		if err != nil {
			panic(err)
		}
		if inserted == 0 {
			panic("failed to insert live record")
		}
	}
}

func (session *Session) GetLastPlayLiveDifficultyDeck(liveDifficultyID int) *model.LastPlayLiveDifficultyDeck {
	lastPlayDeck := model.LastPlayLiveDifficultyDeck{}
	exists, err := Engine.Table("s_user_live_record").
		Where("user_id = ? AND live_difficulty_id = ?", session.UserStatus.UserID, liveDifficultyID).
		Get(&lastPlayDeck)
	if err != nil {
		panic(err)
	}
	if !exists {
		return nil
	}
	return &lastPlayDeck
}

func (session *Session) BuildLastPlayLiveDifficultyDeck(deckID, liveDifficultyID int) model.LastPlayLiveDifficultyDeck {
	lastPlayDeck := model.LastPlayLiveDifficultyDeck{
		UserID:           session.UserStatus.UserID,
		LiveDifficultyID: liveDifficultyID,
		Voltage:          0,     // filled by handler
		IsCleared:        false, // filled by handler
		RecordedAt:       time.Now().Unix()}
	lastPlayDeck.CardWithSuitDict = make([]int, 18)
	userLiveDeck := session.GetUserLiveDeck(deckID)
	userLiveDeckJson, _ := json.Marshal(userLiveDeck)

	gjson.Parse(string(userLiveDeckJson)).ForEach(func(key, value gjson.Result) bool {
		k := key.String()
		if strings.Contains(k, "_master_id_") {
			id := int(k[len(k)-1] - '0')
			if strings.Contains(k, "card_master_id") {
				lastPlayDeck.CardWithSuitDict[id*2-2] = int(value.Int())
			} else {
				lastPlayDeck.CardWithSuitDict[id*2-1] = int(value.Int())
			}
		}
		return true
	})

	liveParties := session.GetUserLivePartiesWithDeckID(deckID)
	for i, party := range liveParties {
		squad := model.DeckSquadDict{}
		partyJson, _ := json.Marshal(party)
		gjson.Parse(string(partyJson)).ForEach(func(key, value gjson.Result) bool {
			k := key.String()
			if strings.Contains(k, "card_master_id_") {
				squad.CardMasterIDs = append(squad.CardMasterIDs, int(value.Int()))
			} else if strings.Contains(k, "user_accessory_id_") {
				ptr := new(int64)
				*ptr = value.Int()
				squad.UserAccessoryIDs = append(squad.UserAccessoryIDs, ptr)
			}
			return true
		})
		lastPlayDeck.SquadDict = append(lastPlayDeck.SquadDict, i)
		lastPlayDeck.SquadDict = append(lastPlayDeck.SquadDict, squad)
	}

	return lastPlayDeck
}

func (session *Session) SetLastPlayLiveDifficultyDeck(deck model.LastPlayLiveDifficultyDeck) {
	// always call after inserting the actual live play, so we can just update
	_, err := Engine.Table("s_user_live_record").Where("user_id = ? and live_difficulty_id = ?", deck.UserID, deck.LiveDifficultyID).
		AllCols().Update(&deck)
	if err != nil {
		panic(err)
	}
}