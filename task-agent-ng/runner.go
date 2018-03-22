package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/raintank/worldping-api/pkg/log"
)

// QueueRun does stuff
func QueueRun(taskname string) (int, error) {
	return 0, nil
}

// RunTask runs stuff
func RunTask(taskName string, taskArgs string, timeout time.Duration) (int, error) {
	log.Debug("Running task: %s", taskName)

	// docker build current directory
	cmdName := taskName
	cmdArgs := []string{taskArgs}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Error(3, "Error creating StdoutPipe for Cmd %", err)
		return 1, err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("command out | %s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Error(3, "Error starting Cmd: %s", err)
		return 2, err
	}

	// kill if running too long
	timer := time.AfterFunc(timeout*time.Second, func() {
		cmd.Process.Kill()
		log.Error(3, "Timedout waiting for process to finish: %s", err)
	})

	err = cmd.Wait()
	timer.Stop()
	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	exitStatus := status.ExitStatus()
	signaled := status.Signaled()
	signal := status.Signal()
	if signaled {
		if signal.String() == "killed" {
			return 3, err
		}
	} else {
		fmt.Println("Status:", exitStatus)
	}
	return 0, err
}
