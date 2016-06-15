package manager

import "strings"

func GetUrlParams(mainpath, subpath string) []string {
	url := mainpath[len(subpath):]
	url = strings.Trim(url, URL_PATH_SEPARATOR)
	return strings.Split(url, URL_PATH_SEPARATOR)
}
