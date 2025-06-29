package compress

type config struct {
	compressMeta bool
	precision    uint
}

func defaultConfig() *config {
	return &config{
		compressMeta: false,
		precision:    256,
	}
}

func (cfg *config) apply(options []ArithmeticCompressorOption) {
	for _, fn := range options {
		fn(cfg)
	}
}

type ArithmeticCompressorOption func(*config)

func WithMetaCompression(cfg *config) {
	cfg.compressMeta = true
}

func WithPrecision(precision uint) ArithmeticCompressorOption {
	return func(cfg *config) {
		cfg.precision = precision
	}
}
