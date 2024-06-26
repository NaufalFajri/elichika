package user_live_deck

import (
	"elichika/client"
	"elichika/userdata"
)

func InsertUserLiveDecks(session *userdata.Session, decks []client.UserLiveDeck) {
	for _, deck := range decks {
		UpdateUserLiveDeck(session, deck)
	}
}
