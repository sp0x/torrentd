package formatting

import (
	"reflect"
	"testing"
	"time"
)

func TestClearSpaces(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Empty should return empty", args{""}, ""},
		{"Space at both ends should be trimmed", args{"  there's a space\t"}, " there's a space "},
		{"Double spaces should be trimmed", args{"there's  a  space"}, "there's a space"},
		{"Tabs should be replaced with spaces", args{"there's a space\t"}, "there's a space "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeSpace(tt.args.raw); got != tt.want {
				t.Errorf("NormalizeSpace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractAttr(t *testing.T) {
	type args struct {
		uri   string
		param string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Should parse simple query strings", args{"?a=3", "a"}, "3"},
		{"Should extract an argument from a schemeless url", args{"somepath.com/a?a=3", "a"}, "3"},
		{"Should extract an argument from an url with multiple args", args{"somepath.com/a?a=3&B=4", "a"}, "3"},
		{"Shouldn't return anything if the url is empty", args{"", "a"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractAttributeFromQuery(tt.args.uri, tt.args.param); got != tt.want {
				t.Errorf("ExtractAttributeFromQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	type args struct {
		str string
	}
	timeZone, _ := time.LoadLocation("UTC")
	dateWithTime := time.Date(2020, 10, 10, 10, 10, 0, 0, timeZone)
	date := time.Date(2020, 10, 10, 0, 0, 0, 0, timeZone)
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"Should handle cyrillic months", args{"10-Окт-20 10:10"}, dateWithTime},
		{"Should handle cyrillic months without time", args{"10-Окт-20"}, date},
		{"Should parse dates with multiple spaces", args{"\t10-Окт-20  "}, date},
		{"Should parse dates with multiple spaces", args{"\t10-Oct-20  10:10"}, dateWithTime},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatTime(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormatTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSizeStrToBytes(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{"Should handle simple sizes", args{"10mb"}, 10567840},
		{"Should handle sizes with tabs in them", args{"10\t\tmb"}, 10567840},
		{"Should handle sizes with text and tabs", args{"size: 10\t\tmb"}, 10567840},
		{"Should handle sizes with capital unit characters", args{"size: 10\t\tMB\t"}, 10567840},
		{"Should handle cyrillic text in sizes", args{"размер: 10\t\tMB\t"}, 10567840},
		{"Should handle sizes with floating format.", args{"размер: 10.00\t\tMB\t"}, 10567840},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SizeStrToBytes(tt.args.str); got != tt.want {
				t.Errorf("SizeStrToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStripToNumber(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Should handle simple numbers", args{"55"}, "55"},
		{"Should handle numbers with text in them", args{"5а5"}, "55"},
		{"Should handle numbers and text mixed, with whitespace", args{"\t5а5s"}, "55"},
		{"Should handle longer strings", args{"\tddd5а5"}, "55"},
		{"Should handle special characters", args{"\tddd5|5"}, "55"},
		{"Should handle simple floats", args{"\tddd5.5"}, "5.5"},
		{"Should handle simple floats with comma", args{"\tddd5,5"}, "5,5"},
		{"Should handle simple floats with any other characters", args{"\tddd5.x5"}, "5.5"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripToNumber(tt.args.str); got != tt.want {
				t.Errorf("StripToNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_fixMonths(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"should convert cyrillic to english months", args{"Окт"}, "Oct"},
		{"should convert cyrillic to english months", args{"Феб"}, "Feb"},
		{"should convert cyrillic to english months", args{"ОктОкт"}, "OctOct"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixMonths(tt.args.str); got != tt.want {
				t.Errorf("fixMonths() = %v, want %v", got, tt.want)
			}
		})
	}
}
