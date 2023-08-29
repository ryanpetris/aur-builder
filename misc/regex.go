package misc

import "regexp"

func RegexGetMatchByGroup(re *regexp.Regexp, str string) map[string]string {
	names := re.SubexpNames()
	result := map[string]string{}

	matches := re.FindStringSubmatch(str)

	if matches == nil {
		return nil
	}

	for idx, name := range names {
		result[name] = matches[idx]
	}

	return result
}
