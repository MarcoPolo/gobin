package handler

import (
	"github.com/grooveshark/golib/gslog"
	"github.com/kinghrothgar/goblin/conf"
	"github.com/kinghrothgar/goblin/storage/store"
	"net"
	"net/http"
	"regexp"
)

var (
	alphaReg = regexp.MustCompile("^[A-Za-z]+$")
)

func getGobData(w http.ResponseWriter, r *http.Request) []byte {
	//parse the multipart form in the request
	err := r.ParseMultipartForm(100000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	//get a ref to the parsed multipart form
	m := r.MultipartForm
	str := m.Value["gob"][0]
	return []byte(str)
}

func validateUID(w http.ResponseWriter, uid string) {
	// This is so someone can't access a horde goblin
	// by just puting the 'horde#uid' instead of 'horde/uid'
	// and prevents a lookup if it's obviously crap
	if len(uid) > conf.UIDLen || !alphaReg.MatchString(uid) {
		gslog.Debug("invalid uid")
		http.Error(w, "invalid uid", http.StatusBadRequest)
	}
	return
}

func validateHordeName(w http.ResponseWriter, hordeName string) {
	// TODO: horde max length?
	if len(hordeName) > 50 || !alphaReg.MatchString(hordeName) {
		gslog.Debug("invalid horde name")
		http.Error(w, "invalid horde name", http.StatusBadRequest)
	}
	return
}

func GetRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello"))
}

func GetGob(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	uid := params.Get(":uid")
	validateUID(w, uid)
	data, _, err := store.GetGob(uid)
	if err != nil {
		gslog.Debug("id does not exist")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Write(data)
}

func GetHorde(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	hordeName := params.Get(":horde")
	uidTimeList, err := store.GetHorde(hordeName)
	if err != nil {
		gslog.Debug("failed to get horde with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	page := ""
	for _, uidTimePair := range uidTimeList {
		page += "http://" + conf.Domain + "/" + uidTimePair.UID + "    " + uidTimePair.Time.String() + "\n"
	}
	w.Write([]byte(page))
}

func PostGob(w http.ResponseWriter, r *http.Request) {
	gobData := getGobData(w, r)
	uid := store.GetNewUID()
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(host)
	gslog.Debug("uid: %s, host: %s, ip: %s", uid, host, ip)
	if err := store.PutGob(uid, gobData, ip); err != nil {
		gslog.Debug("put gob failed with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write([]byte("http://" + conf.Domain + "/" + uid))
}

func PostHordeGob(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	hordeName := params.Get(":horde")
	validateHordeName(w, hordeName)
	gobData := getGobData(w, r)
	uid := store.GetNewUID()
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	ip := net.ParseIP(host)
	gslog.Debug("uid: %s, ip: %s", uid, ip)
	if err := store.PutHordeGob(uid, hordeName, gobData, ip); err != nil {
		gslog.Debug("put horde gob failed with error: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write([]byte("http://" + conf.Domain + "/" + uid))
}

func DelGob(w http.ResponseWriter, r *http.Request) {
}

func DelHordeGob(w http.ResponseWriter, r *http.Request) {
}