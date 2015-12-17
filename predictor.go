package fpc

type predictorClass uint8

const (
	fcmPredictor predictorClass = iota
	dfcmPredictor
)

type predictor interface {
	bit() uint8
}

type fcm struct{}

func (f *fcm) bit() bool { return false }
