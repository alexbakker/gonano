package store

func uncheckedKindToPrefix(kind UncheckedKind) byte {
	switch kind {
	case UncheckedKindPrevious:
		return idPrefixUncheckedBlockPrevious
	case UncheckedKindSource:
		return idPrefixUncheckedBlockSource
	default:
		panic("bad unchecked block kind")
	}
}
