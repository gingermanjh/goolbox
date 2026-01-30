package ffmpeg

import "sort"

// Profile defines encoding settings
type Profile struct {
	Name        string
	Description string
	CRF         string
	Preset      string
	Bitrate     string
	Scale       string
	Audio       bool
}

// BuiltinProfiles contains all available profiles
var BuiltinProfiles = map[string]Profile{
	"standard": {
		Name:        "standard",
		Description: "Default encoding settings",
		CRF:         "18",
		Preset:      "slow",
		Bitrate:     "3000k",
		Scale:       "1920:1080",
		Audio:       false,
	},
}

// DefaultProfile is the profile used when none is specified
const DefaultProfile = "standard"

// GetProfile returns a profile by name
func GetProfile(name string) (Profile, bool) {
	p, ok := BuiltinProfiles[name]
	return p, ok
}

// ListProfiles returns all profiles sorted by name
func ListProfiles() []Profile {
	profiles := make([]Profile, 0, len(BuiltinProfiles))
	for _, p := range BuiltinProfiles {
		profiles = append(profiles, p)
	}
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})
	return profiles
}
