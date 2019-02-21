package searchfield

type Experiment string

const (
	ExperimentID                Experiment = "id"
	ExperimentName              Experiment = "name"
	ExperimentDescription       Experiment = "description"
	ExperimentNameOrDescription Experiment = "name_or_description"
	ExperimentCreated           Experiment = "created"
	ExperimentCreatingUser      Experiment = "user"
	ExperimentStatus            Experiment = "status"
	ExperimentOwner             Experiment = "owner"
)
