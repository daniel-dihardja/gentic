package gentic

type Step interface {
	Run(*State) error
}