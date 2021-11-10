package models

import (
	"fmt"
	"strings"
	"time"
)

// Device represents data for devices used to take photos such as cameras and
// phones
type Device struct {
	ID      int
	Name    string
	IconURL string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// IconURLSrc handles the special case for local running of the server where
// file:// urls are converted into those for a local static server so they can
// be displayed in browsers
func (d *Device) IconURLSrc() string {
	if strings.HasPrefix(d.IconURL, "file://") {
		index := strings.Index(d.IconURL, "/device_icons")
		if index > 0 {
			return fmt.Sprintf(
				"http://localhost:8000/%s",
				strings.TrimPrefix(d.IconURL[index:len(d.IconURL)], "/"),
			)
		}
	}

	return d.IconURL
}
