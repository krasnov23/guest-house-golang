package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("Form should be valid")
	}

}

func TestForm_Valid2(t *testing.T) {

	values := url.Values{}
	form := New(values)
	isValid := form.Valid()
	if !isValid {
		t.Error("Form should be invalid")
	}
}

func TestForm_Required(t *testing.T) {

	r := httptest.NewRequest(http.MethodPost, "/", nil)

	form := New(r.PostForm)

	form.Required("name", "email", "phone")

	if form.Valid() {
		t.Error("Form should not be valid")
	}

	postedData := url.Values{}
	postedData.Add("name", "Ivan")
	postedData.Add("email", "email@email.com")
	postedData.Add("phone", "123")

	r, _ = http.NewRequest(http.MethodPost, "/", nil)

	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("name", "email", "phone")
	if !form.Valid() {
		t.Error("Form should be valid")
	}
}

func TestForm_Has(t *testing.T) {

	r := httptest.NewRequest(http.MethodPost, "/", nil)
	form := New(r.PostForm)
	check := form.Has("name")

	if check {
		t.Error("Form should contain name")
	}

	postedData := url.Values{}
	postedData.Add("name", "Ivan")

	form = New(postedData)

	check = form.Has("name")
	if !check {
		t.Error("Form should contain name")
	}

}

func TestForm_MinLength(t *testing.T) {
	postedData := url.Values{}
	postedData.Add("name", "Ivan")
	form := New(postedData)
	form.MinLength("name", 5)

	if form.Valid() {
		t.Error("Form should be invalid")
	}

}

func TestForm_IsEmail(t *testing.T) {
	//r := httptest.NewRequest(http.MethodPost, "/", nil)

	postedData := url.Values{}
	postedData.Add("email", "Ivan")
	//r.PostForm = postedData
	form := New(postedData)
	form.IsEmail("email")
	if form.Valid() {
		t.Error("Form should be invalid")
	}

	if form.Errors.Get("email") != "This field must be a valid email address." {
		t.Error("Errors is empty")
	}

	postedData = url.Values{}
	postedData.Add("email", "sam@email.com")
	//r.PostForm = postedData
	form = New(postedData)
	form.IsEmail("email")
	if !form.Valid() {
		t.Error("Form should be invalid")
	}

	if form.Errors.Get("email") != "" {
		t.Error("Errors is exist")
	}
}
