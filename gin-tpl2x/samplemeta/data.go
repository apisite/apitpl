package samplemeta

// Todo holds single todo item attrs
type Todo struct {
	Title string
	Done  bool
}

// TodoPageData holds todo page attrs
type TodoPageData struct {
	PageTitle string
	Todos     []Todo
}

// Data holds a sample of some static external data
var Data = TodoPageData{
	PageTitle: "My TODO list",
	Todos: []Todo{
		{Title: "Task 1", Done: false},
		{Title: "Task 2", Done: true},
		{Title: "Task 3", Done: true},
	},
}
