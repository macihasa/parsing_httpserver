package xmlparsers_test

import (
	"testing"

	"github.com/macihasa/parsing_httpserver/pkg/xmlparsers"
)

func BenchmarkDCECstmsMsg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		xmlparsers.DCECustomsMsg("TestOutput", nil, nil)
	}
}
