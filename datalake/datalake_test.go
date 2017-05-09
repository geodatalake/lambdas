// Copyright 2017 Blacksky. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package datalake

import (
	"fmt"
	"testing"
)

func check(s1, s2 string, t *testing.T) {
	if s1 != s2 {
		t.Errorf(fmt.Sprintf("Expected %s, received %s", s1, s2))
	}
}

func makeDlc() *DataLakeConn {
	return NewDataLakeConn("sampleKey", "sampleSecret", "sampleEndpoint")
}

func TestKdate(t *testing.T) {
	expected := "Fnh2W1PqFTbYYfpLHSwHOykknFLhZwSRAHv1EWAyNhQ="
	dateString := "20170505"
	value := makeDlc().kDate(dateString)
	check(expected, value, t)
}

func TestSignature(t *testing.T) {
	expected := "nt5+Go0ljBsaLFn8mr5x0lLWrodI+3K9VS9Xy//E4Pk="
	dateString := "20170505"
	value := makeDlc().signature(dateString)
	check(expected, value, t)
}

func TestAuthKey(t *testing.T) {
	expected := "c2FtcGxlS2V5Om50NStHbzBsakJzYUxGbjhtcjV4MGxMV3JvZEkrM0s5VlM5WHkvL0U0UGs9"
	dateString := "20170505"
	value := makeDlc().authKey(dateString)
	check(expected, value, t)
}

func TestAuthWDate(t *testing.T) {
	expected := "ak:c2FtcGxlS2V5Om50NStHbzBsakJzYUxGbjhtcjV4MGxMV3JvZEkrM0s5VlM5WHkvL0U0UGs9"
	dateString := "20170505"
	value := makeDlc().AuthWDate(dateString)
	check(expected, value, t)
}
