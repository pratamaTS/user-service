package helpers

import "time"

const POSDateTimeLayout = "02-01-2006 15:04:05"

var jakartaLoc *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		jakartaLoc = time.FixedZone("WIB", 7*3600)
		return
	}
	jakartaLoc = loc
}

func FormatPOSDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(jakartaLoc).Format(POSDateTimeLayout)
}

func FormatPOSUnix(sec int64) string {
	if sec <= 0 {
		return ""
	}
	return time.Unix(sec, 0).In(jakartaLoc).Format(POSDateTimeLayout)
}
