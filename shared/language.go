// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package shared

func Pluralize(n int, singular string, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
