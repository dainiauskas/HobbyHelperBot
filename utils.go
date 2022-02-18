package main

import (
	"net/url"

	"mvdan.cc/xurls"
)

func hasLink(v string) (bool, string) {
	rxRelaxed := xurls.Relaxed
	uri := rxRelaxed.FindString(v)

	return uri != "", uri
}

func extractURL(v string) (u *url.URL, q url.Values, err error) {
	u, err = url.Parse(v)
	if err != nil {
		return
	}

	q, err = url.ParseQuery(u.RawQuery)

	return
}
