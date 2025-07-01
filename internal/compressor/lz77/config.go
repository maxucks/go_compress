package lz77

type config struct {
	windowSize int
}

func defaultConfig() *config {
	return &config{
		windowSize: 512,
	}
}

func (cfg *config) apply(options []LZ77CompressorOption) {
	for _, fn := range options {
		fn(cfg)
	}
}

type LZ77CompressorOption func(*config)

func SetBuffer(size int) LZ77CompressorOption {
	return func(cfg *config) {
		cfg.windowSize = size
	}
}
