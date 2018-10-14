package main

import (
	"context"
	"time"

	i2c "github.com/d2r2/go-i2c"
	logger "github.com/d2r2/go-logger"
	shell "github.com/d2r2/go-shell"
	sht3x "github.com/d2r2/go-sht3x"
)

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

func main() {
	defer logger.FinalizeLogger()
	// Create new connection to i2c-bus on 0 line with address 0x44.
	// Use i2cdetect utility to find device address over the i2c-bus
	i2c, err := i2c.NewI2C(0x44, 0)
	if err != nil {
		lg.Fatal(err)
	}
	defer i2c.Close()

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** !!! READ THIS !!!")
	lg.Notify("*** You can change verbosity of output, by modifying logging level of modules \"i2c\", \"sht3x\".")
	lg.Notify("*** Uncomment/comment corresponding lines with call to ChangePackageLogLevel(...)")
	lg.Notify("*** !!! READ THIS !!!")
	lg.Notify("**********************************************************************************************")
	// Uncomment/comment next line to suppress/increase verbosity of output
	// logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
	// logger.ChangePackageLogLevel("sht3x", logger.InfoLevel)

	sensor := sht3x.NewSHT3X()
	// Clear sensor settings
	err = sensor.Reset(i2c)
	if err != nil {
		lg.Fatal(err)
	}

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Read sensor states")
	lg.Notify("**********************************************************************************************")
	hs, err := sensor.GetHeaterStatus(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Heater ON status = %v", hs)
	aps, err := sensor.GetAlertPendingStatus(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Alert pending status = %v", aps)
	tas, err := sensor.GetTemperatureAlertStatus(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Temperature alert pending status = %v", tas)
	has, err := sensor.GetHumidityAlertStatus(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Humidity alert pending status = %v", has)
	rsd, err := sensor.CheckResetDetected(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Reset status detected = %v", rsd)

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Single shot measurement mode")
	lg.Notify("**********************************************************************************************")
	ut, urh, err := sensor.ReadUncompTemperatureAndHumidity(i2c, sht3x.REPEATABILITY_MEDIUM)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Temprature and RH uncompensated = %v, %v", ut, urh)

	temp, rh, err := sensor.ReadTemperatureAndRelativeHumidity(i2c, sht3x.REPEATABILITY_LOW)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Temperature and relative humidity = %v*C, %v%%", temp, rh)

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Periodic data acquisition mode ")
	lg.Notify("**********************************************************************************************")
	period := sht3x.PERIODIC_4MPS
	err = sensor.StartPeriodicTemperatureAndHumidityMeasure(i2c, period, sht3x.REPEATABILITY_LOW)
	if err != nil {
		lg.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		temp, rh, err := sensor.FetchTemperatureAndRelativeHumidityWithContext(context.Background(), i2c, period)
		if err != nil {
			lg.Fatal(err)
		}
		lg.Infof("Temperature and relative humidity = %v*C, %v%%", temp, rh)
	}
	err = sensor.Break(i2c)
	if err != nil {
		lg.Fatal(err)
	}

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Get temperature and humidity alert limits")
	lg.Notify("**********************************************************************************************")

	temp, rh, err = sensor.ReadAlertHighSet(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Read alert HIGH SET limit for temperature and relative humidity = %v*C, %v%%", temp, rh)
	temp, rh, err = sensor.ReadAlertHighClear(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Read alert HIGH CLEAR limit for temperature and relative humidity = %v*C, %v%%", temp, rh)
	temp, rh, err = sensor.ReadAlertLowClear(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Read alert LOW CLEAR limit for temperature and relative humidity = %v*C, %v%%", temp, rh)
	temp, rh, err = sensor.ReadAlertLowSet(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Read alert LOW SET limit for temperature and relative humidity = %v*C, %v%%", temp, rh)

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Set temperature and humidity alert limits.")
	lg.Notify("*** Equation must be respected: HIGH SET > HIGH CLEAR > LOW CLEAR > LOW SET")
	lg.Notify("**********************************************************************************************")

	err = sensor.WriteAlertHighSet(i2c, 110, 90)
	if err != nil {
		lg.Fatal(err)
	}
	err = sensor.WriteAlertHighClear(i2c, 108, 88)
	if err != nil {
		lg.Fatal(err)
	}
	err = sensor.WriteAlertLowClear(i2c, -18, 10)
	if err != nil {
		lg.Fatal(err)
	}
	err = sensor.WriteAlertLowSet(i2c, -20, 8)
	if err != nil {
		lg.Fatal(err)
	}
	temp, rh, err = sensor.ReadAlertHighSet(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Read alert HIGH SET limit for temperature and relative humidity = %v*C, %v%%", temp, rh)
	temp, rh, err = sensor.ReadAlertHighClear(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Read alert HIGH CLEAR limit for temperature and relative humidity = %v*C, %v%%", temp, rh)
	temp, rh, err = sensor.ReadAlertLowClear(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Read alert LOW CLEAR limit for temperature and relative humidity = %v*C, %v%%", temp, rh)
	temp, rh, err = sensor.ReadAlertLowSet(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Read alert LOW SET limit for temperature and relative humidity = %v*C, %v%%", temp, rh)

	lg.Notify("**********************************************************************************************")
	lg.Notify("*** Activate heater for 10 secs and make a measurement ")
	lg.Notify("**********************************************************************************************")
	done := make(chan struct{})
	defer close(done)
	// Create context with cancelation possibility.
	ctx, cancel := context.WithCancel(context.Background())
	// Run goroutine waiting for OS termantion events, including keyboard Ctrl+C.
	shell.CloseContextOnKillSignal(cancel, done)

	err = sensor.SetHeaterStatus(i2c, true)
	if err != nil {
		lg.Fatal(err)
	}
	hs, err = sensor.GetHeaterStatus(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Heater ON status = %v", hs)
	pause := time.Second * 10
	lg.Infof("Waiting %v...", pause)
	select {
	// Check for termination request.
	case <-ctx.Done():
		err = sensor.SetHeaterStatus(i2c, false)
		if err != nil {
			lg.Fatal(err)
		}
		lg.Fatal(ctx.Err())
	// Sleep 10 sec.
	case <-time.After(pause):
	}
	err = sensor.SetHeaterStatus(i2c, false)
	if err != nil {
		lg.Fatal(err)
	}
	hs, err = sensor.GetHeaterStatus(i2c)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Heater ON status = %v", hs)
	temp, rh, err = sensor.ReadTemperatureAndRelativeHumidity(i2c, sht3x.REPEATABILITY_LOW)
	if err != nil {
		lg.Fatal(err)
	}
	lg.Infof("Temperature and relative humidity = %v*C, %v%%", temp, rh)

}
