package misc

import "regexp"

func RegexGetMatchByGroup(re *regexp.Regexp, str string) map[string]string {
	match := re.FindStringSubmatch(str)

	return RegexMapMatchByGroup(re, match)
}

func RegexMapMatchByGroup(re *regexp.Regexp, match []string) map[string]string {
	if match == nil {
		return nil
	}

	names := re.SubexpNames()
	result := map[string]string{}

	for idx, name := range names {
		result[name] = match[idx]
	}

	return result
}
