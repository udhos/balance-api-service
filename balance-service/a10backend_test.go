package main

import (
	"strings"
	"testing"
)

func TestDecodingYAML1(t *testing.T) {
	var be backend
	r := strings.NewReader(str1)
	errDec := decodeBackend(r, true, &be)
	if errDec != nil {
		t.Errorf("decode error: %v", errDec)
	}
	if be.BackendName != "s1" {
		t.Errorf("wrong name: %s", be.BackendName)
	}
	if be.BackendAddress != "2.2.2.2" {
		t.Errorf("wrong name: %s", be.BackendAddress)
	}
}

func TestDecodingYAML2(t *testing.T) {
	var be backend
	r := strings.NewReader(str2)
	errDec := decodeBackend(r, true, &be)
	if errDec != nil {
		t.Errorf("decode error: %v", errDec)
	}
	if be.BackendName != "s1" {
		t.Errorf("wrong name: %s", be.BackendName)
	}
	if be.BackendAddress != "2.2.2.2" {
		t.Errorf("wrong name: %s", be.BackendAddress)
	}
}

const str1 = `
virtualservers:
- name: _1.1.1.1_vserver
  address: 1.1.1.1
  virtualports:
  - port: "1111"
    protocol: tcp
    servicegroup: g1
- name: vs1
  address: 10.10.10.10
  virtualports:
  - port: "5555"
    protocol: tcp
    servicegroup: g1
servicegroups:
- name: g1
  protocol: tcp
  members:
  - name: s1
    port: "4444"
  - name: s1
    port: "3333"
- name: g2
  protocol: tcp
  members:
  - name: s1
    port: "3333"
- name: sg1
  protocol: tcp
  members:
  - name: s1
    port: "4444"
backendname: s1
backendaddress: 2.2.2.2
backendports:
- port: "5555"
  protocol: tcp
- port: "4444"
  protocol: tcp
- port: "3333"
  protocol: tcp
`

const str2 = `
backendname: s1
backendaddress: 2.2.2.2
`
