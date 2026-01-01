package libdave

type Codec int

const (
	CodecUnknown Codec = iota
	CodecOpus
	CodecVP8
	CodecVP9
	CodecH264
	CodecH265
	CodecAV1
)
