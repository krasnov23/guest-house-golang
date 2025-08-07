package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	_ "github.com/justinas/nosurf"
	"github.com/krasnov23/guest-house-golang/internal/models"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"gq", "/generals-quarters", "GET", []postData{}, http.StatusOK},
	{"ms", "/majors-suite", "GET", []postData{}, http.StatusOK},
	{"sa", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	//{"post-sa", "/search-availability", "POST", []postData{
	//	{key: "start", value: "2020-01-01"},
	//	{key: "end", value: "2020-01-02"},
	//}, http.StatusOK},
	//{"post-sa-json", "/search-availability-json", "POST", []postData{
	//	{key: "start", value: "2020-01-01"},
	//	{key: "end", value: "2020-01-02"},
	//}, http.StatusOK},
	//{"post-mr", "/make-reservation", "POST", []postData{
	//	{key: "first_name", value: "John"},
	//	{key: "last_name", value: "Smith"},
	//	{key: "email", value: "a@a.ru"},
	//	{key: "phone", value: "555-555-555"},
	//}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()

	// Создание тестового сервера для прогона тестов
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		if e.method == "GET" {
			// делаем реквест гет от имени клиента, прокидывая адрес сервера + наш роут
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("got status code %d, want %d", resp.StatusCode, e.expectedStatusCode)
			}

		} else {
			values := url.Values{}
			for _, x := range e.params {
				values.Add(x.key, x.value)
			}

			resp, err := ts.Client().PostForm(ts.URL+e.url, values)

			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("got status code %d, want %d", resp.StatusCode, e.expectedStatusCode)
			}
		}
	}
}

func TestRepository_Reservation(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)

	// Загружает или создаёт сессию, ассоциированную с переданным контекстом.
	// Возвращает новый контекст, в который встроена информация о сессии.
	// В контекст добавляется специальный объект сессии (обычно это структура, содержащая ID сессии и ссылку на хранилище)
	// Сам контекст при этом не содержит данных сессии - только ссылку на объект сессии
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	// возвращает копию запроса req с обновлённым контекстом ctx.
	req = req.WithContext(ctx)

	// Позволяет записывать ответ обработчика HTTP и проверять его в тестах (например, статус-код, тело ответа).
	rr := httptest.NewRecorder()

	// На самом деле, session.Put работает так:
	// Кладет данные в сессию по id который будет указан в ctx ()
	// context реквеста при этом не изменяется потому что уже хранит в себе определенным id сессии
	session.Put(ctx, "reservation", reservation)

	// адаптирует метод Reservation под интерфейс http.Handler.
	handler := http.HandlerFunc(Repo.Reservation)

	// Имитирует обработку HTTP-запроса:
	//  Как handler.ServeHTTP(rr, req) получает доступ к reservation?
	// Обработчик (Repo.Reservation) должен сам извлечь данные
	// Где-то внутри Repo.Reservation должен быть код вида: reservation, ok := m.App.Session.Get(req.Context(), "reservation").(models.Reservation)
	// Сессия восстанавливается по ID из контекста
	// При session.Load() (Put) в контекст записывается ID сессии. При последующих вызовах session.Get() библиотека:
	// Достаёт ID сессии из контекста.
	// Ищет данные в хранилище (память/Redis) по этому ID.
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status code %d, want %d", rr.Code, http.StatusOK)
	}

	// Вы не можете напрямую получить значение "reservation" из контекста без использования session.Get(), потому что:
	// Сессионные данные хранятся в отдельном хранилище сессий (память, Redis и т.д.), а не в самом контексте.
	// Контекст содержит только ссылку на сессию (например, ID сессии), но не сами данные сессии.

}

func TestRepository_Reservation_WithoutReservation(t *testing.T) {
	// тест кейс где reservation не в сессии
	req, _ := http.NewRequest("GET", "/make-reservation", nil)

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	// возвращает копию запроса req с обновлённым контекстом ctx.
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// адаптирует метод Reservation под интерфейс http.Handler.
	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("got status code %d, want %d", rr.Code, http.StatusTemporaryRedirect)
	}

}

func TestRepository_Reservation_WithNonExistentRoom(t *testing.T) {
	reservation := models.Reservation{
		RoomID: 3,
		Room: models.Room{
			ID:       3,
			RoomName: "General's Quarters",
		},
	}

	req, _ := http.NewRequest("GET", "/make-reservation", nil)

	// выполняет загрузку или инициализацию сессии
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	// возвращает копию запроса req с обновлённым контекстом ctx.
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	session.Put(ctx, "reservation", reservation)

	// адаптирует метод Reservation под интерфейс http.Handler.
	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("got status code %d, want %d", rr.Code, http.StatusTemporaryRedirect)
	}

	//fmt.Println(session.Get(ctx, "error"))
}

func TestRepository_PostReservation(t *testing.T) {

	reqBody := "start_date=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s&%s&%s&%s&%s&%s",
		reqBody, "end_date=2050-01-01", "first_name=John", "last_name=Doe", "email=jd@jd.com", "phone=123456789", "room_id=1")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	// Если заголовок не установлен:
	// Go не сможет автоматически распарсить тело запроса
	// Форма останется пустой (req.Form/req.PostForm)
	// Обработчик получит пустые данные и, вероятно, вернёт ошибку
	// HTTP-спецификация: Для POST-запросов с данными формы заголовок Content-Type обязателен согласно стандарту.
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("got status code %d, want %d", rr.Code, http.StatusSeeOther)
	}
}

func TestRepository_PostReservation_WithCantParseData(t *testing.T) {
	req, _ := http.NewRequest("POST", "/make-reservation", nil)

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReserbation handler returned wrong response code for missing post body %d, want %d", rr.Code, http.StatusTemporaryRedirect)
	}

	if session.Get(ctx, "error") != "cannot parse data" {
		t.Errorf("another error not `cannot parse data`")
	}

}

func TestRepository_PostReservation_WithCantParseStartDate(t *testing.T) {

	reqBody := "start_date=sssss"
	reqBody = fmt.Sprintf("%s&%s&%s&%s&%s&%s&%s",
		reqBody, "end_date=2050-01-01", "first_name=John", "last_name=Doe", "email=jd@jd.com", "phone=123456789", "room_id=1")

	// Можем также сделать следующим образом
	postedData := url.Values{}
	postedData.Add("start_date", "2050-01-01")

	// Переменную ниже можем запихнуть третим аргументом в NewRequest внутрь strings.NewReader
	//data := postedData.Encode()

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReserbation handler returned wrong response code for missing post body %d, want %d", rr.Code, http.StatusTemporaryRedirect)
	}

	if session.Get(ctx, "error") != "cannot parse start date" {
		t.Errorf("another error not `cannot parse start date`")
	}

}

func TestRepository_PostReservation_WithCantParseEndDate(t *testing.T) {

	reqBody := "start_date=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s&%s&%s&%s&%s&%s",
		reqBody, "end_date=sss", "first_name=John", "last_name=Doe", "email=jd@jd.com", "phone=123456789", "room_id=1")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReserbation handler returned wrong response code for missing post body %d, want %d", rr.Code, http.StatusTemporaryRedirect)
	}

	if session.Get(ctx, "error") != "cannot parse end date" {
		t.Errorf("another error not `cannot parse end date`")
	}

}

func TestRepository_PostReservation_WithCantParseIncorrectRoomId(t *testing.T) {

	reqBody := "start_date=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s&%s&%s&%s&%s&%s",
		reqBody, "end_date=2050-01-01", "first_name=John", "last_name=Doe", "email=jd@jd.com", "phone=123456789", "room_id=dddd")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body %d, want %d", rr.Code, http.StatusTemporaryRedirect)
	}

	if session.Get(ctx, "error") != "invalid data of room_id" {
		t.Errorf("another error not `invalid data of room_id`")
	}

}

func TestRepository_PostReservation_WithCantParseInvalidFormRequirements(t *testing.T) {
	reqBody := "start_date=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s&%s&%s&%s&%s",
		reqBody, "end_date=2050-01-01", "first_name=John", "last_name=Doe", "phone=123456789", "room_id=1")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("PostReservation handler returned wrong response code for missing post body %d, want %d", rr.Code, http.StatusSeeOther)
	}

}

func TestRepository_PostReservation_WithCantParseInvalidInsertReservation(t *testing.T) {
	reqBody := "start_date=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s&%s&%s&%s&%s&%s",
		reqBody, "end_date=2050-01-01", "first_name=John", "last_name=Doe", "email=jd@jd.com", "phone=123456789", "room_id=2")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body %d, want %d", rr.Code, http.StatusTemporaryRedirect)
	}

	if session.Get(ctx, "error") != "cant insert reservation to DB" {
		t.Errorf("another error not `cant insert reservation to DB`")
	}
}

func TestRepository_PostReservation_WithCantParseInvalidInsertRoomRestriction(t *testing.T) {
	reqBody := "start_date=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s&%s&%s&%s&%s&%s",
		reqBody, "end_date=2050-01-01", "first_name=John", "last_name=Doe", "email=jd@jd.com", "phone=123456789", "room_id=3")

	req, _ := http.NewRequest("POST", "/make-reservation", strings.NewReader(reqBody))

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("PostReservation handler returned wrong response code for missing post body %d, want %d", rr.Code, http.StatusTemporaryRedirect)
	}

	if session.Get(ctx, "error") != "cant insert room restrictions to DB" {
		t.Errorf("another error not `cant insert room restrictions to DB`")
	}
}

func TestRepository_AvailabilityJSON_RoomIsNotAvailable(t *testing.T) {

	reqBody := "start_date=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s&%s", reqBody, "end_date=2050-01-01", "room_id=1")

	req, _ := http.NewRequest("POST", "/search-availability-json", strings.NewReader(reqBody))

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler := http.HandlerFunc(Repo.AvailabilityJSON)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var j jsonResponse

	// Ответ (rr.Body.String()) парсится в структуру jsonResponse.
	// Если парсинг не удаётся, тест падает с ошибкой failed parse json.
	err = json.Unmarshal([]byte(rr.Body.String()), &j)

	if err != nil {
		t.Error("failed parse json")
	}

}

func TestRepository_ChooseRoom_MissingURLParameter(t *testing.T) {
	// Создаем запрос с пустым ID
	req, _ := http.NewRequest("GET", "/choose-room/", nil)

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	// Устанавливаем ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	// Проверяем редирект
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, rr.Code)
	}

	// Проверяем сообщение об ошибке в сессии
	flash := Repo.App.Session.PopString(req.Context(), "error")
	if flash != "missing url parameter" {
		t.Errorf("Expected flash 'missing url parameter', got '%s'", flash)
	}
}

func TestRepository_ChooseRoom_NoReservationInSession(t *testing.T) {
	// Создаем запрос с корректным ID
	req, _ := http.NewRequest("GET", "/choose-room/1", nil)

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	ctx = addIdToChiContext(ctx, "1")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, rr.Code)
	}

	flash := Repo.App.Session.PopString(req.Context(), "error")

	if flash != "problem getting reservation from session" {
		t.Errorf("Expected flash 'problem getting reservation from session', got '%s'", flash)
	}
}

func TestRepository_ChooseRoom_Success(t *testing.T) {
	// Создаем тестовую резервацию
	reservation := models.Reservation{
		RoomID: 0,
		// другие поля...
	}

	// Создаем запрос с ID комнаты
	req, _ := http.NewRequest("GET", "/choose-room/1", nil)

	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))

	if err != nil {
		log.Println(err)
	}

	req = req.WithContext(ctx)

	ctx = addIdToChiContext(ctx, "1")
	req = req.WithContext(ctx)

	// Кладем резервацию в сессию
	Repo.App.Session.Put(ctx, "reservation", reservation)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(Repo.ChooseRoom)
	handler.ServeHTTP(rr, req)

	// Проверяем редирект
	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Проверяем, что резервация обновилась
	updatedRes, ok := Repo.App.Session.Get(ctx, "reservation").(models.Reservation)
	if !ok {
		t.Error("Failed to get reservation from session")
	}

	if updatedRes.RoomID != 1 {
		t.Errorf("Expected room ID 1, got %d", updatedRes.RoomID)
	}
}

func addIdToChiContext(parentCtx context.Context, id string) context.Context {

	// Создаёт пустой RouteContext — специальный объект Chi для хранения параметров маршрута (например, id из /users/{id}).
	chiCtx := chi.NewRouteContext()

	// Вручную добавляет параметр id в контекст маршрута.
	//Теперь chi.URLParam(r, "id") в обработчике сможет получить это значение.
	chiCtx.URLParams.Add("id", id)
	//fmt.Println(chi.RouteCtxKey)

	// chi.RouteCtxKey - это ключ, по которому Chi ищет свой контекст
	// Зачем передавать chi.RouteCtxKey
	// Chi ищет свой контекст по этому ключу.
	// Когда вы вызываете chi.URLParam(r, "id"), фреймворк делает примерно следующее:
	// func URLParam(r *http.Request, key string) string {
	//    if rctx := r.Context().Value(chi.RouteCtxKey); rctx != nil {
	//        return rctx.(*RouteContext).URLParams.Get(key) // Ищет по chi.RouteCtxKey
	//    }
	//    return ""
	//}
	// Гарантирует, что Chi получит доступ к параметрам.
	// Без этого ключа Chi не сможет найти свой RouteContext и параметры маршрута
	// Почему нельзя использовать произвольный ключ?
	// Chi ждёт конкретный ключ (chi.RouteCtxKey).
	// Chi не найдёт свой контекст, и URLParam() вернёт "".
	return context.WithValue(parentCtx, chi.RouteCtxKey, chiCtx) // Сохраняем контекст Chi в родительский контекст
}
