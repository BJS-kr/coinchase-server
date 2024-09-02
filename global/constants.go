package global

const OWNER_SYSTEM = "system"
const STATUS_TYPE_USER = "u"
const STATUS_TYPE_COIN = "c"

const (
	MAP_SIZE        int32 = 20
	COIN_COUNT      int   = 10
	ITEM_COUNT      int   = 2
	EFFECT_DURATION int   = 10
)

const (
	UNKNOWN = iota
	USER
	COIN
	ITEM_LENGTHEN_VISION
	ITEM_SHORTEN_VISION
	GROUND
)

const (
	UNKNOWN_EFFECT = iota
	NONE           = 2
	LENGTHEN       = 4
	SHORTEN        = 1
)
