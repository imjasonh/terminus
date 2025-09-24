package game

import "math"

type Vector struct {
	X, Y float64
}

func (v Vector) Add(other Vector) Vector {
	return Vector{v.X + other.X, v.Y + other.Y}
}

func (v Vector) Sub(other Vector) Vector {
	return Vector{v.X - other.X, v.Y - other.Y}
}

func (v Vector) Scale(s float64) Vector {
	return Vector{v.X * s, v.Y * s}
}

func (v Vector) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector) Normalize() Vector {
	length := v.Length()
	if length == 0 {
		return Vector{0, 0}
	}
	return Vector{v.X / length, v.Y / length}
}

func (v Vector) Rotate(angle float64) Vector {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return Vector{
		v.X*cos - v.Y*sin,
		v.X*sin + v.Y*cos,
	}
}
