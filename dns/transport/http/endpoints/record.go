package endpoints

import "net/http"

func (e *endpoints) addRecord(w http.ResponseWriter, r *http.Request)          {}
func (e *endpoints) listRecords(w http.ResponseWriter, r *http.Request)        {}
func (e *endpoints) getRecordByDomain(w http.ResponseWriter, r *http.Request)  {}
func (e *endpoints) getRecordByAddress(w http.ResponseWriter, r *http.Request) {}
func (e *endpoints) updateRecord(w http.ResponseWriter, r *http.Request)       {}
func (e *endpoints) deleteRecord(w http.ResponseWriter, r *http.Request)       {}
