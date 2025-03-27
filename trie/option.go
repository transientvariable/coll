package trie

// Option is a container for optional properties that can be used to initialize a Trie.
type Option struct {
	digitizer Digitizer
}

// WithDigitizer sets the Digitizer Option for the Trie.
func WithDigitizer(digitizer Digitizer) func(*Option) {
	return func(options *Option) {
		options.digitizer = digitizer
	}
}
