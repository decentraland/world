package validation

import "regexp"

var re = regexp.MustCompile("^[a-fA-F0-9]{66}$")

func ValidateEphemeralKey(input string) bool {
	return re.Match([]byte(input))
}
