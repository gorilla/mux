package mux

func Fuzz(data []byte) int {
	_, err := newRouteRegexp(string(data), regexpTypeHost, routeRegexpOptions{})
	if err != nil {
		return 0
	}
	return 1
}
