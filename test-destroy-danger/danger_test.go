package main

import (
	"testing"
)

func TestPut1(t *testing.T) {

        var tests = []struct {
		put string
		putResult string
		vServers string
		groups string
		servers string
        }{
		{"[]","[]","","",""},
	}

	url := "http://192.168.56.20:8080/v1/at2/node/10.255.255.6/rule"
	contentType := "application/json"
	user := "admin"
	pass := "a10"

	for _, v := range tests {
		t.Logf("put=%s putResult=%s vServers=%s groups=%s servers=%s", v.put, v.putResult, v.vServers, v.groups, v.servers)

		buf, errPut := httpPut(url, contentType, user, pass, v.put)
		str := string(buf)

		t.Logf("putResult=%v errPut=%v", str, errPut)

		if errPut != nil {
			t.Errorf("put error: putResult=%v errPut=%v", str, errPut)
		}

		if str != v.putResult {
			t.Errorf("result mismatch: put=%s putResult=%s vServers=%s groups=%s servers=%s", v.put, v.putResult, v.vServers, v.groups, v.servers)
		}
	}

}
