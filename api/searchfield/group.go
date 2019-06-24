package searchfield

type Group string

const (
	GroupCreated                     Group = "created"
	GroupCreatingUser                Group = "user"
	GroupDescription                 Group = "description"
	GroupID                          Group = "id"
	GroupModified                    Group = "modified"
	GroupName                        Group = "name"
	GroupNameOrDescription           Group = "nameOrDescription"
	GroupNameOrDescriptionDeprecated Group = "name_or_description"
	GroupOwner                       Group = "owner"
)
