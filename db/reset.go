package db

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/google/uuid"
)

// RoomType is an enum representing all possible room templates
type RoomType string

const (
	// Home is the room that everyone spawns in, otherwise known as town square
	Home RoomType = "home"

	// Plaza is the room where you can get to the coffee shop, arcade, and stadium
	Plaza = "plaza"

	// Nightclub is the club, accessible from town square
	Nightclub = "nightclub"

	// CoffeeShop is the coffee shop, accessible from plaza
	CoffeeShop = "coffee_shop"

	// Nonprofits is the campground with all of the nonprofit tents
	Nonprofits = "nonprofits"

	// Personal is a template for someone's personal room
	Personal = "personal"

	// PlatArea is the area accessible from town square with the two plat sponsor buildings
	PlatArea = "plat_area"

	// LeftField is the left sponsor area
	LeftField = "left_field"

	// RightField is the right sponsor area
	RightField = "right_field"

	// Plat is a plat-tier sponsor's room
	Plat = "plat"

	// Gold is a gold-tier sponsor's room
	Gold = "gold"

	// Silver is a silver-tier sponsor's room
	Silver = "silver"

	// Bronze is a bronze-tier sponsor's room
	Bronze = "bronze"
)

// CreateRoom builds a room with the given ID from a template file
func createRoomWithData(id string, roomType RoomType, data map[string]interface{}) {
	dat, err := ioutil.ReadFile("config/rooms/" + string(roomType) + ".json")

	if err != nil {
		return
	}

	var roomData map[string]interface{}
	json.Unmarshal(dat, &roomData)
	data["background"] = roomData["background"]

	if sponsorID, ok := data["sponsorId"].(string); ok {
		data["background"] = strings.ReplaceAll(data["background"].(string), "<sponsor>", sponsorID)
	}

	instance.HSet("room:"+id, data)

	elements := roomData["elements"].([]interface{})

	// If this is the nightclub, add floor tiles
	if id == "nightclub" {
		tileStartX := 0.374
		tileStartY := 0.552
		tileSeparator := 0.0305
		numTilesX := 7
		numTilesY := 4

		newTiles := make([]interface{}, numTilesX*numTilesY)

		for i := 0; i < numTilesY; i++ {
			for j := 0; j < numTilesX; j++ {
				newTiles[i*numTilesX+j] = map[string]interface{}{
					"x":    tileStartX + float64(i+j)*tileSeparator,
					"y":    tileStartY + float64((numTilesY-i)+j)*tileSeparator,
					"tile": true,
				}
			}
		}

		elements = append(newTiles, elements...)
	}

	for _, val := range elements {
		elementID := uuid.New().String()
		elementData := val.(map[string]interface{})

		if _, ok := elementData["tile"]; ok {
			// If this is a nightclub floor tile, autofill some attributes
			delete(elementData, "tile")
			elementData["width"] = 0.052
			elementData["path"] = "tiles/blue1.svg"
			elementData["changingImagePath"] = true
			elementData["changingPaths"] = "tiles/blue1.svg,tiles/blue2.svg,tiles/blue3.svg,tiles/blue4.svg,tiles/green1.svg,tiles/green2.svg,tiles/pink1.svg,tiles/pink2.svg,tiles/pink3.svg,tiles/pink4.svg,tiles/yellow1.svg"
			elementData["changingInterval"] = 2000
			elementData["changingRandomly"] = true
		}

		if _, ok := elementData["campfire"]; ok {
			// If this is a campfire, animate it
			delete(elementData, "campfire")
			elementData["width"] = 0.0253
			elementData["path"] = "campfire/campfire1.svg"
			elementData["changingImagePath"] = true
			elementData["changingPaths"] = "campfire/campfire1.svg,campfire/campfire2.svg,campfire/campfire3.svg,campfire/campfire4.svg,campfire/campfire5.svg"
			elementData["changingInterval"] = 250
			elementData["changingRandomly"] = false
		}

		if _, ok := elementData["fountain"]; ok {
			// If this is a fountain, animate it
			delete(elementData, "fountain")
			elementData["path"] = "fountain1.svg"
			elementData["changingImagePath"] = true
			elementData["changingPaths"] = "fountain1.svg,fountain2.svg,fountain3.svg"
			elementData["changingInterval"] = 1000
			elementData["changingRandomly"] = false
		}

		if _, ok := elementData["toggleable"]; ok {
			switch elementData["path"] {
			case "street_lamp.svg":
				elementData["path"] = "street_lamp.svg,street_lamp_off.svg"
			case "bar_closed.svg":
				elementData["path"] = "bar_closed.svg,bar_open.svg"
			case "flashlight_off.svg":
				elementData["path"] = "flashlight_off.svg,flashlight_on.svg"
			default:
				break
			}

			elementData["state"] = 0
		}

		if sponsorID, ok := data["sponsorId"].(string); ok {
			elementData["path"] = strings.ReplaceAll(elementData["path"].(string), "<sponsor>", sponsorID)
		}

		instance.HSet("element:"+elementID, elementData)
		instance.RPush("room:"+id+":elements", elementID)
	}

	for _, val := range roomData["hallways"].([]interface{}) {
		hallwayData := val.(map[string]interface{})

		if _, ok := data["sponsorId"].(string); ok {
			hallwayData["toX"] = data["toX"].(float64)
			hallwayData["toY"] = data["toY"].(float64)

			if val, ok := data["to"].(string); ok {
				hallwayData["to"] = val
			}
		}

		hallwayID := uuid.New().String()
		instance.HSet("hallway:"+hallwayID, val)
		instance.SAdd("room:"+id+":hallways", hallwayID)
	}

	instance.SAdd("rooms", id)
}

func createSponsors() {
	dat, err := ioutil.ReadFile("config/sponsors.json")

	if err != nil {
		return
	}

	var sponsorsData []map[string]string
	json.Unmarshal(dat, &sponsorsData)

	for _, sponsor := range sponsorsData {
		instance.HSet("sponsor:"+sponsor["id"], map[string]interface{}{
			"name": sponsor["name"],
			"zoom": sponsor["zoom"],
		})

		instance.SAdd("sponsors", sponsor["id"])
	}
}

func CreateRoom(id string, roomType RoomType) {
	createRoomWithData(id, roomType, map[string]interface{}{})
}

func createEvents() {
	dat, err := ioutil.ReadFile("config/events.json")

	if err != nil {
		return
	}

	var eventsData []map[string]interface{}
	json.Unmarshal(dat, &eventsData)

	for _, event := range eventsData {
		eventID := uuid.New().String()[:4]
		instance.HSet("event:"+eventID, event)
		instance.SAdd("events", eventID)
	}
}

func reset() {
	instance.FlushDB()
	CreateRoom("home", Home)
	CreateRoom("nightclub", Nightclub)
	CreateRoom("nonprofits", Nonprofits)
	CreateRoom("plat_area", PlatArea)
	CreateRoom("left_field", LeftField)
	CreateRoom("right_field", RightField)
	CreateRoom("plaza", Plaza)
	CreateRoom("coffee_shop", CoffeeShop)

	createRoomWithData("sponsor:cmt", Plat, map[string]interface{}{
		"sponsorId": "cmt",
		"toX":       0.2685,
		"toY":       0.5919,
	})

	createRoomWithData("sponsor:intersystems", Plat, map[string]interface{}{
		"sponsorId": "intersystems",
		"toX":       0.7402,
		"toY":       0.5717,
	})

	createRoomWithData("sponsor:drw", Gold, map[string]interface{}{
		"sponsorId": "drw",
		"to":        "left_field",
		"toX":       0.8215,
		"toY":       0.4943,
	})

	createRoomWithData("sponsor:yext", Gold, map[string]interface{}{
		"sponsorId": "yext",
		"to":        "left_field",
		"toX":       0.6128,
		"toY":       0.702,
	})

	createRoomWithData("sponsor:facebook", Silver, map[string]interface{}{
		"sponsorId": "facebook",
		"to":        "left_field",
		"toX":       0.3211,
		"toY":       0.7636,
	})

	createRoomWithData("sponsor:arrowstreet", Silver, map[string]interface{}{
		"sponsorId": "arrowstreet",
		"to":        "left_field",
		"toX":       0.2018,
		"toY":       0.6347,
	})

	createRoomWithData("sponsor:oca", Bronze, map[string]interface{}{
		"sponsorId": "oca",
		"to":        "left_field",
		"toX":       0.1148,
		"toY":       0.5487,
	})

	createRoomWithData("sponsor:pega", Bronze, map[string]interface{}{
		"sponsorId": "pega",
		"to":        "left_field",
		"toX":       0.0487,
		"toY":       0.4728,
	})

	createRoomWithData("sponsor:ibm", Gold, map[string]interface{}{
		"sponsorId": "ibm",
		"to":        "right_field",
		"toX":       0.1792,
		"toY":       0.5072,
	})

	createRoomWithData("sponsor:nasdaq", Gold, map[string]interface{}{
		"sponsorId": "nasdaq",
		"to":        "right_field",
		"toX":       0.3871,
		"toY":       0.712,
	})

	createRoomWithData("sponsor:citadel", Silver, map[string]interface{}{
		"sponsorId": "citadel",
		"to":        "right_field",
		"toX":       0.6788,
		"toY":       0.755,
	})

	createRoomWithData("sponsor:goldman", Silver, map[string]interface{}{
		"sponsorId": "goldman",
		"to":        "right_field",
		"toX":       0.7916,
		"toY":       0.6347,
	})

	createRoomWithData("sponsor:linode", Silver, map[string]interface{}{
		"sponsorId": "linode",
		"to":        "right_field",
		"toX":       0.8867,
		"toY":       0.543,
	})

	createRoomWithData("sponsor:quantco", Bronze, map[string]interface{}{
		"sponsorId": "quantco",
		"to":        "right_field",
		"toX":       0.9593,
		"toY":       0.4656,
	})

	createEvents()
	createSponsors()

	instance.SAdd("sponsor_emails", "cookj@mit.edu")
	instance.HSet("emailToSponsor", "cookj@mit.edu", "cmt")
}
