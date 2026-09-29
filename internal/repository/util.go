package repository

import "strconv"

// placeholder returns a Postgres positional parameter, e.g. placeholder(3) -> "$3".
func placeholder(n int) string {
	return "$" + strconv.Itoa(n)
}
