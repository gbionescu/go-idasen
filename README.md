# go-idasen

A project for controlling your Ikea Idasen desk through Bluetooth, using go.

## How do I use it?

First of all, build the project or download it from the releases page.

## How does it work?

It connects to the desk through BLE and stores various configuration data that you may want in `~/.go-idasen.json`.

## Usage

```bash
Usage of ./go-idasen:
  -delfav string
        Remove a given favorite position.
  -desk string
        Set desk by name or address.
  -fav string
        Save current position as named favorite.
  -listfav
        List favorite positions.
  -movefav string
        Load a favorite and move there.
  -pos float
        Position to move desk to in cm. Ranges from 65cm to 128cm.
```

### Connect to a desk

Connect to a desk by specifying the desk name - make sure that you press the connect button on the desk first:

```bash
./go-idasen --desk "Desk name"
```

MAC address also works:

```bash
./go-idasen --desk "00:11:22:33:44:55"
```

### Move the desk to a position

You can specify a position to move the desk by running:

```bash
./go-idasen --pos "80"
```

Positions are limited to minimum `65` and maximum `128`.

### Add a favorite position

You can also save a favorite positions:

```bash
./go-idasen --fav my_fav_pos
```
