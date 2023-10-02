package fpc

type predictorClass uint8

const (
	fcmPredictor predictorClass = iota
	dfcmPredictor
)

type predictor interface {
	predict() (predicted uint64)
	update(actual uint64)
	clear()
}

type fcm struct {
	table    []uint64
	size     uint64
	lastHash uint64
}

func newFCM(size uint) *fcm {
	// size must be a power of two
	return &fcm{
		table: make([]uint64, size, size),
		size:  uint64(size),
	}
}

func (f *fcm) hash(actual uint64) uint64 {
	return ((f.lastHash << 6) ^ (actual >> 48)) & (f.size - 1)
}

func (f *fcm) predict() uint64 {
	return f.table[f.lastHash]
}

func (f *fcm) update(actual uint64) {
	f.table[f.lastHash] = actual
	f.lastHash = f.hash(actual)
}

func (f *fcm) clear() {
	clear(f.table)
	f.lastHash = 0
}

type dfcm struct {
	table     []uint64
	size      uint64
	lastHash  uint64
	lastValue uint64
}

func newDFCM(size uint) *dfcm {
	// size must be a power of two
	return &dfcm{
		table: make([]uint64, size, size),
		size:  uint64(size),
	}
}

func (d *dfcm) hash(actual uint64) uint64 {
	return ((d.lastHash << 2) ^ ((actual - d.lastValue) >> 40)) & (d.size - 1)
}

func (d *dfcm) predict() uint64 {
	return d.table[d.lastHash] + d.lastValue
}

func (d *dfcm) update(actual uint64) {
	d.table[d.lastHash] = actual - d.lastValue
	d.lastHash = d.hash(actual)
	d.lastValue = actual
}

func (d *dfcm) clear() {
	clear(d.table)
	d.lastHash = 0
	d.lastValue = 0
}
