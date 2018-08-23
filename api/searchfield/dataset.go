package searchfield

type Dataset string

const (
	DatasetID                Dataset = "id"
	DatasetName              Dataset = "name"
	DatasetDescription       Dataset = "description"
	DatasetNameOrDescription Dataset = "name_or_description"
	DatasetCommitted         Dataset = "committed"
	DatasetCreatingUser      Dataset = "user"
)
