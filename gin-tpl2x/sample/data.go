package sample

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

var Data = TodoPageData{
    PageTitle: "My TODO list",
    Todos: []Todo{
	{Title: "Task 1", Done: false},
	{Title: "Task 2", Done: true},
	{Title: "Task 3", Done: true},
    },
}
