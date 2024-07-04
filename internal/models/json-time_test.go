package models

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestJSONTime_MarshalJSON(t *testing.T) {
	testTime, _ := time.Parse("2006-01-02 15:04:05", "2009-11-10 23:00:00")
	tests := []struct {
		name    string
		t       JSONTime
		want    []byte
		wantErr bool
	}{
		{
			name:    "Format time",
			t:       JSONTime(testTime),
			want:    []byte(fmt.Sprintf(`%q`, "2009-11-10T23:00:00Z")),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.t.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, string(got), string(tt.want)) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONTime_UnmarshalJSON(t *testing.T) {
	testTime, _ := time.Parse("2006-01-02T15:04:05Z", "2009-11-10T23:00:00Z")
	type args struct {
		value []byte
	}
	tests := []struct {
		name    string
		t       JSONTime
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Parse time",
			t:    JSONTime(testTime),
			args: args{
				value: []byte("\"2009-11-10T23:00:00Z\""),
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.t.UnmarshalJSON(tt.args.value), fmt.Sprintf("UnmarshalJSON(%v)", tt.args.value))
		})
	}
}
