package config

import (
	"encoding/json"
	"time"

	"github.com/jansemmelink/log"
)

//dateFormats are only used in json.Unmarshal of a json.Date field
var dateFormats = []string{"2006-01-02", "20060102", "2006-01-02T15:04:05-0700", "20060102150405-0700"}

//AddDateFormat to add a new format to accept in parsing
func AddDateFormat(f string) {
	//do not add duplicates
	for _, e := range dateFormats {
		if e == f {
			return
		}
	}
	dateFormats = append(dateFormats, f)
}

//SetDateFormats replaces all existing formats with this list
func SetDateFormats(f []string) {
	dateFormats = f
}

//DateFormats return set of formats accepted in parsing
func DateFormats() []string {
	return dateFormats
}

//Date wraps the golang time.Time to implement unmarshaling and marshalling as CCYY-MM-DD
type Date struct {
	time.Time
}

//MarshalJSON is used to format the date string in JSON
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return json.Marshal("")
	}
	return json.Marshal(d.Format("2006-01-02"))
}

//UnmarshalJSON parses any date-time as good it can
func (d *Date) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		//parse any of the below formats - in order of preference
		for _, format := range dateFormats {
			var err error
			d.Time, err = time.Parse(format, value)
			if err == nil {
				return nil
			}
		}
		return log.Wrapf(nil, "cannot parse \"%.100s\" as a date", value)
	default:
		return log.Wrapf(nil, "invalid date \"%.100v\"", value)
	}
} //Date.UnmarshalJSON
