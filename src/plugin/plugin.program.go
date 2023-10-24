package zgwit_plugin

import (
	"fmt"
	zgwit_utils "local/utils"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Program struct {
	Switch bool
	Isrun  bool

	client *exec.Cmd
	pid    int
	ProgramStatus
	ProgramExit
}

type ProgramStatus struct {
	statename string

	start_time  int64
	stop_time   int64
	start_count int64
}

type ProgramExit struct {
	exited     bool
	exited_log bool
	exited_pro bool
}

func (program *Program) Runing(work_dir string, command string, log_file_path string) {

	program.exited = false
	program.exited_log = false
	program.exited_pro = false

	go program.loging(log_file_path)
	go program.monitoring(work_dir, command, log_file_path)
}

func (program *Program) Exiting(timeout int) (result bool) {

	program.exited = true

	timer := 0

	for (!program.exited_log || !program.exited_pro) && timer < timeout {
		timer++
		time.Sleep(time.Second)
	}

	if timer >= timeout {
		program.exited = false
		return
	}

	return true
}

func (program *Program) GetStatename() string {

	if program.statename == "" {
		return "-"
	}

	return program.statename
}

func (program *Program) GetRestartCount() int64 {
	return program.start_count
}

func (program *Program) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"statename":  program.statename,
		"start_time": program.start_time,
		"stop_time":  program.stop_time,
	}
}

func (program *Program) GetLog(log_file_path string, row int) (contents []string, err error) {

	return zgwit_utils.ReadFileTail(log_file_path, row)
}

func (program *Program) loging(log_file_path string) {

	timer, number := 0, 0

	for !program.exited_log {

		if program.exited_log = program.exited; program.exited_log {
			goto WAIT
		}

		if timer >= 60 {
			timer = 0

		} else {
			timer = timer + 1
			goto WAIT
		}

		if fileInfo, err := os.Stat(log_file_path); err != nil || fileInfo.Size() < 10*1024*1024 {
			goto WAIT
		}

		if names, err := zgwit_utils.GetDirsNames("./log/"); err != nil {
			log.Printf("program.%s.getnumaber.error: %s\n", log_file_path, err.Error())
			goto WAIT
		} else {
			number = len(names)
		}

		if err := zgwit_utils.CopyFile(log_file_path, fmt.Sprintf("%s.%d", log_file_path, number)); err != nil {
			log.Printf("program.%s.log.copy.error: %s\n", log_file_path, err.Error())
			goto WAIT
		}

		if err := os.Truncate(log_file_path, 0); err != nil {
			log.Printf("program.%s.log.clear.error: %s\n", log_file_path, err.Error())
		}

	WAIT:
		time.Sleep(time.Second)
	}
}

func (program *Program) monitoring(work_dir string, command string, log_file_path string) {

	program.Switch = true

	for !program.exited_pro {

		switch {

		case program.exited:

			program.client.Process.Kill()

		case !program.Isrun && !program.Switch:

			program.statename = "stop"

			goto WAIT

		case !program.Isrun && program.Switch:

			program.statename = "starting"

			go program.mounting(work_dir, command, log_file_path)

		case program.Isrun && program.Switch:

			program.statename = "running"

		case program.Isrun && !program.Switch:

			program.statename = "stoping"

			program.client.Process.Kill()
		}

	WAIT:
		time.Sleep(time.Second)
	}
}

func (program *Program) mounting(work_dir string, command string, log_file_path string) {

	commands := strings.Split(strings.TrimSpace(command), " ")

	if len(commands) == 0 {
		commands = append(commands, "null")
	}

	program.client = exec.Command(commands[0], commands[1:]...)

	program.client.Dir = work_dir
	program.client.Stderr = os.NewFile(0, os.DevNull)

	if outputFile, err := os.OpenFile(log_file_path, os.O_RDWR|os.O_CREATE, 0755); err != nil {
		program.client.Stdout = os.NewFile(0, os.DevNull)

	} else {
		program.client.Stdout = outputFile
	}

	if program.start_time == 0 {
		program.start_time = time.Now().Unix()
	}

	if err := program.client.Start(); err != nil {
		program.statename = "fatal"
		goto TIMER
	}

	program.Isrun = true
	program.pid = program.client.Process.Pid

	program.client.Wait()

	program.Isrun = false
	program.stop_time = time.Now().Unix()
	program.exited_pro = program.exited
	program.statename = "exited"

TIMER:
	program.start_count++
}
