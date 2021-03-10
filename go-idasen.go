package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"encoding/json"

	"gobot.io/x/gobot/platforms/ble"
	"tinygo.org/x/bluetooth"
)

const dataPath = "/.go-idasen.json"
const connectTimeout = 10 * time.Second
const deskMinHeight = 65.00
const deskMaxHeight = 128.0

type deskData struct {
	Name      string             `json:Name`
	Positions map[string]float64 `json:positions`
}

// Returns current desk name
func (desk *deskData) name() string {
	return desk.Name
}

// Sets current desk name
func (desk *deskData) setName(name string) {
	desk.Name = name
	desk.saveSettings()
}

// Returns a formatted list of fav positions
func (desk *deskData) listPositions() string {
	posData := ""
	if desk.Positions != nil {
		for name, position := range desk.Positions {
			posData += fmt.Sprintf("\t%s: %.2f", name, position)
		}
	} else {
		return "No favorite positions saved."
	}

	return posData
}

// Save settings to json
func (desk *deskData) saveSettings() {
	data, err := json.MarshalIndent(desk, "", "    ")
	if err != nil {
		printExit("Could not convert data to json: "+err.Error(), 2)
	}

	err = ioutil.WriteFile(getSettingsPath(), data, 0644)
	if err != nil {
		printExit("Could not write data to file "+err.Error(), 2)
	}
}

// Add a favorite position by name and height
func (desk *deskData) addFav(position float64, name string) {
	desk.Positions[name] = position
	desk.saveSettings()
}

// Remove a named position
func (desk *deskData) delFav(name string) bool {
	_, present := desk.Positions[name]
	if present {
		delete(desk.Positions, name)
	}

	desk.saveSettings()
	return present
}

// Get height for a favorite position
func (desk *deskData) getFav(name string) (float64, error) {
	value, present := desk.Positions[name]
	if present {
		return value, nil
	}

	return 0, nil
}

// Returns where data is saved
func getSettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		printExit("Error getting home path "+err.Error(), 2)
	}

	return home + dataPath
}

// Loads the entire settings struct
// If none found, it returns an empty one
func loadSettings() (deskData, error) {
	settings := getSettingsPath()
	emptyData := deskData{"", make(map[string]float64)}

	if _, err := os.Stat(settings); os.IsNotExist(err) {
		return emptyData, nil
	}

	settingsFile, err := os.Open(settings)
	if err != nil {
		return emptyData, err
	}

	content, err := ioutil.ReadAll(settingsFile)
	if err != nil {
		return emptyData, err
	}

	var settingsContent deskData
	err = json.Unmarshal(content, &settingsContent)
	if err != nil {
		return emptyData, err
	}

	return settingsContent, nil
}

// Prints a message and quits with given exit code
func printExit(message string, exitCode int) {
	if exitCode == 0 {
		fmt.Println(message + "\n")
	} else {
		fmt.Fprintf(os.Stderr, message+"\n")
	}
	os.Exit(exitCode)
}

// Connects to the desk by name or mac address
func getDesk(nameOrAddr string) *deskDriver {
	// Get the default adapter
	currentAdapter := bluetooth.DefaultAdapter
	err := currentAdapter.Enable()
	if err != nil {
		fmt.Errorf("Can't get bluetooth device")
	}
	fmt.Println("Trying to connect to " + nameOrAddr)

	var address string = ""
	// Scan all devices and connect to the given one
	// Quit if not found within 'timeout' seconds
	go func() {
		time.Sleep(connectTimeout)
		if address == "" {
			printExit("could not find device", 2)
		}
	}()
	currentAdapter.Scan(
		func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			if nameOrAddr == result.Address.String() || nameOrAddr == result.LocalName() {
				address = result.Address.String()
				adapter.StopScan()
			}
		})

	bleAdaptor := ble.NewClientAdaptor(address)
	bleDesk := newDeskDriver(bleAdaptor)

	bleAdaptor.Connect()
	bleDesk.Start()

	return bleDesk
}

func main() {
	var desk = flag.String("desk", "", "Set desk by name or address.")
	var pos = flag.Float64("pos", 0, "Position to move desk to in cm. Ranges from 65cm to 128cm.")
	var fav = flag.String("fav", "", "Save current position as named favorite.")
	var movefav = flag.String("movefav", "", "Load a favorite and move there.")
	var listfav = flag.Bool("listfav", false, "List favorite positions.")
	var delfav = flag.String("delfav", "", "Remove a given favorite position.")
	flag.Parse()

	settings, err := loadSettings()
	if err != nil {
		printExit("Error loading settings file "+err.Error(), 2)
	}

	if *listfav {
		printExit(settings.listPositions(), 0)
	}

	if *delfav != "" {
		found := settings.delFav(*delfav)
		if found {
			printExit("Position deleted.", 0)
		} else {
			printExit("Given position does not exist.", 2)
		}
	}

	// Try to figure out if we should connect to a given desk or the stored one
	target := ""
	if *desk != "" {
		target = *desk
	} else if settings.name() != "" {
		target = settings.name()
	} else {
		printExit("No desk name or address specified. Please add `--desk` to specify what desk to connect to.", 2)
	}

	// Update with the new name either way
	settings.setName(target)

	// Get a connection to the desk
	bleDesk := getDesk(target)

	if *pos != 0 {
		if *pos >= deskMinHeight && *pos <= deskMaxHeight {
			bleDesk.move(*pos)
		} else {
			printExit("Position must be between 65 and 128.", 2)
		}
	}

	if *fav != "" {
		settings.addFav(bleDesk.getPosition(), *fav)
		printExit("Saved current position to "+*fav, 0)
	}

	if *movefav != "" {
		position, err := settings.getFav(*movefav)
		if err != nil {
			printExit("No such favorite "+*movefav, 2)
		}
		bleDesk.move(position)
	}
}
