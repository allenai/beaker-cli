package searchfield

type Dataset string

const (
	DatasetCommitted                   Dataset = "committed"
	DatasetCreatingUser                Dataset = "user"
	DatasetDescription                 Dataset = "description"
	DatasetID                          Dataset = "id"
	DatasetName                        Dataset = "name"
	DatasetNameOrDescription           Dataset = "nameOrDescription"
	DatasetNameOrDescriptionDeprecated Dataset = "name_or_description"
	DatasetOwner                       Dataset = "owner"
)
