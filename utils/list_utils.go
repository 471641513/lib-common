package utils

import (
	"fmt"
	"sort"
	"strings"
)

func IntListJoins(l []int64, sep string) string {
	strList := make([]string, len(l))
	for idx, iface := range l {
		strList[idx] = fmt.Sprintf("%v", iface)
	}
	return strings.Join(strList, sep)
}

type SortInt64 []int64

func (sv SortInt64) Len() int           { return len(sv) }
func (sv SortInt64) Less(i, j int) bool { return sv[i] < sv[j] }
func (sv SortInt64) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }

func IntListIntersect(l1 []int64, l2 []int64) (intersect []int64) {
	if len(l1) == 0 || len(l2) == 0 {
		return
	}
	sort.Sort(SortInt64(l1))
	sort.Sort(SortInt64(l2))

	var idx1, idx2 int
	for {
		if idx1 >= len(l1) || idx2 >= len(l2) {
			break
		}

		if l1[idx1] < l2[idx2] {
			idx1 += 1
		} else if l1[idx1] > l2[idx2] {
			idx2 += 1
		} else {
			intersect = append(intersect, l1[idx1])
			idx1 += 1
			idx2 += 1
		}
	}
	return
}
