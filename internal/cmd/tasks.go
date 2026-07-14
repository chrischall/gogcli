package cmd

type TasksCmd struct {
	Lists  TasksListsCmd  `cmd:"" name:"lists" help:"List task lists"`
	List   TasksListCmd   `cmd:"" name:"list" aliases:"ls" help:"List tasks" mcp:"tasks_list,read" mcpdesc:"List tasks in a task list."`
	Get    TasksGetCmd    `cmd:"" name:"get" aliases:"info,show" help:"Get a task" mcp:"tasks_get,read" mcpdesc:"Get a task by ID."`
	Add    TasksAddCmd    `cmd:"" name:"add" help:"Add a task" aliases:"create" mcp:"tasks_add,write" mcpdesc:"Add a task to a task list. Requires --allow-write."`
	Update TasksUpdateCmd `cmd:"" name:"update" aliases:"edit,set" help:"Update a task" mcp:"tasks_update,write" mcpdesc:"Update a task. Requires --allow-write."`
	Done   TasksDoneCmd   `cmd:"" name:"done" help:"Mark task completed" aliases:"complete" mcp:"tasks_complete,write" mcpdesc:"Mark a task completed. Requires --allow-write."`
	Undo   TasksUndoCmd   `cmd:"" name:"undo" help:"Mark task needs action" aliases:"uncomplete,undone"`
	Delete TasksDeleteCmd `cmd:"" name:"delete" aliases:"rm,del,remove" help:"Delete a task" mcp:"tasks_delete,write" mcpdesc:"Delete a task. Requires --allow-write."`
	Clear  TasksClearCmd  `cmd:"" name:"clear" help:"Clear completed tasks"`
	Raw    TasksRawCmd    `cmd:"" name:"raw" help:"Dump raw Google Tasks API response as JSON (Tasks.Get; lossless; for scripting and LLM consumption)"`
}
