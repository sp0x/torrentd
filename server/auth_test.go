package server

import (
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer"
	"reflect"
	"testing"
	"text/tabwriter"
)

func TestServer_checkAPIKey(t *testing.T) {
	type fields struct {
		tracker    *indexer.Facade
		tabWriter  *tabwriter.Writer
		config     config.Config
		Port       int
		Hostname   string
		Params     Params
		PathPrefix string
		Password   string
		version    string
	}
	type args struct {
		inputKey string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				tracker:    tt.fields.tracker,
				tabWriter:  tt.fields.tabWriter,
				config:     tt.fields.config,
				Port:       tt.fields.Port,
				Hostname:   tt.fields.Hostname,
				Params:     tt.fields.Params,
				PathPrefix: tt.fields.PathPrefix,
				Password:   tt.fields.Password,
				version:    tt.fields.version,
			}
			if got := s.checkAPIKey(tt.args.inputKey); got != tt.want {
				t.Errorf("checkAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_sharedKey(t *testing.T) {
	type fields struct {
		tracker    *indexer.Facade
		tabWriter  *tabwriter.Writer
		config     config.Config
		Port       int
		Hostname   string
		Params     Params
		PathPrefix string
		Password   string
		version    string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				tracker:    tt.fields.tracker,
				tabWriter:  tt.fields.tabWriter,
				config:     tt.fields.config,
				Port:       tt.fields.Port,
				Hostname:   tt.fields.Hostname,
				Params:     tt.fields.Params,
				PathPrefix: tt.fields.PathPrefix,
				Password:   tt.fields.Password,
				version:    tt.fields.version,
			}
			got, err := s.sharedKey()
			if (err != nil) != tt.wantErr {
				t.Errorf("sharedKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sharedKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}
