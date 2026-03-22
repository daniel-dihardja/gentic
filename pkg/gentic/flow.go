package gentic

type Flow struct {
	steps []Step
}

func NewFlow(steps ...Step) Flow {
	return Flow{steps: steps}
}

func (f Flow) Run(s *State) error {
	for _, step := range f.steps {
		if err := step.Run(s); err != nil {
			return err
		}
	}
	return nil
}