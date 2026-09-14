package validator

var (
	//	Built-in validation rules message's
	Required = "%s field is required."
	Max      = "%s field can be at most %s characters long."
	Min      = "%s field must be at least %s characters long."

	//	Custom validation messages
	Username = `%s field must contain at least one letter and can contain only letters, numbers, and "._" characters.`
	Path     = `%s field must contain valid folder names separated by single "/" characters.`

	//	Default validation message
	Default = "%s field failed on %s validation rule."
)
