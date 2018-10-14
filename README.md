Sensirion SHT3x relative humidity and temperature sensor's family
=================================================================

[![Build Status](https://travis-ci.org/d2r2/go-sht3x.svg?branch=master)](https://travis-ci.org/d2r2/go-sht3x)
[![Go Report Card](https://goreportcard.com/badge/github.com/d2r2/go-sht3x)](https://goreportcard.com/report/github.com/d2r2/go-sht3x)
[![GoDoc](https://godoc.org/github.com/d2r2/go-sht3x?status.svg)](https://godoc.org/github.com/d2r2/go-sht3x)
[![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

SHT30, SHT31, SHT35 ([general specification](https://raw.github.com/d2r2/go-sht3x/master/docs/Sensirion_Humidity_Sensors_SHT3x_Datasheet_digital.pdf), [alert mode specification](https://raw.github.com/d2r2/go-sht3x/master/docs/Sensirion_Humidity_Sensors_SHT3x_Application_Note_Alert_Mode_DIS.pdf)) high accuracy temperature and relative humidity sensor. Easily integrated with Arduino and Raspberry PI via i2c communication interface:
![image](https://raw.github.com/d2r2/go-sht3x/master/docs/SHT3X.jpg)

This sensor has extra feature - integrated heater which could be helpfull in some specific application (such as periodic condensate removal, for example).

Here is a library written in [Go programming language](https://golang.org/) for Raspberry PI and counterparts, which gives you in the output relative humidity and temperature values (making all necessary i2c-bus interracting and values computing).

Golang usage
------------


```go
func main() {
	// Create new connection to i2c-bus on 0 line with address 0x44.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(0x44, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer i2c.Close()

	sensor := sht3x.NewSHT3X()
	temp, rh, err := sensor.ReadTemperatureAndHumidity(i2c, sht3x.REPEATABILITY_LOW)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Temperature and relative humidity = %v*C, %v%%", temp, rh)
```


Getting help
------------

GoDoc [documentation](http://godoc.org/github.com/d2r2/go-sht3x)

Installation
------------

```bash
$ go get -u github.com/d2r2/go-sht3x
```

Troubleshoting
--------------

- *How to obtain fresh Golang installation to RPi device (either any RPi clone):*
If your RaspberryPI golang installation taken by default from repository is outdated, you may consider
to install actual golang mannualy from official Golang [site](https://golang.org/dl/). Download
tar.gz file containing armv6l in the name. Follow installation instructions.

- *How to enable I2C bus on RPi device:*
If you employ RaspberryPI, use raspi-config utility to activate i2c-bus on the OS level.
Go to "Interfaceing Options" menu, to active I2C bus.
Probably you will need to reboot to load i2c kernel module.
Finally you should have device like /dev/i2c-1 present in the system.

- *How to find I2C bus allocation and device address:*
Use i2cdetect utility in format "i2cdetect -y X", where X may vary from 0 to 5 or more,
to discover address occupied by peripheral device. To install utility you should run
`apt install i2c-tools` on debian-kind system. `i2cdetect -y 1` sample output:
	```
	     0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
	00:          -- -- -- -- -- -- -- -- -- -- -- -- --
	10: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	20: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	30: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	40: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	50: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	60: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
	70: -- -- -- -- -- -- 76 --    
	```

Contact
-------

Please use [Github issue tracker](https://github.com/d2r2/go-sht3x/issues) for filing bugs or feature requests.


License
-------

Go-sht3x is licensed under MIT License.
