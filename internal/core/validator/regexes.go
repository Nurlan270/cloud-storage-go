package validator

import "regexp"

var (
	username1Regex = regexp.MustCompile(`^[a-zA-Z0-9._]+$`)
	username2Regex = regexp.MustCompile(`[a-zA-Z]`)

	pathRegex = regexp.MustCompile(
		`^/?[A-Za-z0-9._-](?:[A-Za-z0-9._ -]*[A-Za-z0-9._-])?(?:/[A-Za-z0-9._-](?:[A-Za-z0-9._ -]*[A-Za-z0-9._-])?)*/?$`,
	)
)
