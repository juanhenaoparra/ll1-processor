package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type Grammar map[string][]string

type LL1 struct {
	First  [][]string `json:"first"`
	Follow [][]string `json:"follow"`
}

type LL1Response struct {
	Grammar *Grammar `json:"grammar,omitempty"`
	Result  *LL1     `json:"result,omitempty"`
}

func LL1Process(w http.ResponseWriter, r *http.Request) {
	req := &Grammar{}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		rend(w, r, NewAPIError(http.StatusBadRequest, "corrupted body payload"))
		return
	}

	defer CloseOrLog(r.Body)

	err = json.Unmarshal(bodyBytes, req)
	if err != nil {
		rend(w, r, NewAPIError(http.StatusBadRequest, "invalid body"))
		return
	}

	//TODO: add processing after marshalling
}
