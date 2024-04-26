package Grizzly

import (
	"os"
	"strconv"
)

type castResult[T any] struct {
	value T
	err   error
}

func (c *castResult[T]) Get() T {
	if c.err != nil {
		panic(c.err)
	}
	return c.value
}

func (c *castResult[T]) Set(value T, err error) {
	c.value = value
	c.err = err
}

type Caster struct {
	Bool   castResult[bool]
	Int    castResult[int64]
	String castResult[string]
}

func Cast(value string) *Caster {
	c := &Caster{}
	c.Bool.Set(strconv.ParseBool(value))
	c.Int.Set(strconv.ParseInt(value, 10, 64))
	c.String.Set(value, nil)
	return c
}

func GetEnv(key string, defaultValue *string) *Caster {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue == nil {
			panic("Environment variable " + key + " is not set")
		}
		caster := Cast(*defaultValue) // Convert defaultValue to string
		return caster                 // Dereference the defaultValue pointer
	}
	c := Cast(value)
	return c
}
