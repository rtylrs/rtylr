package pagination

const maxInt64 = int64(^uint64(0) >> 1)

type Policy struct {
	DefaultPageSize int64
	MaxPageSize     int64
	MaxOffset       int64
}

type Page struct {
	Number int64
	Size   int64
	Offset int64
}

func (p Policy) Normalize(number, size int64) Page {
	if p.DefaultPageSize < 1 {
		p.DefaultPageSize = 1
	}
	if p.MaxPageSize < p.DefaultPageSize {
		p.MaxPageSize = p.DefaultPageSize
	}
	if number < 1 {
		number = 1
	}
	if size < 1 {
		size = p.DefaultPageSize
	}
	if size > p.MaxPageSize {
		size = p.MaxPageSize
	}

	maxOffset := p.MaxOffset
	if maxOffset <= 0 {
		maxOffset = maxInt64
	}
	var offset int64
	if number > 1 && number-1 > maxOffset/size {
		offset = maxOffset
	} else {
		offset = (number - 1) * size
	}
	return Page{Number: number, Size: size, Offset: offset}
}

func SliceStart(page, size int64, length int) int {
	if page <= 1 || size <= 0 || length <= 0 {
		return 0
	}
	if page-1 > int64(length)/size {
		return length
	}
	return int((page - 1) * size)
}
