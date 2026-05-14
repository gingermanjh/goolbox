package ffmpeg

import "testing"

func TestGetProfile(t *testing.T) {
	t.Run("standard exists with speed 1.0", func(t *testing.T) {
		p, ok := GetProfile("standard")
		if !ok {
			t.Fatal("standard profile should exist")
		}
		if p.Speed != 1.0 {
			t.Errorf("standard.Speed = %v, want 1.0", p.Speed)
		}
	})

	t.Run("2x exists with speed 2.0", func(t *testing.T) {
		p, ok := GetProfile("2x")
		if !ok {
			t.Fatal("2x profile should exist")
		}
		if p.Speed != 2.0 {
			t.Errorf("2x.Speed = %v, want 2.0", p.Speed)
		}
	})

	t.Run("unknown profile returns ok=false", func(t *testing.T) {
		_, ok := GetProfile("does-not-exist")
		if ok {
			t.Error("unknown profile should not be found")
		}
	})
}

func TestListProfilesIsSorted(t *testing.T) {
	profiles := ListProfiles()
	if len(profiles) < 2 {
		t.Skip("need at least 2 profiles to verify sort order")
	}
	for i := 1; i < len(profiles); i++ {
		if profiles[i-1].Name > profiles[i].Name {
			t.Errorf("profiles not sorted: %q came before %q", profiles[i-1].Name, profiles[i].Name)
		}
	}
}

func TestDefaultProfileExists(t *testing.T) {
	if _, ok := GetProfile(DefaultProfile); !ok {
		t.Errorf("DefaultProfile %q must exist in BuiltinProfiles", DefaultProfile)
	}
}
