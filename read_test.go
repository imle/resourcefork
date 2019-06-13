package resourcefork

import (
	"reflect"
	"testing"
)

func TestReadResourceForkFromBytes(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *ResourceFork
		wantErr bool
	}{
		{
			name:    "empty bytes",
			args:    args{b: []byte{}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid headers",
			args:    args{b: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadResourceForkFromBytes(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadResourceForkFromBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadResourceForkFromBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadResourceForkFromPath(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name    string
		args    args
		want    *ResourceFork
		wantErr bool
	}{
		{
			name: "",
			args: args{p: "./test.ndat"},
			want: &ResourceFork{
				Resources: map[string]map[uint16]Resource{
					"chär": {128: Resource{Type: "chär", ID: 128, Name: ".Trader", Data: []byte{0, 0, 1, 106, 0, 0, 78, 32, 0, 158, 0, 158, 0, 156, 0, 150, 0, 154, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 200, 32, 8, 32, 9, 32, 10, 255, 255, 0, 45, 0, 45, 0, 45, 255, 255, 255, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 23, 0, 6, 4, 131, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}}},
					"csüm": {128: Resource{Type: "csüm", ID: 128, Name: "", Data: []byte{0, 0, 0, 4}}},
					"dsïg": {128: Resource{Type: "dsïg", ID: 128, Name: "", Data: []byte{0, 0, 0, 4}}},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadResourceForkFromPath(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadResourceForkFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadResourceForkFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decodeMacRoman(t *testing.T) {
	type args struct {
		macRomanByteString []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "char test",
			args: args{
				macRomanByteString: []byte{99, 104, 138, 114},
			},
			want: "chär",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeMacRoman(tt.args.macRomanByteString); got != tt.want {
				t.Errorf("decodeMacRoman() = %v, want %v", got, tt.want)
			}
		})
	}
}
