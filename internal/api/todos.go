package api

// Todo represents a todo item.
type Todo struct {
	TodoID              int         `json:"todoId"`
	UserID              int         `json:"userId"`
	Login               string      `json:"login,omitempty"`
	ProjectShort        string      `json:"projectShort,omitempty"`
	ItemRef             string      `json:"itemRef,omitempty"`
	FieldLabel          string      `json:"fieldLabel,omitempty"`
	FieldID             int         `json:"fieldId,omitempty"`
	Auto                bool        `json:"auto,omitempty"`
	OriginatorUserID    int         `json:"originatorUserId,omitempty"`
	OriginatorLogin     string      `json:"originatorLogin,omitempty"`
	Action              *TodoAction `json:"action,omitempty"`
	CreatedAt           string      `json:"createdAt,omitempty"`
	ClosedAt            string      `json:"closedAt,omitempty"`
	CreatedAtUserFormat string      `json:"createdAtUserFormat,omitempty"`
	ClosedAtUserFormat  string      `json:"closedAtUserFormat,omitempty"`
	Future              bool        `json:"future,omitempty"`
	ProjectID           int         `json:"projectId,omitempty"`
}

// TodoAction describes the action associated with a todo.
type TodoAction struct {
	Text     string `json:"text,omitempty"`
	TodoType string `json:"todoType,omitempty"`
}

// TodoCount represents a count of todos per user/project.
type TodoCount struct {
	UserID       int    `json:"userId"`
	Login        string `json:"login,omitempty"`
	ProjectID    int    `json:"projectId,omitempty"`
	ProjectShort string `json:"projectShort,omitempty"`
	NbTodos      int    `json:"nbTodos"`
	FirstTodos   []Todo `json:"firstTodos,omitempty"`
}
