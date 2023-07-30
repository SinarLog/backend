package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Credits: https://stackoverflow.com/questions/37889726/how-to-store-a-point-in-postgres-sql-database-using-gorm

// Point stores the (x,y) location on the Earth as a byte.
// X represents the longitude (the x-axis that shifts left and right)
// and Y represents the latitude (the y-axis that shifts up and down).
// See https://gis.stackexchange.com/a/68856
type Point struct {
	X, Y float64
}

func (p Point) Value() (driver.Value, error) {
	out := []byte{'('}
	out = strconv.AppendFloat(out, p.X, 'f', -1, 64)
	out = append(out, ',')
	out = strconv.AppendFloat(out, p.Y, 'f', -1, 64)
	out = append(out, ')')
	return out, nil
}

func (p *Point) Scan(src any) (err error) {
	var data []byte
	switch src := src.(type) {
	case []byte:
		data = src
	case string:
		data = []byte(src)
	case nil:
		return nil
	default:
		return errors.New("(*Point).Scan: unsupported data type")
	}

	if len(data) == 0 {
		return nil
	}

	data = data[1 : len(data)-1]
	for i := 0; i < len(data); i++ {
		if data[i] == ',' {
			if p.X, err = strconv.ParseFloat(string(data[:i]), 64); err != nil {
				return err
			}

			if p.Y, err = strconv.ParseFloat(string(data[i+1:]), 64); err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func (p Point) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.X,
			validation.Required.Error("geolocation is required"),
			validation.By(pointRequired),
		),
		validation.Field(&p.Y,
			validation.Required.Error("geolocation is required"),
			validation.By(pointRequired),
		),
	)
}

// Calculates in metres the distance from point p to v.
func (p Point) DistanceTo(v Point) float64 {
	R := 6371e3
	φ1 := p.Y * math.Pi / 180
	φ2 := v.Y * math.Pi / 180
	Δφ := (v.Y - p.Y) * math.Pi / 180
	Δλ := (v.X - p.X) * math.Pi / 180

	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c // in metres
}

type JSONB map[string]interface{}

func (a JSONB) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *JSONB) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("(*JSONB).Scan: unsupported data type")
	}

	return json.Unmarshal(b, &a)
}

/*
Validation types helpers
*/

// Validates that the property of a Point data type is required.
func pointRequired(value any) error {
	if value.(float64) == float64(0) {
		return fmt.Errorf("geolocation is required")
	}
	return nil
}
