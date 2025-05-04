package tokenizer

import (
	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
)

type Service struct {
	tk *tokenizer.Tokenizer
}

func (s *Service) Count(input string) (int, error) {
	enc, err := s.tk.EncodeSingle(input)
	if err != nil {
		return 0, err
	}

	return len(enc.Tokens), nil
}

type Option func(*Service)

func WithPretrainedFromCache(model string, filename string) Option {
	return func(s *Service) {
		configFile, err := tokenizer.CachedPath("bert-base-uncased", "tokenizer.json")
		if err != nil {
			panic(err)
		}

		tk, err := pretrained.FromFile(configFile)
		if err != nil {
			panic(err)
		}

		s.tk = tk
	}
}

func New(o ...Option) *Service {
	ret := &Service{}

	for _, opt := range o {
		opt(ret)
	}

	return ret
}
