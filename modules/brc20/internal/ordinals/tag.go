package ordinals

// Tags represent data fields in a runestone. Unrecognized odd tags are ignored. Unrecognized even tags produce a cenotaph.
type Tag uint8

var (
	TagBody    = Tag(0)
	TagPointer = Tag(2)
	// TagUnbound is unrecognized
	TagUnbound = Tag(66)

	TagContentType     = Tag(1)
	TagParent          = Tag(3)
	TagMetadata        = Tag(5)
	TagMetaprotocol    = Tag(7)
	TagContentEncoding = Tag(9)
	TagDelegate        = Tag(11)
	// TagNop is unrecognized
	TagNop = Tag(255)
)

var allTags = map[Tag]struct{}{
	TagPointer: {},

	TagContentType:     {},
	TagParent:          {},
	TagMetadata:        {},
	TagMetaprotocol:    {},
	TagContentEncoding: {},
	TagDelegate:        {},
}

func (t Tag) IsValid() bool {
	_, ok := allTags[t]
	return ok
}

var chunkedTags = map[Tag]struct{}{
	TagMetadata: {},
}

func (t Tag) IsChunked() bool {
	_, ok := chunkedTags[t]
	return ok
}

func (t Tag) Bytes() []byte {
	if t == TagBody {
		return []byte{} // body tag is empty data push
	}
	return []byte{byte(t)}
}

func ParseTag(input interface{}) (Tag, error) {
	switch input := input.(type) {
	case Tag:
		return input, nil
	case int:
		return Tag(input), nil
	case int8:
		return Tag(input), nil
	case int16:
		return Tag(input), nil
	case int32:
		return Tag(input), nil
	case int64:
		return Tag(input), nil
	case uint:
		return Tag(input), nil
	case uint8:
		return Tag(input), nil
	case uint16:
		return Tag(input), nil
	case uint32:
		return Tag(input), nil
	case uint64:
		return Tag(input), nil
	default:
		panic("invalid tag input type")
	}
}
