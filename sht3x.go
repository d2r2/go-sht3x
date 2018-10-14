package sht3x

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"time"

	i2c "github.com/d2r2/go-i2c"
	shell "github.com/d2r2/go-shell"
	"github.com/davecgh/go-spew/spew"
)

// Command byte's sequences
var (
	// Measure values in "single shot mode".
	CMD_SINGLE_MEASURE_HIGH_CSE   = []byte{0x2C, 0x06} // Single Measure of Temp. and Hum.; High precise; Clock stretching enabled
	CMD_SINGLE_MEASURE_MEDIUM_CSE = []byte{0x2C, 0x0D} // Single Measure of Temp. and Hum.; Medium precise; Clocl stretching enabled
	CMD_SINGLE_MEASURE_LOW_CSE    = []byte{0x2C, 0x10} // Single Measure of Temp. and Hum.; Low precise; Clock stretching enabled
	CMD_SINGLE_MEASURE_HIGH       = []byte{0x24, 0x00} // Single Measure of Temp. and Hum.; High precise
	CMD_SINGLE_MEASURE_MEDIUM     = []byte{0x24, 0x0B} // Single Measure of Temp. and Hum.; Medium precise
	CMD_SINGLE_MEASURE_LOW        = []byte{0x24, 0x16} // Single Measure of Temp. and Hum.; Low precise

	// Measure values in "periodic acquisition mode".
	CMD_PERIOD_MEASURE_05MPS_HIGH   = []byte{0x20, 0x32} // Periodic Measure of Temp. and Hum.; 0.5 Measurements Per Second; High precise
	CMD_PERIOD_MEASURE_05MPS_MEDIUM = []byte{0x20, 0x24} // Periodic Measure of Temp. and Hum.; 0.5 Measurements Per Second; Medium precise
	CMD_PERIOD_MEASURE_05MPS_LOW    = []byte{0x20, 0x2F} // Periodic Measure of Temp. and Hum.; 0.5 Measurements Per Second; Low precise
	CMD_PERIOD_MEASURE_1MPS_HIGH    = []byte{0x21, 0x30} // Periodic Measure of Temp. and Hum.; 1 Measurements Per Second; High precise
	CMD_PERIOD_MEASURE_1MPS_MEDIUM  = []byte{0x21, 0x26} // Periodic Measure of Temp. and Hum.; 1 Measurements Per Second; Medium precise
	CMD_PERIOD_MEASURE_1MPS_LOW     = []byte{0x21, 0x2D} // Periodic Measure of Temp. and Hum.; 1 Measurements Per Second; Low precise
	CMD_PERIOD_MEASURE_2MPS_HIGH    = []byte{0x22, 0x36} // Periodic Measure of Temp. and Hum.; 2 Measurements Per Second; High precise
	CMD_PERIOD_MEASURE_2MPS_MEDIUM  = []byte{0x22, 0x20} // Periodic Measure of Temp. and Hum.; 2 Measurements Per Second; Medium precise
	CMD_PERIOD_MEASURE_2MPS_LOW     = []byte{0x22, 0x2B} // Periodic Measure of Temp. and Hum.; 2 Measurements Per Second; Low precise
	CMD_PERIOD_MEASURE_4MPS_HIGH    = []byte{0x23, 0x34} // Periodic Measure of Temp. and Hum.; 4 Measurements Per Second; High precise
	CMD_PERIOD_MEASURE_4MPS_MEDIUM  = []byte{0x23, 0x22} // Periodic Measure of Temp. and Hum.; 4 Measurements Per Second; Medium precise
	CMD_PERIOD_MEASURE_4MPS_LOW     = []byte{0x23, 0x29} // Periodic Measure of Temp. and Hum.; 4 Measurements Per Second; Low precise
	CMD_PERIOD_MEASURE_10MPS_HIGH   = []byte{0x27, 0x37} // Periodic Measure of Temp. and Hum.; 10 Measurements Per Second; High precise
	CMD_PERIOD_MEASURE_10MPS_MEDIUM = []byte{0x27, 0x21} // Periodic Measure of Temp. and Hum.; 10 Measurements Per Second; Medium precise
	CMD_PERIOD_MEASURE_10MPS_LOW    = []byte{0x27, 0x2A} // Periodic Measure of Temp. and Hum.; 10 Measurements Per Second; Low precise

	// Alert management commands.
	// Works in conjunction with "periodic acquisition mode".
	// Equation must be respected: HIGH SET > HIGH CLEAR > LOW CLEAR > LOW SET.
	CMD_ALERT_READ_HIGH_SET    = []byte{0xE1, 0x1F} // Read high defined level, exceeding which alert is triggered ON
	CMD_ALERT_READ_HIGH_CLEAR  = []byte{0xE1, 0x14} // Read high defined level, falling behind which alert is triggering OFF
	CMD_ALERT_READ_LOW_CLEAR   = []byte{0xE1, 0x09} // Read low defined level, exceeding which alert is triggered OFF
	CMD_ALERT_READ_LOW_SET     = []byte{0xE1, 0x02} // Read low defined level, falling behind which alert is triggering ON
	CMD_ALERT_WRITE_HIGH_SET   = []byte{0x61, 0x1D} // Write high level, exceeding which alert is triggered ON
	CMD_ALERT_WRITE_HIGH_CLEAR = []byte{0x61, 0x16} // Write high level, falling behind which alert is triggering OFF
	CMD_ALERT_WRITE_LOW_CLEAR  = []byte{0x61, 0x0B} // Write low level, exceeding which alert is triggered OFF
	CMD_ALERT_WRITE_LOW_SET    = []byte{0x61, 0x00} // Write low level, falling behind which alert is triggering ON

	// Heater management commands.
	CMD_ENABLE_HEATER  = []byte{0x30, 0x6D} // Switch heater on
	CMD_DISABLE_HEATER = []byte{0x30, 0x66} // Switch heater off

	// Status register commands.
	CMD_READ_STATUS_REG  = []byte{0xF3, 0x2D} // Read status register
	CMD_CLEAR_STATUS_REG = []byte{0x30, 0x41} // Clear status register

	// Other commands.
	CMD_PERIOD_FETCH = []byte{0xE0, 0x00} // Read data after being measured by periodic acquisition mode command
	CMD_ART          = []byte{0x2B, 0x32} // Activate "accelerated response time"
	CMD_BREAK        = []byte{0x30, 0x93} // Interrupt "periodic acqusition mode" and return to "single shot mode"
	CMD_RESET        = []byte{0x30, 0xA2} // Soft reset command
)

// MeasureRepeatability used to define measure precision.
type MeasureRepeatability int

const (
	REPEATABILITY_LOW    MeasureRepeatability = iota + 1 // Low precision
	REPEATABILITY_MEDIUM                                 // Medium precision
	REPEATABILITY_HIGH                                   // High precision
)

// String define stringer interface.
func (v MeasureRepeatability) String() string {
	switch v {
	case REPEATABILITY_LOW:
		return "Measure Repeatability Low"
	case REPEATABILITY_MEDIUM:
		return "Measure Repeatability Medium"
	case REPEATABILITY_HIGH:
		return "Measure Repeatability High"
	default:
		return "<unknown>"
	}
}

// GetWorstMeasureTime define how long to wait for the measure process
// to complete according to specification.
func (v MeasureRepeatability) GetWorstMeasureTime() time.Duration {
	switch v {
	case REPEATABILITY_LOW:
		return 4500 * time.Microsecond
	case REPEATABILITY_MEDIUM:
		return 6500 * time.Microsecond
	case REPEATABILITY_HIGH:
		return 15500 * time.Microsecond
	default:
		return 0
	}
}

// StatusRegFlag determine sensor states.
// It shows various sensor pending events and returns heater status.
type StatusRegFlag uint16

const (
	ALERT_PENDING         StatusRegFlag = 0x8000
	HEATER_ENABLED        StatusRegFlag = 0x2000
	HUMIDITY_ALERT        StatusRegFlag = 0x0800
	TEMPERATURE_ALERT     StatusRegFlag = 0x0400
	RESET_DETECTED        StatusRegFlag = 0x0010
	COMMAND_FAILED        StatusRegFlag = 0x0002
	WRITE_DATA_CRC_FAILED StatusRegFlag = 0x0001
)

// String define stringer interface.
func (v StatusRegFlag) String() string {
	const divider = " | "
	var buf bytes.Buffer
	if v&ALERT_PENDING != 0 {
		buf.WriteString("ALERT_PENDING" + divider)
	}
	if v&HEATER_ENABLED != 0 {
		buf.WriteString("HEATER_ENABLED" + divider)
	}
	if v&HUMIDITY_ALERT != 0 {
		buf.WriteString("HUMIDITY_ALERT" + divider)
	}
	if v&TEMPERATURE_ALERT != 0 {
		buf.WriteString("TEMPERATURE_ALERT" + divider)
	}
	if v&RESET_DETECTED != 0 {
		buf.WriteString("RESET_DETECTED" + divider)
	}
	if v&COMMAND_FAILED != 0 {
		buf.WriteString("COMMAND_FAILED" + divider)
	}
	if v&WRITE_DATA_CRC_FAILED != 0 {
		buf.WriteString("WRITE_DATA_CRC_FAILED" + divider)
	}
	if buf.Len() > 0 {
		buf.Truncate(buf.Len() - len(divider))
	}
	return buf.String()
}

// PeriodicMeasure identify pause between subsequent measures
// in "periodic data acquisition" mode.
type PeriodicMeasure int

const (
	PERIODIC_05MPS PeriodicMeasure = iota + 1 // 1 measurement per each 2 seconds
	PERIODIC_1MPS                             // 1 measurement per second
	PERIODIC_2MPS                             // 2 measurements per second
	PERIODIC_4MPS                             // 4 measurements per second
	PERIODIC_10MPS                            // 10 measurements per second
)

// String define stringer interface.
func (v PeriodicMeasure) String() string {
	switch v {
	case PERIODIC_05MPS:
		return "Periodic Measurement 0.5 MPS"
	case PERIODIC_1MPS:
		return "Periodic Measurement 1 MPS"
	case PERIODIC_2MPS:
		return "Periodic Measurement 2 MPS"
	case PERIODIC_4MPS:
		return "Periodic Measurement 4 MPS"
	case PERIODIC_10MPS:
		return "Periodic Measurement 10 MPS"
	default:
		return "<unknown>"
	}
}

// GetDuration identify pause between measures depending on PeriodicMeasure value.
func (v PeriodicMeasure) GetDuration() time.Duration {
	var timeDur time.Duration
	switch v {
	case PERIODIC_05MPS:
		timeDur = time.Millisecond * 2000
	case PERIODIC_1MPS:
		timeDur = time.Millisecond * 1000
	case PERIODIC_2MPS:
		timeDur = time.Millisecond * 500
	case PERIODIC_4MPS:
		timeDur = time.Millisecond * 250
	case PERIODIC_10MPS:
		timeDur = time.Millisecond * 100
	}
	return timeDur
}

// SHT3X is a sensor itself.
type SHT3X struct {
	lastStatusReg *uint16
}

// NewSHT3X return new sensor instance.
func NewSHT3X() *SHT3X {
	v := &SHT3X{}
	return v
}

// ReadStatusReg return status register flags.
// You should use constants of type StatusRegFlag to distinguish
// individual states received from sensor.
func (v *SHT3X) ReadStatusReg(i2c *i2c.I2C) (uint16, error) {
	if v.lastStatusReg == nil {
		_, err := i2c.WriteBytes(CMD_READ_STATUS_REG)
		if err != nil {
			return 0, err
		}
		reg, err := v.readDataWithCRCCheck(i2c, 1)
		if err != nil {
			return 0, err
		}
		v.lastStatusReg = &reg[0]
	}
	return *v.lastStatusReg, nil
}

// readDataWithCRCCheck read block of data which ordinary contain
// uncompensated temperature and humidity values.
func (v *SHT3X) readDataWithCRCCheck(i2c *i2c.I2C, blockCount int) ([]uint16, error) {
	const blockSize = 2 + 1
	data := make([]struct {
		Data [2]byte
		CRC  byte
	}, blockCount)

	err := readDataToStruct(i2c, blockSize*blockCount, binary.BigEndian, data)
	if err != nil {
		return nil, err
	}
	var results []uint16
	for i := 0; i < blockCount; i++ {
		calcCRC := calcCRC_SHT3X(0xFF, data[i].Data[:2])
		crc := data[i].CRC
		if calcCRC != crc {
			err := errors.New(spew.Sprintf(
				"CRCs doesn't match: CRC from sensor (0x%0X) != calculated CRC (0x%0X)",
				crc, calcCRC))
			return nil, err
		} else {
			lg.Debugf("CRCs verified: CRC from sensor (0x%0X) = calculated CRC (0x%0X)",
				crc, calcCRC)
		}
		results = append(results, getU16BE(data[i].Data[:2]))

	}
	return results, nil
}

// Reset reboot a sensor.
func (v *SHT3X) Reset(i2c *i2c.I2C) error {
	lg.Debug("Reset sensor...")
	_, err := i2c.WriteBytes(CMD_RESET)
	if err != nil {
		return err
	}
	// Powerup time from specification
	time.Sleep(time.Microsecond * 1500)
	return nil
}

// SetHeaterStatus enable or disable heater.
func (v *SHT3X) SetHeaterStatus(i2c *i2c.I2C, enableHeater bool) error {
	lg.Debug("Setting heater on/off...")
	var cmd []byte
	if enableHeater {
		cmd = CMD_ENABLE_HEATER
	} else {
		cmd = CMD_DISABLE_HEATER
	}
	_, err := i2c.WriteBytes(cmd)
	if err != nil {
		return err
	}
	// No conversion time defined in docs for this command,
	// but error thrown out, if no any pause provided.
	time.Sleep(time.Millisecond * 1)
	return nil
}

// GetHeaterStatus return heater status: enabled (true) or disabled (false).
func (v *SHT3X) GetHeaterStatus(i2c *i2c.I2C) (bool, error) {
	lg.Debug("Getting heater status...")
	v.lastStatusReg = nil
	ur, err := v.ReadStatusReg(i2c)
	if err != nil {
		return false, err
	}
	return (StatusRegFlag)(ur)&HEATER_ENABLED != 0, nil
}

// GetAlertPendingStatus return alert pending status: found (true) or not (false).
func (v *SHT3X) GetAlertPendingStatus(i2c *i2c.I2C) (bool, error) {
	lg.Debug("Getting alert pending status...")
	v.lastStatusReg = nil
	ur, err := v.ReadStatusReg(i2c)
	if err != nil {
		return false, err
	}
	return (StatusRegFlag)(ur)&ALERT_PENDING != 0, nil
}

// GetHumidityAlertStatus return humidity alert pending status: found (true) or not (false).
func (v *SHT3X) GetHumidityAlertStatus(i2c *i2c.I2C) (bool, error) {
	lg.Debug("Getting humidity alert status...")
	v.lastStatusReg = nil
	ur, err := v.ReadStatusReg(i2c)
	if err != nil {
		return false, err
	}
	return (StatusRegFlag)(ur)&HUMIDITY_ALERT != 0, nil
}

// GetTemperatureAlertStatus return humidity alert pending status: found (true) or not (false).
func (v *SHT3X) GetTemperatureAlertStatus(i2c *i2c.I2C) (bool, error) {
	lg.Debug("Getting temperature alert status...")
	v.lastStatusReg = nil
	ur, err := v.ReadStatusReg(i2c)
	if err != nil {
		return false, err
	}
	return (StatusRegFlag)(ur)&TEMPERATURE_ALERT != 0, nil
}

// CheckResetDetected return system reset detected : found (true) or not (false).
func (v *SHT3X) CheckResetDetected(i2c *i2c.I2C) (bool, error) {
	lg.Debug("Checking system reset status...")
	v.lastStatusReg = nil
	ur, err := v.ReadStatusReg(i2c)
	if err != nil {
		return false, err
	}
	return (StatusRegFlag)(ur)&RESET_DETECTED != 0, nil
}

// CheckCommandFailed return last command status: failed (true) or not (false).
func (v *SHT3X) CheckCommandFailed(i2c *i2c.I2C) (bool, error) {
	lg.Debug("Checking last command status...")
	v.lastStatusReg = nil
	ur, err := v.ReadStatusReg(i2c)
	if err != nil {
		return false, err
	}
	return (StatusRegFlag)(ur)&COMMAND_FAILED != 0, nil
}

// CheckWrittedChecksumIsIncorrect return last command status: not correct (true) correct (false).
func (v *SHT3X) CheckWrittenChecksumIsIncorrect(i2c *i2c.I2C) (bool, error) {
	lg.Debug("Checking last writted data checksum status...")
	v.lastStatusReg = nil
	ur, err := v.ReadStatusReg(i2c)
	if err != nil {
		return false, err
	}
	return (StatusRegFlag)(ur)&WRITE_DATA_CRC_FAILED != 0, nil
}

// initiateMeasure used to initiate temperature and humidity measurement process.
func (v *SHT3X) initiateMeasure(i2c *i2c.I2C, cmd []byte,
	precision MeasureRepeatability) error {

	_, err := i2c.WriteBytes(cmd)
	if err != nil {
		return err
	}
	// Wait according to conversion time specification
	pause := precision.GetWorstMeasureTime()
	time.Sleep(pause)
	return nil
}

// ReadUncompTemperatureAndHumidity returns uncompensated humidity and
// temperature obtained from sensor in "single shot mode".
func (v *SHT3X) ReadUncompTemperatureAndHumidity(i2c *i2c.I2C,
	precision MeasureRepeatability) (uint16, uint16, error) {

	lg.Debug("Measuring temprature and humidity...")
	var cmd []byte
	switch precision {
	case REPEATABILITY_LOW:
		cmd = CMD_SINGLE_MEASURE_LOW
	case REPEATABILITY_MEDIUM:
		cmd = CMD_SINGLE_MEASURE_MEDIUM
	case REPEATABILITY_HIGH:
		cmd = CMD_SINGLE_MEASURE_HIGH
	}
	err := v.initiateMeasure(i2c, cmd, precision)
	if err != nil {
		return 0, 0, err
	}

	data, err := v.readDataWithCRCCheck(i2c, 2)
	if err != nil {
		return 0, 0, err
	}
	return data[0], data[1], nil
}

// ReadTemperatureAndRelativeHumidity returns humidity and
// temperature obtained from sensor in "single shot mode".
func (v *SHT3X) ReadTemperatureAndRelativeHumidity(i2c *i2c.I2C,
	precision MeasureRepeatability) (float32, float32, error) {

	ut, urh, err := v.ReadUncompTemperatureAndHumidity(i2c, precision)
	if err != nil {
		return 0, 0, err
	}
	lg.Debugf("Temperature and humidity uncompensated = %v, %v", ut, urh)
	temp := v.uncompTemperatureToCelsius(ut)
	rh := v.uncompHumidityToRelativeHumidity(urh)
	return temp, rh, nil
}

// Convert uncompensated humidity to relative humidity.
func (v *SHT3X) uncompHumidityToRelativeHumidity(uh uint16) float32 {
	rh := float32(uh) * 100 / (0x10000 - 1)
	rh2 := round32(rh, 2)
	return rh2
}

// Convert uncompensated temperature to celsius value.
func (v *SHT3X) uncompTemperatureToCelsius(ut uint16) float32 {
	temp := float32(ut)*175/(0x10000-1) - 45
	temp2 := round32(temp, 2)
	return temp2
}

// Reverse conversion of relative humidity to uncompensated one.
func (v *SHT3X) relativeHumidityToUncompHimidity(rh float32) uint16 {
	uh := uint16(rh * (0x10000 - 1) / 100)
	return uh
}

// Reverse conversion of celsius to uncompensated temperature.
func (v *SHT3X) celsiusToUncompTemperature(celsius float32) uint16 {
	ut := uint16((celsius + 45) * (0x10000 - 1) / 175)
	return ut
}

// StartPeriodicTemperatureAndHumidityMeasure send command to the sensor
// to start continuous measurement process of temperature and humidity
// with the pace defined by period parameter. Measurement process should be
// interrupted by Break command. Use Fetch... methods to read results.
func (v *SHT3X) StartPeriodicTemperatureAndHumidityMeasure(i2c *i2c.I2C,
	period PeriodicMeasure, precision MeasureRepeatability) error {

	var cmd []byte
	switch period {
	case PERIODIC_05MPS:
		switch precision {
		case REPEATABILITY_LOW:
			cmd = CMD_PERIOD_MEASURE_05MPS_LOW
		case REPEATABILITY_MEDIUM:
			cmd = CMD_PERIOD_MEASURE_05MPS_MEDIUM
		case REPEATABILITY_HIGH:
			cmd = CMD_PERIOD_MEASURE_05MPS_HIGH
		}
	case PERIODIC_1MPS:
		switch precision {
		case REPEATABILITY_LOW:
			cmd = CMD_PERIOD_MEASURE_1MPS_LOW
		case REPEATABILITY_MEDIUM:
			cmd = CMD_PERIOD_MEASURE_1MPS_MEDIUM
		case REPEATABILITY_HIGH:
			cmd = CMD_PERIOD_MEASURE_1MPS_HIGH
		}
	case PERIODIC_2MPS:
		switch precision {
		case REPEATABILITY_LOW:
			cmd = CMD_PERIOD_MEASURE_2MPS_LOW
		case REPEATABILITY_MEDIUM:
			cmd = CMD_PERIOD_MEASURE_2MPS_MEDIUM
		case REPEATABILITY_HIGH:
			cmd = CMD_PERIOD_MEASURE_2MPS_HIGH
		}
	case PERIODIC_4MPS:
		switch precision {
		case REPEATABILITY_LOW:
			cmd = CMD_PERIOD_MEASURE_4MPS_LOW
		case REPEATABILITY_MEDIUM:
			cmd = CMD_PERIOD_MEASURE_4MPS_MEDIUM
		case REPEATABILITY_HIGH:
			cmd = CMD_PERIOD_MEASURE_4MPS_HIGH
		}
	case PERIODIC_10MPS:
		switch precision {
		case REPEATABILITY_LOW:
			cmd = CMD_PERIOD_MEASURE_10MPS_LOW
		case REPEATABILITY_MEDIUM:
			cmd = CMD_PERIOD_MEASURE_10MPS_MEDIUM
		case REPEATABILITY_HIGH:
			cmd = CMD_PERIOD_MEASURE_10MPS_HIGH
		}
	}
	err := v.initiateMeasure(i2c, cmd, precision)
	if err != nil {
		return err
	}

	return nil
}

// Break interrupt "periodic data acquisition mode" and
// return sensor to "single shot mode".
func (v *SHT3X) Break(i2c *i2c.I2C) error {
	lg.Debug("Interrupt periodic data acquisition mode...")
	_, err := i2c.WriteBytes(CMD_BREAK)
	if err != nil {
		return err
	}
	return nil
}

// FetchUncompTemperatureAndHumidityWithContext return
// uncompensated temperature and humidity obtained from sensor.
func (v *SHT3X) FetchUncompTemperatureAndHumidity(i2c *i2c.I2C,
	period PeriodicMeasure) (uint16, uint16, error) {
	// Create default context
	ctx := context.Background()
	// Reroute call
	return v.FetchUncompTemperatureAndHumidityWithContext(ctx,
		i2c, period)
}

// FetchUncompTemperatureAndHumidityWithContext return
// uncompensated temperature and humidity obtained from sensor.
// Use context parameter, since operation is time consuming
// (can take up to 2 seconds, waiting for results).
func (v *SHT3X) FetchUncompTemperatureAndHumidityWithContext(parent context.Context,
	i2c *i2c.I2C, period PeriodicMeasure) (uint16, uint16, error) {

	_, err := i2c.WriteBytes(CMD_PERIOD_FETCH)
	if err != nil {
		return 0, 0, err
	}

	done := make(chan struct{})
	defer close(done)
	// Create context with cancelation possibility.
	ctx, cancel := context.WithCancel(parent)
	// Run goroutine waiting for OS termantion events, including keyboard Ctrl+C.
	shell.CloseContextOnKillSignal(cancel, done)

	retryCount := 5
	var data []uint16
	timeDur := period.GetDuration()
	first := true
	for retryCount >= 0 {
		data, err = v.readDataWithCRCCheck(i2c, 2)
		// Once sensor doesn't ready provide data, sensor is replying with i2c NACK
		// and it throw error "read /dev/i2c-x: no such device or address".
		// So, we are retrying after pause specific to period parameter
		// which define "measures per second" value.
		if err != nil {
			if retryCount == 0 {
				return 0, 0, err
			}
			// sleep timeDur time
			select {
			// check for termination request
			case <-ctx.Done():
				// interrupt loop, if pending termination
				return 0, 0, ctx.Err()
			// sleep before new attempt.
			case <-time.After(timeDur):
			}
			if first {
				timeDur = timeDur / 10
			}
			retryCount--
		} else {
			break
		}
		first = false
	}
	return data[0], data[1], nil
}

// FetchTemperatureAndRelativeHumidity wait for uncompensated temperature
// and humidity values and convert them to float values (celcius and related humidity).
func (v *SHT3X) FetchTemperatureAndRelativeHumidity(i2c *i2c.I2C,
	period PeriodicMeasure) (float32, float32, error) {
	// Create default context
	ctx := context.Background()
	// Reroute call
	return v.FetchTemperatureAndRelativeHumidityWithContext(ctx,
		i2c, period)
}

// FetchTemperatureAndRelativeHumidityWithContext wait for uncompensated temperature
// and humidity values and convert them to float values (celcius and related humidity).
// Use context parameter, since operation is time consuming
// (can take up to 2 seconds, waiting for results).
func (v *SHT3X) FetchTemperatureAndRelativeHumidityWithContext(parent context.Context,
	i2c *i2c.I2C, period PeriodicMeasure) (float32, float32, error) {

	ut, urh, err := v.FetchUncompTemperatureAndHumidityWithContext(parent, i2c, period)
	if err != nil {
		return 0, 0, err
	}
	lg.Debugf("Temperature and RH uncompensated = %v, %v", ut, urh)
	temp := v.uncompTemperatureToCelsius(ut)
	rh := v.uncompHumidityToRelativeHumidity(urh)
	return temp, rh, nil
}

// Read alert temperature and humidity limits from sensor.
func (v *SHT3X) readAlertData(i2c *i2c.I2C, cmd []byte) (float32, float32, error) {
	_, err := i2c.WriteBytes(cmd)
	if err != nil {
		return 0, 0, err
	}
	data, err := v.readDataWithCRCCheck(i2c, 1)
	if err != nil {
		return 0, 0, err
	}

	uh := data[0] & 0xFE00
	ut := data[0] & 0x01FF << 7

	temp := v.uncompTemperatureToCelsius(ut)
	rh := v.uncompHumidityToRelativeHumidity(uh)
	return temp, rh, nil
}

// Write alert temperature and humidity limits to the sensor.
func (v *SHT3X) writeAlertData(i2c *i2c.I2C, cmd []byte, temp, hum float32) error {
	ut := v.celsiusToUncompTemperature(temp)
	uh := v.relativeHumidityToUncompHimidity(hum)

	u := uh&0xFE00 | (ut & 0xFF80 >> 7)
	data := []byte{byte(u & 0xFF00 >> 8), byte(u & 0x00FF)}
	crc := calcCRC_SHT3X(0xFF, data)
	b := append(cmd, data...)
	b = append(b, crc)

	_, err := i2c.WriteBytes(b)
	if err != nil {
		return err
	}
	// No conversion time defined in docs for this command,
	// but error thrown out, if no any pause provided.
	time.Sleep(time.Millisecond * 1)

	return nil
}

// ReadAlertHighSet read sensor alert HIGH SET limits
// for temperature and humidity.
func (v *SHT3X) ReadAlertHighSet(i2c *i2c.I2C) (float32, float32, error) {
	lg.Debug("Getting allert HIGH SET limit...")
	temp, rh, err := v.readAlertData(i2c, CMD_ALERT_READ_HIGH_SET)
	if err != nil {
		return 0, 0, err
	}
	return temp, rh, nil

}

// ReadAlertHighClear read sensor alert HIGH CLEAR limits
// for temperature and humidity.
func (v *SHT3X) ReadAlertHighClear(i2c *i2c.I2C) (float32, float32, error) {
	lg.Debug("Getting allert HIGH CLEAR limit...")
	temp, rh, err := v.readAlertData(i2c, CMD_ALERT_READ_HIGH_CLEAR)
	if err != nil {
		return 0, 0, err
	}
	return temp, rh, nil

}

// ReadAlertLowClear read sensor alert LOW CLEAR limits
// for temperature and humidity.
func (v *SHT3X) ReadAlertLowClear(i2c *i2c.I2C) (float32, float32, error) {
	lg.Debug("Getting allert LOW CLEAR limit...")
	temp, rh, err := v.readAlertData(i2c, CMD_ALERT_READ_LOW_CLEAR)
	if err != nil {
		return 0, 0, err
	}
	return temp, rh, nil

}

// ReadAlertLowSet read sensor alert LOW SET limits
// for temperature and humidity.
func (v *SHT3X) ReadAlertLowSet(i2c *i2c.I2C) (float32, float32, error) {
	lg.Debug("Getting allert LOW SET limit...")
	temp, rh, err := v.readAlertData(i2c, CMD_ALERT_READ_LOW_SET)
	if err != nil {
		return 0, 0, err
	}
	return temp, rh, nil

}

// WriteAlertHighSet write alert HIGH SET limits
// for temperature and humidity to the sensor.
func (v *SHT3X) WriteAlertHighSet(i2c *i2c.I2C, temp, hum float32) error {
	lg.Debug("Setting allert HIGH SET limit...")
	err := v.writeAlertData(i2c, CMD_ALERT_WRITE_HIGH_SET, temp, hum)
	if err != nil {
		return err
	}
	return nil

}

// WriteAlertHighClear write alert HIGH CLEAR limits
// for temperature and humidity to the sensor.
func (v *SHT3X) WriteAlertHighClear(i2c *i2c.I2C, temp, hum float32) error {
	lg.Debug("Setting allert HIGH CLEAR limit...")
	err := v.writeAlertData(i2c, CMD_ALERT_WRITE_HIGH_CLEAR, temp, hum)
	if err != nil {
		return err
	}
	return nil

}

// WriteAlertLowClear write alert LOW CLEAR limits
// for temperature and humidity to the sensor.
func (v *SHT3X) WriteAlertLowClear(i2c *i2c.I2C, temp, hum float32) error {
	lg.Debug("Setting allert LOW CLEAR limit...")
	err := v.writeAlertData(i2c, CMD_ALERT_WRITE_LOW_CLEAR, temp, hum)
	if err != nil {
		return err
	}
	return nil

}

// WriteAlertLowSet write alert LOW SET limits
// for temperature and humidity to the sensor.
func (v *SHT3X) WriteAlertLowSet(i2c *i2c.I2C, temp, hum float32) error {
	lg.Debug("Setting allert LOW SET limit...")
	err := v.writeAlertData(i2c, CMD_ALERT_WRITE_LOW_SET, temp, hum)
	if err != nil {
		return err
	}
	return nil

}
