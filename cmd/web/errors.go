package main

import "errors"

var (
	ErrUnprocessableForm = errors.New("main: unprocessable form")

	ErrBadDate = errors.New("main: could not parse date")
)
