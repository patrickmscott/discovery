package discovery

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestSimpleStruct(t *testing.T) {
	var watch WatchMessage
	var msg interface{} = &watch
	var decoder *json.Decoder = json.NewDecoder(
		strings.NewReader("{\"groups\":[\"a\",\"b\"]}"))
	err := decoder.Decode(&msg)
	if err != nil {
		t.Error(err)
	}
	if watch.Groups[0] != "a" || watch.Groups[1] != "b" {
		t.Error(watch.Groups)
	}

	var groups []string
	decoder = json.NewDecoder(strings.NewReader("[\"a\",\"b\"]"))
	err = decoder.Decode(&groups)
	if err != nil {
		t.Error(err)
	}
	if len(groups) != 2 || groups[0] != "a" || groups[1] != "b" {
		t.Error(groups)
	}
}

func TestJoinMessageJson(t *testing.T) {
	var join JoinMessage
	var msg interface{} = &join
	decoder := json.NewDecoder(strings.NewReader(`
		{"host": "host", "Port":8080, "Group":"my_service"}
		{"host": "host", "port":8080, "group":"my_service", "extra": true}
		{"host": "host", "port":8080, "group":"my_service", "customData": "data"}`))

	for {
		err := decoder.Decode(&msg)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if join.Port != 8080 || join.Group != "my_service" {
			t.Fatal()
		}
	}

	data := []string{
		`{port:8080, group:my_service}`,
		`{"port":"1", "group":"my_service"}`,
		`{"port":0xff0000, "group":"my_service"}`,
		`{"port":8080, "group": {}}`,
		`{"host":123}`}
	for i := 0; i < len(data); i++ {
		decoder = json.NewDecoder(strings.NewReader(data[i]))
		err := decoder.Decode(&msg)
		if err == nil {
			t.Fatal("Should have failed")
		}
		if err == io.EOF {
			break
		}
	}
}
