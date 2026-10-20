package domain

// StageKind says whether a stage holds open deals or closes them.
type StageKind string

const (
	StageOpen StageKind = "open"
	StageWon  StageKind = "won"
	StageLost StageKind = "lost"
)

func (k StageKind) IsValid() bool {
	return k == StageOpen || k == StageWon || k == StageLost
}

// Stage is a step of a pipeline. Probability (0-100) is the chance a deal in
// it is won, which weights pipeline value.
type Stage struct {
	ID          int64
	PipelineID  int64
	Name        string
	Position    int
	Probability int
	Kind        StageKind
}

// Pipeline is one sales motion's ordered stages.
type Pipeline struct {
	ID        int64
	Name      string
	Position  int
	IsDefault bool
	Stages    []Stage
}

// StageOfKind returns the pipeline's first stage of a kind.
func (p Pipeline) StageOfKind(kind StageKind) (Stage, bool) {
	for _, stage := range p.Stages {
		if stage.Kind == kind {
			return stage, true
		}
	}
	return Stage{}, false
}

// Stage returns the pipeline's stage with an id.
func (p Pipeline) Stage(id int64) (Stage, bool) {
	for _, stage := range p.Stages {
		if stage.ID == id {
			return stage, true
		}
	}
	return Stage{}, false
}

// DefaultPipelineName and DefaultStages seed a business's first pipeline. The
// sales redesign migration seeds the same stages for converted legacy leads.
const DefaultPipelineName = "Sales"

func DefaultStages() []StageInput {
	return []StageInput{
		{Name: "Qualified", Probability: 10, Kind: StageOpen},
		{Name: "Meeting", Probability: 25, Kind: StageOpen},
		{Name: "Proposal", Probability: 50, Kind: StageOpen},
		{Name: "Negotiation", Probability: 75, Kind: StageOpen},
		{Name: "Won", Probability: 100, Kind: StageWon},
		{Name: "Lost", Probability: 0, Kind: StageLost},
	}
}

// StageInput is a stage as edited. ID is nil for a new stage; a stage left out
// of an edit is removed.
type StageInput struct {
	ID          *int64
	Name        string
	Probability int
	Kind        StageKind
}

// PipelineInput creates or replaces a pipeline and its stages, in order.
type PipelineInput struct {
	Name   string
	Stages []StageInput
}

// Normalize checks the pipeline closes deals both ways: at least one open
// stage, and exactly one won and one lost stage.
func (in *PipelineInput) Normalize() error {
	in.Name = trimTo(in.Name, 100)
	if in.Name == "" {
		return invalid("a pipeline needs a name")
	}
	if len(in.Stages) > 20 {
		return invalid("a pipeline can have at most 20 stages")
	}
	counts := map[StageKind]int{}
	for i := range in.Stages {
		stage := &in.Stages[i]
		stage.Name = trimTo(stage.Name, 100)
		if stage.Name == "" {
			return invalid("every stage needs a name")
		}
		if !stage.Kind.IsValid() {
			return invalid("unknown stage kind %q", stage.Kind)
		}
		switch stage.Kind {
		case StageWon:
			stage.Probability = 100
		case StageLost:
			stage.Probability = 0
		}
		if stage.Probability < 0 || stage.Probability > 100 {
			return invalid("stage probability must be between 0 and 100")
		}
		counts[stage.Kind]++
	}
	if counts[StageOpen] == 0 || counts[StageWon] != 1 || counts[StageLost] != 1 {
		return invalid("a pipeline needs open stages and exactly one won and one lost stage")
	}
	return nil
}
