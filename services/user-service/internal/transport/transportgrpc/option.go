package transportgrpc

import (
	"github.com/go-playground/validator/v10"
)

// Option allows functional options to be passed into server
type Option func(s *Server)

// WithValidator allows the caller to override the default validator
func WithValidator(validate *validator.Validate) Option {
	return func(s *Server) {
		s.validate = validate
	}
}
