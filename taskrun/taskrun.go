package taskrun

import (
	"bytes"
	"commlib"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"taskparser"
	"time"
)

// Implement interface for task
type TaskExecer interface {
	InitTask()
	TaskRunning() error
	getExitcode(error)
}

// Define running time structure of task
type TaskRun struct {
	E_status        string        //Task status
	S_time          time.Time     //Start time
	E_time          time.Time     //End time
	R_status        int           //Running status, not used
	Duration        time.Duration //Task running duration
	Output          string        //Task output
	ExitCode        int
	ExitErr         interface{}
	Stdoutfile_path string //Output to file
	Stderrfile_path string
}

// Recreate task structure containing basic info, running info, cmd structure etc info
type Task struct {
	task    taskparser.S_PARAM_STORE //basic info from command parsing
	taskRun TaskRun                  //running time
	cmd     *exec.Cmd
}

var (
	Task_map = make(map[string]map[string]*Task, 0) //Whole task map
	top_wp   sync.WaitGroup
	mid_wp   sync.WaitGroup
)

// Run cmd with timeout checking
func CmdRunWithTimeout(cmd *exec.Cmd, timeout time.Duration) (error, bool) {
	done := make(chan error)
	// Waiting Task end
	go func() {
		done <- cmd.Wait()

	}()

	commlib.Mtrloggger.Println("[INFO] Execute Command with non-blocking mode. Time is:", timeout)

	var err error
	// Trigger timeout handling
	select {
	case <-time.After(timeout):
		if err = cmd.Process.Kill(); err != nil {
			commlib.Mtrloggger.Println("[ERROR]Failed to kill process when time takes out: %s, error: %s",
				cmd.Path, err.Error())
		}
		go func() {
			<-done // allow goroutine to exit although forcing kill it
		}()
		return err, true
	case err = <-done:
		commlib.Mtrloggger.Println("[INFO] Command is executed in time")
		if err == nil {
			return err, true
		} else {
			return err, false
		}
	}

}

// task structure method, get cmd exit code
func (task *Task) getExitcode(err error) {
	var exitcode int

	if err != nil {
		task.taskRun.ExitErr = err.Error()
		if exiterr, ok := err.(*exec.ExitError); ok {
			es := exiterr.Sys().(syscall.WaitStatus)
			exitcode = es.ExitStatus()
		} else {
			exitcode = -1
		}

	} else {

		if task.cmd.ProcessState != nil {
			ws := task.cmd.ProcessState.Sys().(syscall.WaitStatus)
			exitcode = ws.ExitStatus()
		} else {
			exitcode = -1
		}

	}
	//commlib.Mtrloggger.Println("[getExitc] Exit code is:", exitcode)
	task.taskRun.ExitCode = exitcode
}

// function, create multiple writer for stderr and stdout
func GetMultpleStdWriter(file_path string) (*os.File, *bytes.Buffer) {
	var (
		byte_buf bytes.Buffer
	)
	if stdout_filep, cf_err := commlib.CreateFile(file_path); cf_err == nil {
		return stdout_filep, &byte_buf
	} else {
		return nil, &byte_buf
	}
}

// task structure 's method, to initialize task info before running
func (task *Task) InitTask() {
	commlib.Mtrloggger.Println("Finished InitTask method")

	task.taskRun.Stdoutfile_path = task.task.Logpath + ".stdout"
	task.taskRun.Stderrfile_path = task.task.Logpath + ".stderr"
	task.taskRun.E_status = "Running"
	task.taskRun.S_time = time.Now()
	commlib.Mtrloggger.Printf("[INFO]Task Structure After Init %+v\n", task)

}

func (task *Task) DestroyTask() {
	task.taskRun.E_time = time.Now()
	task.taskRun.Duration = task.taskRun.E_time.Sub(task.taskRun.S_time)
	task.taskRun.E_status = "Done"
}

// task structure's method, executing cmd flow
func (task *Task) TaskRunning() error {
	task.InitTask()
	task.cmd = exec.Command(task.task.Runobj, task.task.Params...)

	// Assign file descriptor and bytes buffer to stdout of cmd
	stdout_filep, stdout_bytes := GetMultpleStdWriter(task.taskRun.Stdoutfile_path)
	defer stdout_filep.Close()
	task.cmd.Stdout = io.MultiWriter(stdout_filep, stdout_bytes)

	// Assign file descriptor and bytes buffer to stderr of cmd
	stderr_filep, stderr_bytes := GetMultpleStdWriter(task.taskRun.Stdoutfile_path)
	defer stderr_filep.Close()
	task.cmd.Stdout = io.MultiWriter(stderr_filep, stderr_bytes)

	// Trigger task run
	if err := task.cmd.Start(); err != nil {
		commlib.Mtrloggger.Println("[ERROR]Failed to start cmd due to ", err.Error())
		os.Exit(-1)
	}
	err, _ := CmdRunWithTimeout(task.cmd, time.Duration(task.task.Timeout)*time.Second)

	task.getExitcode(err)
	task.DestroyTask()

	task.taskRun.Output = stdout_bytes.String() + stderr_bytes.String()

	commlib.Mtrloggger.Printf("[DEBUG] Current Running Task structure : %#v\n", task)
	// Collect task info
	if value, ok := Task_map[task.task.Group]; ok {
		value[task.task.Name] = task
	} else {
		tmptaskmap := map[string]*Task{task.task.Name: task}
		Task_map[task.task.Group] = tmptaskmap
	}

	return err
}

// multiple task running in parallel
func RunTaskMultiple(tasksMap map[string][]taskparser.S_PARAM_STORE) {

	commlib.CreateFolder(taskparser.TASK_WORK_SPACE)
	// traverse task map
	for _, tasks := range tasksMap {
		// Add top level lock
		top_wp.Add(1)
		go func(ts []taskparser.S_PARAM_STORE) {
			for index, task := range ts {
				taskp := &Task{task: task}
				/*
					go func(task_p *Task) {
						defer mid_wp.Done()
						commlib.Mtrloggger.Printf("%s,%+v\n", "------------------", task_p)
						task_p.TaskRunning()
					}(taskp)
				*/
				/* Check previous task's priority and make sure if current task
				needs to wait them
				*/
				if index > 0 && taskp.task.Priority == ts[index-1].Priority {
					mid_wp.Add(1)
					go func(task_p *Task) {
						defer mid_wp.Done()
						//commlib.Mtrloggger.Printf("%s,%+v\n", "------------------", task_p)
						task_p.TaskRunning()
					}(taskp)
				} else if index != 0 {
					// Long wait until previous task finished
					for {
						if _, ok := Task_map[ts[index-1].Group][ts[index-1].Name]; ok {
							preTaskStatus := Task_map[ts[index-1].Group][ts[index-1].Name].taskRun.E_status
							if preTaskStatus == "Done" {
								break
							}
						}
						time.Sleep(3 * time.Second)

					}
					taskp.TaskRunning()
				} else {
					taskp.TaskRunning()
				}

			}
			mid_wp.Wait()
			top_wp.Done()
		}(tasks)
	}
	top_wp.Wait()
	commlib.Mtrloggger.Println("[INFO] Finished all tasks")
	//commlib.Mtrloggger.Printf("[INFO] task map list info %v\n", Task_map)
}
