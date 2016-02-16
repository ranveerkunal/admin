package admin

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"labix.org/v2/mgo"

	"github.com/goincremental/negroni-sessions"
	"github.com/martini-contrib/encoder"
)

var (
	admin = flag.String("admin", os.Getenv("ADMIN"), "")
)

type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", time.Time(t).UnixNano())), nil
}

type Log struct {
	OK bool
	S  interface{}
	T  Time
}

func NewLog(ok bool, s interface{}) (l Log) {
	l.OK = ok
	l.T = Time(time.Now())
	l.S = s
	return
}

type FBSubscriptionURL struct {
	U string
	V url.Values
}

type Status struct {
	AppDomain string
	Log       map[string]Log
}

func Authorize(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fmt.Printf("Invoked Auth: %v\n", r.URL)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(host)
	s := sessions.GetSession(r)
	if err != nil || (s.Get("user_id") != *admin && !ip.IsLoopback()) {
		http.Error(rw, "Please spare me :P", http.StatusUnauthorized)
		return
	}
	next(rw, r)
}

func FetchStatus(s *Status, ms *mgo.Session, enc encoder.Encoder, r *http.Request) (int, []byte) {
	s.Log["MongoDB"] = NewLog(true, "")
	if err := ms.Ping(); err != nil {
		s.Log["MongoDB"] = NewLog(false, err.Error())
	}
	return http.StatusOK, encoder.Must(enc.Encode(s))
}

func OK(rw http.ResponseWriter, r *http.Request) {
	og.Fatal("Setting OK")
	rw.Header().Set("Content-Type", "text/plain")
	rw.Write([]byte("OK"))
}
