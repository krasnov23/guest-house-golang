package render

import (
	"encoding/gob"
	"github.com/alexedwards/scs/v2"
	"github.com/krasnov23/guest-house-golang/internal/config"
	"github.com/krasnov23/guest-house-golang/internal/models"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var session *scs.SessionManager
var testApp config.AppConfig

func TestMain(m *testing.M) {

	gob.Register(models.Reservation{})

	testApp.InProduction = false

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	testApp.InfoLog = infoLog

	// будет отображать Дата + время + сообщение в консоли
	errLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	testApp.ErrorLog = errLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	testApp.Session = session
	testApp.UseCache = true

	// все что написано сверху нужно нам для того чтобы имитировать переменную app в render.go
	app = &testApp

	os.Exit(m.Run())
}
