package compress

type config struct {
	// decoderChunkSize int
	compressMeta bool
}

func defaultConfig() *config {
	return &config{
		// decoderChunkSize: 100,
		compressMeta: false,
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
