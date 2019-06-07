package searchfield

type Dataset string

const (
	DatasetCreated           Dataset = "created"
	DatasetCommitted         Dataset = "committed"
	DatasetCreatingUser      Dataset = "user"
	DatasetDescription       Dataset = "description"
	DatasetID                Dataset = "id"
	DatasetName              Dataset = "name"
	DatasetNameOrDescription Dataset = "nameOrDescription"
	DatasetOwner             Dataset = "owner"
)
