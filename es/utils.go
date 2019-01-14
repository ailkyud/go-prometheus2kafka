package es

import (
	"regexp"
	"strings"
	"time"

	"github.com/hongxincn/promexp/node2es/config"
)

func GetIndex(timestamp int64) string {
	index := config.Config.Es.Index
	re := regexp.MustCompile(`\$\{\+([^\}]*)\}`)
	if !re.MatchString(index) {
		return index
	}
	datePattern := re.FindStringSubmatch(index)
	dateStr := getIndexSubDate(timestamp, datePattern[1])
	return re.ReplaceAllString(index, dateStr)
}

func getIndexSubDate(timestamp int64, formatString string) string {
	t := time.Unix(timestamp, 0)

	fs := strings.Replace(formatString, "yyyy", "2006", 1)
	fs = strings.Replace(fs, "MM", "01", 1)
	fs = strings.Replace(fs, "dd", "02", 1)

	dayString := t.Format(fs)
	return dayString
}
