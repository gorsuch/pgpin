package main

import (
	"bytes"
	"encoding/json"
	"github.com/darkhelmet/env"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Setup and teardown.

func clear() {
	_, err := dataConn.Exec("DELETE from pins")
	must(err)
	_, err = dataConn.Exec("DELETE from dbs")
	must(err)
}

func init() {
	if env.StringDefault("TEST_LOGS", "false") != "true" {
		log.SetOutput(ioutil.Discard)
	}
	os.Setenv("DATABASE_URL", env.String("TEST_DATABASE_URL"))
	dataStart()
	clear()
	webBuild()
}

// Helpers.

func mustRequest(method, url string, body io.Reader) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, body)
	must(err)
	res := httptest.NewRecorder()
	webMux.ServeHTTP(res, req)
	return res
}

func mustDecode(res *httptest.ResponseRecorder, data interface{}) {
	must(json.NewDecoder(res.Body).Decode(data))
}

func mustDataDbAdd(name string, url string) *Db {
	db, err := dataDbAdd(name, url)
	must(err)
	return db
}

func mustDataPinCreate(dbId string, name string, query string) *Pin {
	pin, err := dataPinCreate(dbId, name, query)
	must(err)
	return pin
}

func mustWorkerTick() {
	_, err := workerTick()
	must(err)
}

func mustCanonicalizeJson(in []byte) string {
	data := make([]interface{}, 0)
	must(json.Unmarshal(in, &data))
	out, err := json.Marshal(data)
	must(err)
	return string(out)

}

// DB endpoints.

func TestDbAdd(t *testing.T) {
	defer clear()
	b := bytes.NewReader([]byte(`{"name": "dbs-1", "url": "postgres://u:p@h:1234/d-1"}`))
	res := mustRequest("POST", "/dbs", b)
	assert.Equal(t, 201, res.Code)
	dbOut := &Db{}
	mustDecode(res, dbOut)
	assert.Equal(t, "dbs-1", dbOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1", dbOut.Url)
	assert.NotEmpty(t, dbOut.Id)
	assert.WithinDuration(t, time.Now(), dbOut.AddedAt, 3*time.Second)
}

func TestDbAddDuplicateName(t *testing.T) {
	defer clear()
	mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	b := bytes.NewReader([]byte(`{"name": "dbs-1", "url": "postgres://u:p@h:1234/d-other"}`))
	res := mustRequest("POST", "/dbs", b)
	assert.Equal(t, 400, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "duplicate-db-name", data["id"])
	assert.NotEmpty(t, data["message"])
}

func TestDbGet(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	res := mustRequest("GET", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	dbOut := &Db{}
	mustDecode(res, dbOut)
	assert.Equal(t, dbIn.Id, dbOut.Id)
	assert.Equal(t, "dbs-1", dbOut.Name)
	assert.Equal(t, "postgres://u:p@h:1234/d-1", dbOut.Url)
	assert.WithinDuration(t, time.Now(), dbOut.AddedAt, 3*time.Second)
	assert.WithinDuration(t, time.Now(), dbOut.UpdatedAt, 3*time.Second)
}

func TestDbRename(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	b := bytes.NewReader([]byte(`{"name": "dbs-1a"}`))
	res := mustRequest("PUT", "/dbs/"+dbIn.Id, b)
	assert.Equal(t, 200, res.Code)
	dbPutOut := &Db{}
	mustDecode(res, dbPutOut)
	assert.Equal(t, "dbs-1a", dbPutOut.Name)
	res = mustRequest("GET", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	dbGetOut := &Db{}
	mustDecode(res, dbGetOut)
	assert.Equal(t, "dbs-1a", dbGetOut.Name)
	assert.True(t, dbGetOut.UpdatedAt.After(dbIn.UpdatedAt))
}

func TestDbRemove(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	res := mustRequest("DELETE", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	res = mustRequest("GET", "/dbs/"+dbIn.Id, nil)
	assert.Equal(t, 404, res.Code)
}

func TestDbRemoveWithPins(t *testing.T) {
	defer clear()
	db := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	mustDataPinCreate(db.Id, "pins-1", "select count (*) from pins")
	res := mustRequest("DELETE", "/dbs/"+db.Id, nil)
	assert.Equal(t, 400, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "removing-db-with-pins", data["id"])
	assert.NotEmpty(t, data["message"])
}

func TestDbList(t *testing.T) {
	defer clear()
	dbIn1 := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	dbIn2 := mustDataDbAdd("dbs-2", "postgres://u:p@h:1234/d-2")
	_, err := dataDbRemove(dbIn2.Id)
	must(err)
	res := mustRequest("GET", "/dbs", nil)
	assert.Equal(t, 200, res.Code)
	dbsOut := []*Db{}
	mustDecode(res, &dbsOut)
	assert.Equal(t, len(dbsOut), 1)
	assert.Equal(t, dbIn1.Id, dbsOut[0].Id)
	assert.Equal(t, "dbs-1", dbsOut[0].Name)
}

// Pin endpoints.

func TestPinCreateAndGet(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	b := bytes.NewReader([]byte(`{"name": "pin-1", "db_id": "` + dbIn.Id + `", "query": "select count(*) from pins"}`))
	res := mustRequest("POST", "/pins", b)
	assert.Equal(t, 201, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	mustWorkerTick()
	res = mustRequest("GET", "/pins/"+pinOut.Id, nil)
	assert.Equal(t, 200, res.Code)
	mustDecode(res, pinOut)
	assert.NotEmpty(t, pinOut.Id)
	assert.Equal(t, "pin-1", pinOut.Name)
	assert.Equal(t, dbIn.Id, pinOut.DbId)
	assert.Equal(t, "select count(*) from pins", pinOut.Query)
	assert.WithinDuration(t, time.Now(), pinOut.CreatedAt, 3*time.Second)
	assert.True(t, pinOut.QueryStartedAt.After(pinOut.CreatedAt))
	assert.True(t, pinOut.QueryFinishedAt.After(*pinOut.QueryStartedAt))
	assert.Equal(t, `["count"]`, mustCanonicalizeJson(pinOut.ResultsFields))
	assert.Equal(t, `[[1]]`, mustCanonicalizeJson(pinOut.ResultsRows))
	assert.Nil(t, pinOut.ResultsError)
}

func TestPinCreateDuplicateName(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	mustDataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	b := bytes.NewReader([]byte(`{"name": "pins-1", "db_id": "` + dbIn.Id + `", "query": "select count(*) from pins"}`))
	res := mustRequest("POST", "/pins", b)
	assert.Equal(t, 400, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "duplicate-pin-name", data["id"])
	assert.NotEmpty(t, data["message"])
}

func TestPinRename(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", "postgres://u:p@h:1234/d-1")
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	b := bytes.NewReader([]byte(`{"name": "pins-1a"}`))
	res := mustRequest("PUT", "/pins/"+pinIn.Id, b)
	assert.Equal(t, 200, res.Code)
	pinPutOut := &Pin{}
	mustDecode(res, pinPutOut)
	assert.Equal(t, "pins-1a", pinPutOut.Name)
	res = mustRequest("GET", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinGetOut := &Pin{}
	mustDecode(res, pinGetOut)
	assert.Equal(t, "pins-1a", pinGetOut.Name)
	assert.True(t, pinGetOut.UpdatedAt.After(pinIn.UpdatedAt))
}

func TestPinMultipleColumns(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select name, query from pins")
	mustWorkerTick()
	res := mustRequest("GET", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, `["name","query"]`, mustCanonicalizeJson(pinOut.ResultsFields))
	assert.Equal(t, `[["pins-1","select name, query from pins"]]`, mustCanonicalizeJson(pinOut.ResultsRows))
	assert.Nil(t, pinOut.ResultsError)
}

func TestPinTooManyRows(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	pinInId := mustDataPinCreate(dbIn.Id, "pins-1", "select generate_series(0, 10000)")
	mustWorkerTick()
	res := mustRequest("GET", "/pins/"+pinInId.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "null", string([]byte(pinOut.ResultsFields)))
	assert.Equal(t, "null", string([]byte(pinOut.ResultsRows)))
	assert.Equal(t, "too many rows in query results", *pinOut.ResultsError)
}

func TestPinDelete(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	pinIn := mustDataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	res := mustRequest("DELETE", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 200, res.Code)
	pinOut := &Pin{}
	mustDecode(res, pinOut)
	assert.Equal(t, "pins-1", pinOut.Name)
	res = mustRequest("GET", "/pins/"+pinIn.Id, nil)
	assert.Equal(t, 404, res.Code)
}

func TestPinList(t *testing.T) {
	defer clear()
	dbIn := mustDataDbAdd("dbs-1", env.String("DATABASE_URL"))
	pinIn1 := mustDataPinCreate(dbIn.Id, "pins-1", "select count(*) from pins")
	pinIn2 := mustDataPinCreate(dbIn.Id, "pins-2", "select * from pins")
	_, err := dataPinDelete(pinIn1.Id)
	must(err)
	res := mustRequest("GET", "/pins", nil)
	assert.Equal(t, 200, res.Code)
	pinsOut := []*Pin{}
	mustDecode(res, &pinsOut)
	assert.Equal(t, len(pinsOut), 1)
	assert.Equal(t, pinIn2.Id, pinsOut[0].Id)
	assert.Equal(t, "pins-2", pinsOut[0].Name)
}

// Misc endpoints.

func TestNotFound(t *testing.T) {
	res := mustRequest("GET", "/wat", nil)
	assert.Equal(t, 404, res.Code)
	data := make(map[string]string)
	mustDecode(res, &data)
	assert.Equal(t, "not-found", data["id"])
	assert.Equal(t, "not found", data["message"])
}

func TestStatus(t *testing.T) {
	res := mustRequest("GET", "/status", nil)
	assert.Equal(t, 200, res.Code)
	status := &Status{}
	mustDecode(res, status)
	assert.Equal(t, "ok", status.Message)
}

// Worker behaviour.

func TestWorkerNoop(t *testing.T) {
	processed, err := workerTick()
	assert.False(t, processed)
	assert.Nil(t, err)
}
