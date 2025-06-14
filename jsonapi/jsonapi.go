package jsonapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"mailinglist/mdb"
	"net/http"
)



func setJsonHeader(w http.ResponseWriter) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
}

func fromJson[T any](body io.Reader, target T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	json.Unmarshal(buf.Bytes(), &target)
}

func returnJson[T any](w http.ResponseWriter, withData func() (T, error)){
	setJsonHeader(w)
	data , serverErr := withData()
	if serverErr != nil {
		w.WriteHeader(500)
		serverErrJson , err := json.Marshal(&serverError)
		if err != nil {
			log.Print(err)
			return 
		}
		w.Write(serverErrJson)
		return
	}

	dataJson , err := jsonMarshal(&data)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}

	w.Write(dataJson)
}

func returnErr(w http.ResponseWriter, err error, code int) {
	returnJson(w, func() (interface{}, error) {
		errorMessage := struct {
			Err string
		}{
			Err: err.Error(),
		}
		w.WriteHeader(code)
		return errorMessage, nil
	})
}

func CreateEmail(db *sql.DB) http.Handler{
	return http.HandlerFunc(func (w http.ResponseWriter, req *http.Request){
		if req.Method != "POST" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.CreateEmail(db, entry.Email); err!= nil {
			returnErr(w, err, 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON CreateEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
	})
	})

}

func GetEmail(db *sql.DB) http.Handler{
	return http.HandlerFunc(func (w http.ResponseWriter, req *http.Request){
		if req.Method != "GET" {
			return
		}
		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)


		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON GetEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
	})
	})

}

func UpdateEmail(db *sql.DB) http.Handler{
	return http.HandlerFunc(func (w http.ResponseWriter, req *http.Request){
		if req.Method != "PUT" {
			return
		}
			entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.UpdateEmail(db, entry.Email); err!= nil {
			returnErr(w, err, 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON UpdateEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
	})
	})

}

func Serve(db *sql.DB, bind String) {
	http.Handle("email/create", CreateEmail(db))
	http.Handle("email/get", GetEmail(db))
	http.Handle("email/get_batch", GetEmailBatch(db))
	http.Handle("email/update", UpdateEmail(db))
	http.Handle("email/delete", DeleteEmail(db))
}