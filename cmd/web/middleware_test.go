package main

import (
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {

	var myH myHandler

	// Прокидываем объект исполняющий интерфейс myHandler , который нужен нам как аргумент функции NoSurf
	h := NoSurf(&myH)

	switch v := h.(type) {
	case http.Handler:
		// ничего не произойдет
	default:
		t.Errorf("Unsupported type %T", v)
	}

}

func TestSessionLoad(t *testing.T) {

	var myH myHandler

	// Прокидываем объект исполняющий интерфейс myHandler , который нужен нам как аргумент функции NoSurf
	h := SessionLoad(&myH)

	switch v := h.(type) {
	case http.Handler:
		// ничего не произойдет
	default:
		t.Errorf("Unsupported type %T", v)
	}

}
