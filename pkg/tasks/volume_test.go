package tasks

import (
	"testing"
	"time"
)

type testTimespec struct {
	birthTime     time.Time
	changeTime    time.Time
	modTime       time.Time
	hasBirthTime  bool
	hasChangeTime bool
}

func (t testTimespec) ModTime() time.Time { return t.modTime }

func (testTimespec) AccessTime() time.Time { return time.Time{} }

func (t testTimespec) ChangeTime() time.Time { return t.changeTime }

func (t testTimespec) BirthTime() time.Time { return t.birthTime }

func (t testTimespec) HasChangeTime() bool { return t.hasChangeTime }

func (t testTimespec) HasBirthTime() bool { return t.hasBirthTime }

func TestCreationTime(t *testing.T) {
	birthTime := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	changeTime := time.Date(2024, time.January, 3, 3, 4, 5, 0, time.UTC)
	modTime := time.Date(2024, time.January, 4, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name  string
		times testTimespec
		want  time.Time
	}{
		{
			name:  "uses birth time when available",
			times: testTimespec{birthTime: birthTime, changeTime: changeTime, modTime: modTime, hasBirthTime: true, hasChangeTime: true},
			want:  birthTime,
		},
		{
			name:  "uses change time when birth time is unix epoch",
			times: testTimespec{birthTime: time.Unix(0, 0), changeTime: changeTime, modTime: modTime, hasBirthTime: true, hasChangeTime: true},
			want:  changeTime,
		},
		{
			name:  "uses change time when birth time is zero",
			times: testTimespec{changeTime: changeTime, modTime: modTime, hasBirthTime: true, hasChangeTime: true},
			want:  changeTime,
		},
		{
			name:  "uses modification time when birth and change times are unix epoch",
			times: testTimespec{birthTime: time.Unix(0, 0), changeTime: time.Unix(0, 0), modTime: modTime, hasBirthTime: true, hasChangeTime: true},
			want:  modTime,
		},
		{
			name:  "uses modification time when birth and change times are zero",
			times: testTimespec{modTime: modTime, hasBirthTime: true, hasChangeTime: true},
			want:  modTime,
		},
		{
			name:  "uses modification time when birth and change times are unavailable",
			times: testTimespec{modTime: modTime},
			want:  modTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := creationTime(tt.times); !got.Equal(tt.want) {
				t.Errorf("creationTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNeedsCreationTimeRefresh(t *testing.T) {
	validTime := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)

	tests := []struct {
		name      string
		timestamp time.Time
		want      bool
	}{
		{
			name:      "requires refresh for zero time",
			timestamp: time.Time{},
			want:      true,
		},
		{
			name:      "requires refresh for unix epoch",
			timestamp: time.Unix(0, 0),
			want:      true,
		},
		{
			name:      "does not require refresh for valid time",
			timestamp: validTime,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsCreationTimeRefresh(tt.timestamp); got != tt.want {
				t.Errorf("needsCreationTimeRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}
