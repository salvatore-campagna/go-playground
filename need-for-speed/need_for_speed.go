/*
Package speed provides types and functions to simulate a remote-controlled car's behavior,
including tracking its battery usage, speed, and ability to complete a track of a given distance.

Types:
  - Car: Represents a remote-controlled car with attributes for battery level, battery drain per drive, speed, and distance covered.
  - Track: Represents a racing track with a specified distance.

Functions:

  - NewCar(speed, batteryDrain int) Car:
    Creates a new Car with a specified speed and battery drain per drive, starting with a full battery (100%) and zero distance.

  - NewTrack(distance int) Track:
    Creates a new Track with the specified distance.

  - Drive(car Car) Car:
    Simulates driving the car one time. If there is sufficient battery to complete the drive, it reduces the car's battery and increases its distance. If the battery is insufficient, the car remains stationary.

  - CanFinish(car Car, track Track) bool:
    Checks if the car has enough battery to complete the given track distance. Returns true if the car can finish the track, and false otherwise.
*/
package speed

// Car represents a remote-controlled car with battery, speed, and distance attributes.
// It keeps track of the battery level, battery drain per drive, current speed, and distance covered.
type Car struct {
	battery      int
	batteryDrain int
	speed        int
	distance     int
}

// NewCar creates a new remote controlled car with full battery and given specifications.
func NewCar(speed, batteryDrain int) Car {
	return Car{
		battery:      100,
		batteryDrain: batteryDrain,
		speed:        speed,
		distance:     0,
	}
}

// Track represents a racing track with a specified distance.
type Track struct {
	distance int
}

// NewTrack creates a new track
func NewTrack(distance int) Track {
	return Track{
		distance: distance,
	}
}

// Drive drives the car one time. If there is not enough battery to drive one more time,
// the car will not move.
func Drive(car Car) Car {
	c := car
	if car.battery >= car.batteryDrain {
		c.battery -= car.batteryDrain
		c.distance += car.speed
	}

	return c

}

// CanFinish checks if a car is able to finish a certain track.
func CanFinish(car Car, track Track) bool {
	drainsRequired := track.distance / car.speed
	requiredBattery := drainsRequired * car.batteryDrain
	return car.battery >= requiredBattery
}
