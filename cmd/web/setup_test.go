package main

import (
	"net/http"
)

/*func TestMain(m *testing.M) {

	os.Exit(m.Run())
}*/

// в данном случае для запуска теста мы имитируем Handler interface в структуре myHandler
type myHandler struct{}

func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
