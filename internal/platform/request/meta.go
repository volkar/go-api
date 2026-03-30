package request

import (
	"net"
	"net/http"
	"strconv"

	"github.com/mileusna/useragent"
)

type Metadata struct {
	IP        string
	UserAgent string
	Device    string
	Os        string
	Browser   string
	Location  string
}

/* Get metadata from request */
func GetMetaFromRequest(r *http.Request) Metadata {
	ua := useragent.Parse(r.UserAgent())
	// IP
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	// OS
	os := "Unknown"
	if ua.OS != "" && ua.OSVersion != "" {
		os = ua.OS + " " + ua.OSVersion
	} else if ua.OS != "" {
		os = ua.OS
	}
	// Browser
	b := "Unknown"
	if ua.Name != "" && ua.VersionNo.Major != 0 {
		b = ua.Name + " " + strconv.Itoa(ua.VersionNo.Major)
	} else if ua.Name != "" {
		b = ua.Name
	}
	// Device
	d := "Unknown"
	if ua.Device != "" {
		d = ua.Device
	}
	// Location can be extracted from IP by something like MaxMind GeoLite2
	// Use IP as location for simplicity
	l := ip

	return Metadata{
		IP:        ip,
		UserAgent: r.UserAgent(),
		Device:    d,
		Os:        os,
		Browser:   b,
		Location:  l,
	}
}
