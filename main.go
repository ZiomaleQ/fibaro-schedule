package main

import (
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
var auth = "Basic "

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
							err = turnOn(settings.Device)

							if err != nil {
								fmt.Println(err)
							}
						}
						lastKnownValue = true
						inSchedule = true
					}
				}
			}
		}

		if lastKnownValue && !inSchedule {
			err = turnOff(settings.Device)

			if err != nil {
				fmt.Println(err)
			}

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

func turnOff(device Device) error {
	return errors.Join(errors.New("Device turn off"), deviceAction(device, device.TurnOff))
}

func turnOn(device Device) error {
	return errors.Join(errors.New("Device turn on"), deviceAction(device, device.TurnOn))
}

func deviceAction(device Device, action string) error {
	payload := strings.NewReader(`{"args":[]}`)

	deviceId := strconv.Itoa(device.Id)

	req, err := http.NewRequest("http://hc3l-00071046.local/api/devices/"+deviceId+"/action/"+action, "POST", payload)

	if err != nil {
		return err
	}

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

type DeviceState struct {
	Properties struct {
		Parameters []struct {
			ID                int     `json:"id"`
			LastReportedValue float64 `json:"lastReportedValue"`
			LastSetValue      float64 `json:"lastSetValue"`
			Size              int     `json:"size"`
			Value             float64 `json:"value"`
		} `json:"parameters"`
	} `json:"properties"`
}
