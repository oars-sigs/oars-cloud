package version

import "encoding/json"

var AppVersion = "unknown"
var GitCommit = "unknown"
var BuildTime = "unknown"
var GoVersion = "unknown"

type Version struct {
	AppVersion string
	GitCommit  string
	BuildTime  string
	GoVersion  string
}

func (v *Version) String() string {
	data, _ := json.Marshal(v)
	return string(data)
}

func Get() *Version {
	return &Version{
		AppVersion: AppVersion,
		GitCommit:  GitCommit,
		BuildTime:  BuildTime,
		GoVersion:  GoVersion,
	}
}
