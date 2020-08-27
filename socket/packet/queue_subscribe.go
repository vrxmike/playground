package packet

import (
	"encoding/json"

	"github.com/go-redis/redis/v7"
	"github.com/techx/playground/db"
	"github.com/techx/playground/db/models"
)

// sent by hackers and sponsors to subscribe to queue updates
type QueueSubscribePacket struct {
	BasePacket

	SponsorID string `json:"sponsorId"`

	Characters []*models.Character `json:"characters"`
}

func NewQueueSubscribePacket(SponsorID string) *QueueSubscribePacket {
	p := QueueSubscribePacket{}
	p.SponsorID = SponsorID
	
	hackerIDs, _ := db.GetInstance().LRange("sponsor:" + SponsorID + ":hackerqueue", 0 , -1).Result()

	pip := db.GetInstance().Pipeline()
	characterCmds := make([]*redis.StringStringMapCmd, len(hackerIDs))
	
	for i, hackerID := range hackerIDs {
		characterCmds[i] = pip.HGetAll("character:" + hackerID)
	}
	
	pip.Exec()
	characters := make([]*models.Character, len(characterCmds))
	
	for i, characterCmd := range characterCmds {
		characterRes, _ := characterCmd.Result()
		characters[i] = new(models.Character)
		db.Bind(characterRes, characters[i])
	}

	p.Characters = characters
	
	return &p
}

func (p QueueSubscribePacket) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

func (p QueueSubscribePacket) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}
