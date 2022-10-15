package client

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRequestTcp(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "",
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//RequestTcp("127.0.0.1:10030")
			result, err := RequestTcp("127.0.0.1:10030", "admin/info", "", 1500)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			var str []byte
			str, err = json.Marshal(result)
			fmt.Println(string(str))
		})
	}
}
