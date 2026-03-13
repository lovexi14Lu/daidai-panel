package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ParseResult struct {
	Valid       bool
	HasSecond   bool
	Fields      map[string]string
	Description string
	Error       string
}

func Parse(expression string) ParseResult {
	expression = strings.TrimSpace(expression)
	parts := strings.Fields(expression)

	if len(parts) == 5 {
		return parse5Field(parts)
	}
	if len(parts) == 6 {
		return parse6Field(parts)
	}

	return ParseResult{Valid: false, Error: "cron expression must have 5 or 6 fields"}
}

func parse5Field(parts []string) ParseResult {
	fields := map[string]string{
		"minute":      parts[0],
		"hour":        parts[1],
		"day":         parts[2],
		"month":       parts[3],
		"day_of_week": parts[4],
	}

	if err := validateField(parts[0], 0, 59, "minute"); err != "" {
		return ParseResult{Valid: false, Error: err}
	}
	if err := validateField(parts[1], 0, 23, "hour"); err != "" {
		return ParseResult{Valid: false, Error: err}
	}
	if err := validateField(parts[2], 1, 31, "day"); err != "" {
		return ParseResult{Valid: false, Error: err}
	}
	if err := validateField(parts[3], 1, 12, "month"); err != "" {
		return ParseResult{Valid: false, Error: err}
	}
	if err := validateField(parts[4], 0, 6, "day_of_week"); err != "" {
		return ParseResult{Valid: false, Error: err}
	}

	return ParseResult{
		Valid:       true,
		HasSecond:   false,
		Fields:      fields,
		Description: describe(fields, false),
	}
}

func parse6Field(parts []string) ParseResult {
	fields := map[string]string{
		"second":      parts[0],
		"minute":      parts[1],
		"hour":        parts[2],
		"day":         parts[3],
		"month":       parts[4],
		"day_of_week": parts[5],
	}

	if err := validateField(parts[0], 0, 59, "second"); err != "" {
		return ParseResult{Valid: false, Error: err}
	}
	if err := validateField(parts[1], 0, 59, "minute"); err != "" {
		return ParseResult{Valid: false, Error: err}
	}
	if err := validateField(parts[2], 0, 23, "hour"); err != "" {
		return ParseResult{Valid: false, Error: err}
	}
	if err := validateField(parts[3], 1, 31, "day"); err != "" {
		return ParseResult{Valid: false, Error: err}
	}
	if err := validateField(parts[4], 1, 12, "month"); err != "" {
		return ParseResult{Valid: false, Error: err}
	}
	if err := validateField(parts[5], 0, 6, "day_of_week"); err != "" {
		return ParseResult{Valid: false, Error: err}
	}

	return ParseResult{
		Valid:       true,
		HasSecond:   true,
		Fields:      fields,
		Description: describe(fields, true),
	}
}

func validateField(field string, min, max int, name string) string {
	if field == "*" || field == "?" {
		return ""
	}

	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "/") {
			segments := strings.SplitN(part, "/", 2)
			base := segments[0]
			step := segments[1]
			if base != "*" && base != "?" {
				if err := validateRange(base, min, max, name); err != "" {
					return err
				}
			}
			stepVal, err := strconv.Atoi(step)
			if err != nil || stepVal < 1 {
				return fmt.Sprintf("%s 的步长值无效: %s", name, step)
			}
			continue
		}

		if strings.Contains(part, "-") {
			segments := strings.SplitN(part, "-", 2)
			if err := validateRange(segments[0], min, max, name); err != "" {
				return err
			}
			if err := validateRange(segments[1], min, max, name); err != "" {
				return err
			}
			continue
		}

		if part == "L" || part == "W" || strings.Contains(part, "#") {
			continue
		}

		if err := validateRange(part, min, max, name); err != "" {
			return err
		}
	}
	return ""
}

func validateRange(val string, min, max int, name string) string {
	val = strings.ToLower(val)
	monthNames := map[string]int{"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6, "jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12}
	weekNames := map[string]int{"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6}

	if _, ok := monthNames[val]; ok {
		return ""
	}
	if _, ok := weekNames[val]; ok {
		return ""
	}

	num, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Sprintf("%s 的值无效: %s", name, val)
	}
	if num < min || num > max {
		return fmt.Sprintf("%s value %d out of range [%d-%d]", name, num, min, max)
	}
	return ""
}

func describe(fields map[string]string, hasSecond bool) string {
	minute := fields["minute"]
	hour := fields["hour"]
	day := fields["day"]
	month := fields["month"]
	dow := fields["day_of_week"]

	result := ""

	if hasSecond {
		sec := fields["second"]
		if strings.HasPrefix(sec, "*/") {
			result += "每" + strings.TrimPrefix(sec, "*/") + "秒"
			return result
		}
	}

	if strings.HasPrefix(minute, "*/") {
		return "每" + strings.TrimPrefix(minute, "*/") + "分钟"
	}

	if strings.HasPrefix(hour, "*/") {
		return "每" + strings.TrimPrefix(hour, "*/") + "小时"
	}

	if dow != "*" && dow != "?" {
		weekDays := map[string]string{"0": "日", "1": "一", "2": "二", "3": "三", "4": "四", "5": "五", "6": "六", "7": "日"}
		if dow == "1-5" {
			result += "工作日 "
		} else if dow == "0,6" || dow == "6,0" {
			result += "周末 "
		} else {
			result += "每周"
			for _, d := range strings.Split(dow, ",") {
				if name, ok := weekDays[strings.TrimSpace(d)]; ok {
					result += name + ","
				}
			}
			result = strings.TrimSuffix(result, ",") + " "
		}
	} else if month != "*" && month != "?" {
		result += "每年" + month + "月 "
	}

	if day != "*" && day != "?" {
		result += day + "日 "
	}

	if hour != "*" && hour != "?" {
		result += hour + "时"
	}
	if minute != "*" && minute != "?" {
		result += minute + "分"
	}

	if result == "" {
		result = "每分钟"
	}

	return strings.TrimSpace(result)
}

func NextRunTimes(expression string, count int) []time.Time {
	result := Parse(expression)
	if !result.Valid {
		return nil
	}

	times := make([]time.Time, 0, count)
	now := time.Now()

	for i := 0; i < count && i < 1000; i++ {
		next := calculateNext(result.Fields, result.HasSecond, now)
		if next.IsZero() {
			break
		}
		times = append(times, next)
		now = next.Add(time.Second)
	}

	return times
}

func calculateNext(fields map[string]string, hasSecond bool, after time.Time) time.Time {
	t := after.Add(time.Second)
	if !hasSecond {
		t = after.Truncate(time.Minute).Add(time.Minute)
	}

	for i := 0; i < 366*24*60; i++ {
		if matchesFields(fields, hasSecond, t) {
			return t
		}
		if hasSecond {
			t = t.Add(time.Second)
		} else {
			t = t.Add(time.Minute)
		}
	}
	return time.Time{}
}

func matchesFields(fields map[string]string, hasSecond bool, t time.Time) bool {
	if hasSecond {
		if !matchField(fields["second"], t.Second(), 0, 59) {
			return false
		}
	}
	if !matchField(fields["minute"], t.Minute(), 0, 59) {
		return false
	}
	if !matchField(fields["hour"], t.Hour(), 0, 23) {
		return false
	}
	if !matchField(fields["day"], t.Day(), 1, 31) {
		return false
	}
	if !matchField(fields["month"], int(t.Month()), 1, 12) {
		return false
	}
	dow := int(t.Weekday())
	if !matchField(fields["day_of_week"], dow, 0, 6) {
		return false
	}
	return true
}

func matchField(field string, value, min, max int) bool {
	if field == "*" || field == "?" {
		return true
	}

	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)
		if matchPart(part, value, min, max) {
			return true
		}
	}
	return false
}

func matchPart(part string, value, min, max int) bool {
	if strings.Contains(part, "/") {
		segments := strings.SplitN(part, "/", 2)
		step, _ := strconv.Atoi(segments[1])
		if step <= 0 {
			return false
		}
		base := segments[0]
		startVal := min
		if base != "*" && base != "?" {
			if strings.Contains(base, "-") {
				rangeParts := strings.SplitN(base, "-", 2)
				rangeStart, _ := strconv.Atoi(rangeParts[0])
				rangeEnd, _ := strconv.Atoi(rangeParts[1])
				if value < rangeStart || value > rangeEnd {
					return false
				}
				return (value-rangeStart)%step == 0
			}
			startVal, _ = strconv.Atoi(base)
		}
		if value < startVal {
			return false
		}
		return (value-startVal)%step == 0
	}

	if strings.Contains(part, "-") {
		rangeParts := strings.SplitN(part, "-", 2)
		rangeStart, _ := strconv.Atoi(rangeParts[0])
		rangeEnd, _ := strconv.Atoi(rangeParts[1])
		return value >= rangeStart && value <= rangeEnd
	}

	num, err := strconv.Atoi(part)
	if err != nil {
		return false
	}
	return value == num
}

func GetTemplates() []map[string]string {
	return []map[string]string{
		{"name": "每分钟", "expression": "0 * * * * *", "description": "每分钟执行一次", "category": "高频"},
		{"name": "每5分钟", "expression": "0 */5 * * * *", "description": "每5分钟执行一次", "category": "高频"},
		{"name": "每10分钟", "expression": "0 */10 * * * *", "description": "每10分钟执行一次", "category": "高频"},
		{"name": "每15分钟", "expression": "0 */15 * * * *", "description": "每15分钟执行一次", "category": "高频"},
		{"name": "每30分钟", "expression": "0 */30 * * * *", "description": "每30分钟执行一次", "category": "常用"},
		{"name": "每小时", "expression": "0 0 * * * *", "description": "每小时整点执行", "category": "常用"},
		{"name": "每2小时", "expression": "0 0 */2 * * *", "description": "每2小时执行一次", "category": "常用"},
		{"name": "每6小时", "expression": "0 0 */6 * * *", "description": "每6小时执行一次", "category": "常用"},
		{"name": "每天0点", "expression": "0 0 0 * * *", "description": "每天凌晨0点执行", "category": "每天"},
		{"name": "每天6点", "expression": "0 0 6 * * *", "description": "每天早上6点执行", "category": "每天"},
		{"name": "每天9点", "expression": "0 0 9 * * *", "description": "每天上午9点执行", "category": "每天"},
		{"name": "每天12点", "expression": "0 0 12 * * *", "description": "每天中午12点执行", "category": "每天"},
		{"name": "每天18点", "expression": "0 0 18 * * *", "description": "每天下午6点执行", "category": "每天"},
		{"name": "工作日9点", "expression": "0 0 9 * * 1-5", "description": "工作日上午9点执行", "category": "工作日"},
		{"name": "工作日18点", "expression": "0 0 18 * * 1-5", "description": "工作日下午6点执行", "category": "工作日"},
		{"name": "周末10点", "expression": "0 0 10 * * 0,6", "description": "周末上午10点执行", "category": "周末"},
		{"name": "每周一0点", "expression": "0 0 0 * * 1", "description": "每周一凌晨0点执行", "category": "每周"},
		{"name": "每月1日0点", "expression": "0 0 0 1 * *", "description": "每月1日凌晨0点执行", "category": "每月"},
		{"name": "每月15日0点", "expression": "0 0 0 15 * *", "description": "每月15日凌晨0点执行", "category": "每月"},
		{"name": "每10秒", "expression": "*/10 * * * * *", "description": "每10秒执行一次", "category": "秒级"},
		{"name": "每30秒", "expression": "*/30 * * * * *", "description": "每30秒执行一次", "category": "秒级"},
	}
}
