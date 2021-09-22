package api

import (
	"reflect"
	"testing"

	"github.com/juruen/rmapi/transport"
)

func TestAuthHttpCtx(t *testing.T) {
	type args struct {
		reAuth         bool
		nonInteractive bool
	}
	tests := []struct {
		name string
		args args
		want *transport.HttpClientCtx
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AuthHttpCtx(tt.args.reAuth, tt.args.nonInteractive); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthHttpCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readCode(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readCode(); got != tt.want {
				t.Errorf("readCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newDeviceToken(t *testing.T) {
	type args struct {
		http *transport.HttpClientCtx
		code string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newDeviceToken(tt.args.http, tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("newDeviceToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("newDeviceToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newUserToken(t *testing.T) {
	type args struct {
		http *transport.HttpClientCtx
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newUserToken(tt.args.http)
			if (err != nil) != tt.wantErr {
				t.Errorf("newUserToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("newUserToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
