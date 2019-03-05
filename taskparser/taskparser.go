package taskparser

//package cmd_parser

import (
	"commlib"
	"flag"
	//"fmt"
	"os"
	"path"
	"sort"
	"strings"
)

// Define basic case structure
type S_PARAM_STORE struct {
	Group    string
	Name     string
	Priority int
	Runobj   string
	Params   []string
	Logpath  string
	Timeout  int
}

var (
	S_P_S_MAP = make(map[string][]S_PARAM_STORE)
	//TASK_WORK_SPACE = path.Join("/tmp/mtr/log/", time.Now().Format("2006_01_02_15_04_08"))
	TASK_WORK_SPACE = "/tmp/mtr/log/"
	group           string
	name            string
	priority        int
	runobj          string
	timeout         int
	argsfile        string
)

func defineParams(f *flag.FlagSet) {
	f.StringVar(&name, "n", "task_sample", "Task need to be run")
	f.StringVar(&group, "g", "group_sample", "Task need to be run")
	f.IntVar(&priority, "p", 9, "Priority setting")
	f.StringVar(&runobj, "r", "ls", "Command line needed to be run")
	f.IntVar(&timeout, "t", 3, "Timeout for task run")
	f.StringVar(&argsfile, "f", "", "File which multiple lines parameters can write in")
}

// Warp parameter value
func (sps S_PARAM_STORE) handleParams(f *flag.FlagSet, parm_list []string) {

	var s_p_s_l []S_PARAM_STORE
	f.Parse(parm_list)

	sps.Group = group
	sps.Name = name
	sps.Priority = priority
	sps.Runobj = runobj
	sps.Params = f.Args()
	sps.Logpath = path.Join(TASK_WORK_SPACE, sps.Group, sps.Name)
	sps.Timeout = timeout

	if value, ok := S_P_S_MAP[sps.Group]; ok {
		S_P_S_MAP[sps.Group] = append(value, sps)
	} else {
		s_p_s_l = append(s_p_s_l, sps)
		S_P_S_MAP[sps.Group] = s_p_s_l
	}
	return

}

func ParserParams() {
	sps := S_PARAM_STORE{}
	var cmdflagset = flag.NewFlagSet("cmdflag", flag.ExitOnError)
	defineParams(cmdflagset)

	status, _ := commlib.In_array("-f", os.Args)

	if status {
		// Handel when using file as parameters
		cmdflagset.Parse(os.Args[1:])
		args_list, _ := commlib.ReadFile(argsfile)
		commlib.Mtrloggger.Printf("%+v\n", args_list)
		for _, value := range args_list {
			value_split := strings.Split(value, " ")
			sps.handleParams(cmdflagset, value_split)
		}
	} else {
		// Handel when using cmd as parameters
		sps.handleParams(cmdflagset, os.Args[1:])
	}

	//Sort task map according to each task's priority
	for _, value := range S_P_S_MAP {
		sort.SliceStable(value, func(i, j int) bool {
			return value[i].Priority < value[j].Priority
		})
	}
	commlib.Mtrloggger.Printf("[Debug]Full cases info : %#n", S_P_S_MAP)

}
