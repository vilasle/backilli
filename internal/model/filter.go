package model


type Filter interface{
	Check() bool
} 