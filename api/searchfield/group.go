package searchfield

type Group string

const (
	GroupID                Group = "id"
	GroupName              Group = "name"
	GroupDescription       Group = "description"
	GroupNameOrDescription Group = "name_or_description"
	GroupCreatingUser      Group = "user"
	GroupCreated           Group = "created"
	GroupModified          Group = "modified"
)
