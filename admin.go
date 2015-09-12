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

	"github.com/martini-contrib/encoder"
	"github.com/martini-contrib/sessions"
	"github.com/ranveerkunal/weblogger"
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

func Authorize(s sessions.Session, wlog *weblogger.Logger, r *http.Request, w http.ResponseWriter) {
	fmt.Printf("Invoked Auth: %v\n", r.URL)
	fmt.Printf("Invoked Auth: %v\n", s.Get("user_id"))
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(host)
	if err != nil || (s.Get("user_id") != *admin && !ip.IsLoopback()) {
		wlog.Remotef("user: %s, ip: %s", s.Get("user_id"), ip)
		http.Error(w, "Please spare me :P", http.StatusUnauthorized)
	}
}

func FetchStatus(s *Status, ms *mgo.Session, enc encoder.Encoder, r *http.Request) (int, []byte) {
	s.Log["MongoDB"] = NewLog(true, "")
	if err := ms.Ping(); err != nil {
		s.Log["MongoDB"] = NewLog(false, err.Error())
	}
	return http.StatusOK, encoder.Must(enc.Encode(s))
}

func OK(s *Status, r *http.Request) (int, []byte) {
	return http.StatusOK, []byte("ok")
}
