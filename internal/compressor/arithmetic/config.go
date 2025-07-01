package arithmetic

type config struct {
	precision uint
	chunkSize int
}

func defaultConfig() *config {
	return &config{
		precision: 166,
		chunkSize: 50,
	}
}

func (cfg *config) apply(options []ArithmeticCompressorOption) {
	for _, fn := range options {
		fn(cfg)
	}
}

type ArithmeticCompressorOption func(*config)

func WithPrecision(precision uint) ArithmeticCompressorOption {
	return func(cfg *config) {
		cfg.precision = precision
	}
}

func WithChunkSize(chunkSize uint) ArithmeticCompressorOption {
	return func(cfg *config) {
		cfg.chunkSize = int(chunkSize)
	}
}
