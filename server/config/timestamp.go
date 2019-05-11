package config

import (
	"encoding/json"
	"time"

	"github.com/jansemmelink/log"
)

//timestampFormats are only used in json.Unmarshal of a json.Timestamp field
var timestampFormats = []string{"2006-01-02T15:04:05-0700", "20060102150405-0700", "2006-01-02", "20060102"}

//AddTimeFormat to add a new format to accept in parsing
func AddTimeFormat(f string) {
	//do not add duplicates
	for _, e := range timestampFormats {
		if e == f {
			return
		}
	}
	timestampFormats = append(timestampFormats, f)
}

//SetTimeFormats replaces all existing formats with this list
func SetTimeFormats(f []string) {
	timestampFormats = f
}

//TimeFormats return set of formats accepted in parsing
func TimeFormats() []string {
	return timestampFormats
}

//Timestamp wraps the golang time.Time to implement unmarshaling and marshalling with consistent formatting and not marshalling zero values
type Timestamp struct {
	time.Time
}

//MarshalJSON is used to format the timestamp string in JSON
func (d Timestamp) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return json.Marshal("")
	}
	return json.Marshal(d.Format("2006-01-02T15:04:05-0700"))
}

//UnmarshalJSON parses any date-time as good it can
func (d *Timestamp) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		//parse any of the below formats - in order of preference
		for _, format := range timestampFormats {
			var err error
			d.Time, err = time.Parse(format, value)
			if err == nil {
				return nil
			}
			//log.Debugf("%s != %s: %v", value, format, err)
		}
		return log.Wrapf(nil, "cannot parse \"%.100s\" as a timestamp using any of %v", value, timestampFormats)
	default:
		return log.Wrapf(nil, "invalid timestamp \"%.100v\"", value)
	}
} //Timestamp.UnmarshalJSON
