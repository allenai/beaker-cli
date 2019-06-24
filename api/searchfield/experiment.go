package searchfield

type Experiment string

const (
	ExperimentCreated                     Experiment = "created"
	ExperimentCreatingUser                Experiment = "user"
	ExperimentDescription                 Experiment = "description"
	ExperimentID                          Experiment = "id"
	ExperimentName                        Experiment = "name"
	ExperimentNameOrDescription           Experiment = "nameOrDescription"
	ExperimentNameOrDescriptionDeprecated Experiment = "name_or_description"
	ExperimentOwner                       Experiment = "owner"
	ExperimentStatus                      Experiment = "status"
)
