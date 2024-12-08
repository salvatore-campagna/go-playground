/*
Package meteorology provides types and methods to represent and format meteorological data,
including temperature, wind speed, and general weather conditions.
*/
package meteorology

import "fmt"

// TemperatureUnit represents the unit of temperature measurement (Celsius or Fahrenheit).
type TemperatureUnit int

const (
	Celsius    TemperatureUnit = 0
	Fahrenheit TemperatureUnit = 1
)

// String returns the string representation of the TemperatureUnit.
func (tu TemperatureUnit) String() string {
	return []string{"°C", "°F"}[tu]
}

// Temperature represents a temperature value with its unit.
type Temperature struct {
	degree int
	unit   TemperatureUnit
}

// String returns the string representation of the Temperature in the format "degree unit".
func (t Temperature) String() string {
	return fmt.Sprintf("%d %s", t.degree, t.unit)
}

// SpeedUnit represents the unit of speed measurement (km/h or mph).
type SpeedUnit int

const (
	KmPerHour    SpeedUnit = 0
	MilesPerHour SpeedUnit = 1
)

// String returns the string representation of the SpeedUnit.
func (su SpeedUnit) String() string {
	return []string{"km/h", "mph"}[su]
}

// Speed represents a speed value with its unit.
type Speed struct {
	magnitude int
	unit      SpeedUnit
}

// String returns the string representation of the Speed in the format "magnitude unit".
func (s Speed) String() string {
	return fmt.Sprintf("%d %s", s.magnitude, s.unit)
}

// MeteorologyData contains comprehensive weather data for a specific location.
type MeteorologyData struct {
	location      string
	temperature   Temperature
	windDirection string
	windSpeed     Speed
	humidity      int
}

// String returns a formatted string representation of the MeteorologyData.
func (md MeteorologyData) String() string {
	return fmt.Sprintf("%s: %s, Wind %s at %s, %d%% Humidity", md.location, md.temperature, md.windDirection, md.windSpeed, md.humidity)
}
