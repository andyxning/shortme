package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/andyxning/shortme/conf"
	"github.com/andyxning/shortme/short"

	"github.com/gorilla/mux"
)

func Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortededURL := vars["shortenedURL"]

	longURL, err := short.Shorter.Expand(shortededURL)
	if err != nil {
		log.Printf("redirect short url error. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	} else {
		if len(longURL) != 0 {
			w.Header().Set("Location", longURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func ShortURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("read short request error. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusInternalServerError)})
		w.Write(errMsg)
		return
	}

	var shortReq shortReq
	err = json.Unmarshal(body, &shortReq)
	if err != nil {
		log.Printf("parse short request error. %v", err)
		w.WriteHeader(http.StatusBadRequest)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusBadRequest)})
		w.Write(errMsg)
		return
	}

	var longURL *url.URL
	longURL, err = url.Parse(shortReq.LongURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errMsg, _ := json.Marshal(errorResp{Msg: "requested url is malformed"})
		w.Write(errMsg)
		return
	} else {
		if longURL.Host == conf.Conf.Common.DomainName {
			w.WriteHeader(http.StatusBadRequest)
			errMsg, _ := json.Marshal(errorResp{Msg: "requested url is already shortened"})
			w.Write(errMsg)
			return
		}
		if strings.ToLower(longURL.Scheme) != "http" && strings.ToLower(longURL.Scheme) != "https" {
			w.WriteHeader(http.StatusBadRequest)
			errMsg, _ := json.Marshal(errorResp{Msg: "requested url is not a http or https url"})
			w.Write(errMsg)
			return
		}
	}

	var shortenedURL string

	//check the longURL is existed
	shortenedURL, err = short.Shorter.Query(shortReq.LongURL)
	if err == nil || shortenedURL != "" {
		shortenedURL = (&url.URL{
			Scheme: conf.Conf.Common.Schema,
			Host:   conf.Conf.Common.DomainName,
			Path:   shortenedURL,
		}).String()
		shortResp, _ := json.Marshal(shortResp{ShortURL: shortenedURL})
		w.Write(shortResp)
		return
	}

	shortenedURL, err = short.Shorter.Short(shortReq.LongURL)
	shortenedURL = (&url.URL{
		Scheme: conf.Conf.Common.Schema,
		Host:   conf.Conf.Common.DomainName,
		Path:   shortenedURL,
	}).String()
	if err != nil {
		log.Printf("short url error. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusInternalServerError)})
		w.Write(errMsg)
		return
	} else {
		shortResp, _ := json.Marshal(shortResp{ShortURL: shortenedURL})
		w.Write(shortResp)
	}
}

func ExpandURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("read expand request error. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusInternalServerError)})
		w.Write(errMsg)
		return
	}

	var expandReq expandReq
	err = json.Unmarshal(body, &expandReq)
	if err != nil {
		log.Printf("parse expand request error. %v", err)
		w.WriteHeader(http.StatusBadRequest)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusBadRequest)})
		w.Write(errMsg)
		return
	}

	var shortURL *url.URL
	shortURL, err = url.Parse(expandReq.ShortURL)
	if err != nil {
		log.Printf(`"%v" is not a valid url`, expandReq.ShortURL)
		w.WriteHeader(http.StatusBadRequest)
		errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusBadRequest)})
		w.Write(errMsg)
		return
	} else {
		var expandedURL string
		expandedURL, err = short.Shorter.Expand(strings.TrimLeft(shortURL.Path, "/"))
		if err != nil {
			log.Printf("expand url error. %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			errMsg, _ := json.Marshal(errorResp{Msg: http.StatusText(http.StatusInternalServerError)})
			w.Write(errMsg)
			return
		} else {
			expandResp, _ := json.Marshal(expandResp{LongURL: expandedURL})
			w.Write(expandResp)
		}
	}
}
