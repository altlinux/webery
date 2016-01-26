/*
 * Copyright (C) 2016 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package util

func InSliceString(n string, list []string) bool {
	for _, s := range list {
		if n == s {
			return true
		}
	}
	return false
}
