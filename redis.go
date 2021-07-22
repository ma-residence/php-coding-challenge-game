package main

import (
	"math/rand"
	"strconv"

	redis "gopkg.in/redis.v5"
)

// Client is the Redis client structure
type Client struct {
	*redis.Client
}

func getCoordinate(pattern string, client Client) int {
	data, _ := client.Get(pattern).Result()
	coordinate, _ := strconv.Atoi(data)
	return coordinate
}

func getCoordinates(pattern string, client Client) Coordinates {
	return Coordinates{X: getCoordinate(pattern+"x", client), Y: getCoordinate(pattern+"y", client)}
}

// GetUser returns the user coordinates
func (client Client) GetUser() Coordinates {
	return getCoordinates("pccg:user:", client)
}

// GetTarget returns the target coordinates if the user is in 2 cases away
func (client Client) GetTarget(userCoordinates Coordinates) Coordinates {
	targetCoordinates := getCoordinates("pccg:target:", client)

	if abs(userCoordinates.X-targetCoordinates.X) < 3 {
		if abs(userCoordinates.Y-targetCoordinates.Y) < 3 {
			return targetCoordinates
		}
	}
	return Coordinates{}
}

// Move moves the user in the specified direction
func (client Client) Move(direction string) Coordinates {
	coordinates := client.GetUser()
	x := coordinates.X
	y := coordinates.Y

	switch direction {
	case "up":
		if y > 1 {
			y = y - 1
		}
	case "down":
		if y < 21 {
			y = y + 1
		}
	case "left":
		if x > 1 {
			x = x - 1
		}
	case "right":
		if x < 21 {
			x = x + 1
		}
	}
	setCoordinates("pccg:user:", Coordinates{X: x, Y: y}, client)

	return Coordinates{X: x, Y: y}
}

func setCoordinates(pattern string, coordinates Coordinates, client Client) Coordinates {
	_ = client.Set(pattern+"x", coordinates.X, 0)
	_ = client.Set(pattern+"y", coordinates.Y, 0)

	return coordinates
}

func setTarget(coordinates Coordinates, client Client) Coordinates {
	return setCoordinates("pccg:target:", coordinates, client)
}

func setUser(coordinates Coordinates, client Client) Coordinates {
	return setCoordinates("pccg:user:", coordinates, client)
}

// InitUser set the starting coordinates for the user & reset the score
func (client Client) InitUser() Coordinates {
	coordinates := Coordinates{X: 10, Y: 10}
	_ = client.Del("pccg:user:score").Err()
	return setCoordinates("pccg:user:", coordinates, client)
}

// InitTarget set the starting coordinates for the target
func (client Client) InitTarget() Coordinates {
	return setCoordinates("pccg:target:", Coordinates{X: rand.Intn(20) + 1, Y: rand.Intn(20) + 1}, client)
}

func getScore(client Client) int {
	data, _ := client.Get("pccg:user:score").Result()
	score, _ := strconv.Atoi(data)
	return score
}

// Shot shots at the target
func (client Client) Shot(coordinates Coordinates) string {
	position := getCoordinates("pccg:target:", client)
	if position.X != coordinates.X || position.Y != coordinates.Y {
		return "miss"
	}

	score := getScore(client)
	if score == 2 {
		return "kill"
	}

	_ = client.Set("pccg:user:score", score+1, 0)

	return "touch"
}

func abs(val int) int {
	if val < 0 {
		return -val
	}
	return val
}

// New creates a new Redis Client
func NewRedis(URL string) (*Client, error) {
	options, err := redis.ParseURL(URL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(options)
	_, err = client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}
