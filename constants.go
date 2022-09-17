package sortedset

const (
	SkiplistMaxLevel  = 32   /* Should be enough for 2^32 elements */
	SkiplistLevelRate = 0.25 /* Skiplist P = 1/4 */
	eps               = 0.00001
	defaultLimit      = int((^uint(0)) >> 1)
)
