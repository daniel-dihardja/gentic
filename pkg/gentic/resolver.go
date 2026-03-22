package gentic

type IntentResolver interface {
	Resolve(*State) Flow
}