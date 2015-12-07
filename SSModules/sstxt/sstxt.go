package sstxt

import (
	//"fmt"
	//"regexp"
	"strings"
)

func RemoveDoubleSpace(x string) string {

	for strings.Contains(x, "  ") {
		x = strings.Replace(x, "  ", " ", -1)
	}

	x = strings.Trim(x, " ")

	return x

}

func Low(x string) string {
	return strings.ToLower(x)
}

/*
func TransMutation(x string, conf *Config_STR) string {

	var (
		ckl = int(0)
		num = int(0)
	)

	for strings.Contains(x, "  ") {
		x = strings.Replace(x, "  ", " ", -1)
	}

	num = len(conf.TRANS_NAMES)

	for ckl = 0; ckl < num; ckl++ {
		x = strings.Replace(x, conf.TRANS_NAMES[ckl][0], conf.TRANS_NAMES[ckl][1], -1)
		x = strings.Replace(x, strings.ToLower(conf.TRANS_NAMES[ckl][0]), strings.ToLower(conf.TRANS_NAMES[ckl][1]), -1)
	}

	for strings.Contains(x, "  ") {
		x = strings.Replace(x, "  ", " ", -1)
	}

	x = strings.Trim(x, " ")

	return x

}

func PosMutation(x string, conf *Config_STR) string {

	var (
		ckl = int(0)
		num = int(0)
	)

	x = strings.Replace(x, "\"", "", -1)
	x = strings.Replace(x, "'", "", -1)

	for strings.Contains(x, "  ") {
		x = strings.Replace(x, "  ", " ", -1)
	}

	num = len(conf.TRANS_POS)

	for ckl = 0; ckl < num; ckl++ {
		x = strings.Replace(x, conf.TRANS_POS[ckl][0], conf.TRANS_POS[ckl][1], -1)
		x = strings.Replace(x, strings.ToLower(conf.TRANS_POS[ckl][0]), strings.ToLower(conf.TRANS_POS[ckl][1]), -1)
	}

	for strings.Contains(x, "  ") {
		x = strings.Replace(x, "  ", " ", -1)
	}

	x = strings.Trim(x, " ")

	return x

}

func PeopleMutation(x string, mode string) []string {

	var (

		//		y	=	[]string{"","","","",""}
		y []string

		ckl = int(0)
		num = int(0)
	)

	for strings.Contains(x, "  ") {
		x = strings.Replace(x, "  ", " ", -1)
	}

	y = strings.Split(strings.Trim(strings.ToLower(x), " "), " ")

	num = len(y)

	for ckl = 0; ckl < num; ckl++ {
		if mode == "RUS" {
			y[ckl] = fmt.Sprintf("%s%s", strings.ToUpper(y[ckl][:2]), y[ckl][2:])
		} else {
			y[ckl] = fmt.Sprintf("%s%s", strings.ToUpper(y[ckl][:1]), y[ckl][1:])
		}
	}

	return y

}

func TextMutation(x string) string {

	for strings.Contains(x, "  ") {
		x = strings.Replace(x, "  ", " ", -1)
	}

	x = strings.Trim(x, " ")

	return x

}

func PhoneMutation(x string) string {
	cleanNumberRegExp := regexp.MustCompile(`[^0-9\+]`)
	return cleanNumberRegExp.ReplaceAllLiteralString(x, "")
}

func NameMutation(x string) string {
	return TextMutation(strings.ToLower(x))
}

*/
