package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/ble"
)

const baseHeight = 63.00
const positionDiff = 1.0

const deskBLEPosition = "99fa0021-338a-1024-8a49-009c0215f78a"
const deskBleControl = "99fa0002-338a-1024-8a49-009c0215f78a"

var deskBLEUp = uint16(71)
var deskBLEDown = uint16(70)

// BLE desk driver stucture that plugs in to gobot
type deskDriver struct {
	name       string
	connection gobot.Connection
	gobot.Eventer
}

// Create a new Idasen BLE driver
func newDeskDriver(a ble.BLEConnector) *deskDriver {
	n := &deskDriver{
		name:       gobot.DefaultName("IdasenDriver"),
		connection: a,
		Eventer:    gobot.NewEventer(),
	}

	return n
}

// Connection returns the Driver's Connection to the associated Adaptor
func (b *deskDriver) Connection() gobot.Connection { return b.connection }

// Name returns the Driver name
func (b *deskDriver) Name() string { return b.name }

// SetName sets the Driver name
func (b *deskDriver) SetName(n string) { b.name = n }

// adaptor returns BLE adaptor
func (b *deskDriver) adaptor() ble.BLEConnector {
	return b.Connection().(ble.BLEConnector)
}

// Start tells driver to get ready to do work
func (b *deskDriver) Start() (err error) {
	return
}

// Halt stops battery driver (void)
func (b *deskDriver) Halt() (err error) { return }

// Gets current desk position and returns a float between 65 and 128
func (b *deskDriver) getPosition() (level float64) {
	c, err := b.adaptor().ReadCharacteristic(deskBLEPosition)
	if err != nil {
		log.Println(err)
		return
	}

	return baseHeight + float64(binary.LittleEndian.Uint16(c))/100
}

// Moves desk up by sending a BLE command
// Does not have any control on how much to move
func (b *deskDriver) moveUp() {
	moveCmd := make([]byte, 2)
	binary.LittleEndian.PutUint16(moveCmd, deskBLEUp)

	err := b.adaptor().WriteCharacteristic(deskBleControl, moveCmd)
	if err != nil {
		log.Println(err)
		return
	}
}

// Moves desk down by sending a BLE command
// Does not have any control on how much to move
func (b *deskDriver) moveDown() {
	moveCmd := make([]byte, 2)
	binary.LittleEndian.PutUint16(moveCmd, deskBLEDown)

	err := b.adaptor().WriteCharacteristic(deskBleControl, moveCmd)
	if err != nil {
		log.Println(err)
		return
	}
}

// Moves the desk to a given position
func (b *deskDriver) move(position float64) {
	for {
		crtPosition := b.getPosition()
		// If current position is within range, break - we're done
		if math.Abs(crtPosition-position) <= positionDiff {
			fmt.Println("BYE")
			return
		} else if crtPosition < position {
			b.moveUp()
		} else if crtPosition >= position {
			b.moveDown()
		}
	}
}
