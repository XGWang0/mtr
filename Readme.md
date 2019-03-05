
### MTR Tool
The tool is used to execute task in parrel.

You can input your tasks in file and trigger MTR run

#### Format of Task
Please follow a sample from cmdfile file in repo.

Or running
```c
go run mtr.go -help
```
to get help info.

if you have multiple tasks want to run, only add a new line for a task and follow help parameter n file


#### TODO
1. It is not perfect about  checking previous task[s] priority , and detect if current task need to wait them done or run directly
2. Code format need to be update
3. Some struct is not reasonable or convenient, need to optimize.
