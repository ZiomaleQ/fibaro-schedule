package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var client = &http.Client{}

func main() {
	data, err := os.ReadFile("settings.json")

	if err != nil {
		fmt.Println(err)
		return
	}

	var settings Settings

	err = json.Unmarshal(data, &settings)

	if err != nil {
		fmt.Println(err)
		return
	}

	var lastKnownValue bool = false

	for {
		now := time.Now()
		var inSchedule bool

		for _, schedule := range settings.Schedules {
			if schedule.From <= now.Format("15:04") && schedule.To > now.Format("15:04") {
				for _, day := range schedule.Every {
					if Weekday[day] == now.Weekday() {
						if !lastKnownValue {
							err = turnOn(settings)

							if err != nil {
								panic(err)
							}

							fmt.Println("Turned on")
						}
						lastKnownValue = true
						inSchedule = true
					}
				}
			}
		}

		if lastKnownValue && !inSchedule {
			err = turnOff(settings)

			if err != nil {
				panic(err)
			}

			fmt.Println("Turned off")

			lastKnownValue = false
		}

		time.Sleep(1 * time.Minute)
	}
}

type Settings struct {
	Device    Device     `json:"Device"`
	Schedules []Schedule `json:"Schedules"`
	User      User       `json:"User"`
}

type Device struct {
	Id      int    `json:"Id"`
	TurnOn  string `json:"TurnOn"`
	TurnOff string `json:"TurnOff"`
}

type Schedule struct {
	From  string   `json:"From"`
	To    string   `json:"To"`
	Every []string `json:"Every"`
}

type User struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
}

var Weekday = map[string]time.Weekday{
	"Monday":    time.Monday,
	"Tuesday":   time.Tuesday,
	"Wednesday": time.Wednesday,
	"Thursday":  time.Thursday,
	"Friday":    time.Friday,
	"Saturday":  time.Saturday,
	"Sunday":    time.Sunday,
}

func turnOff(settings Settings) error {
	err := deviceAction(settings, settings.Device.TurnOff)

	if err == nil {
		return nil
	} else {
		return errors.Join(errors.New("Device turn off"), err)
	}
}

func turnOn(settings Settings) error {

	err := deviceAction(settings, settings.Device.TurnOn)

	if err == nil {
		return nil
	} else {
		return errors.Join(errors.New("Device turn on"), err)
	}
}

func deviceAction(settings Settings, action string) error {
	payload := strings.NewReader(`{"args":[]}`)

	deviceId := strconv.Itoa(settings.Device.Id)

	req, err := http.NewRequest("POST", "http://192.168.1.37/api/devices/"+deviceId+"/action/"+action, payload)

	if err != nil {
		return err
	}

	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(settings.User.Email+":"+settings.User.Password))

	req.Header.Add("accept", "*/*")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Fibaro-Version", "2")
	req.Header.Add("Accept-language", "pl")
	req.Header.Add("Authorization", auth)

	_, err = client.Do(req)

	if err != nil {
		return err
	}

	return nil
}
