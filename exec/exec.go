package exec

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/oktalz/present/types"
)

func Cmd(tc types.TerminalCommand) []byte {
	cmd := exec.Command(tc.App, tc.Cmd...)
	cmd.Dir = tc.Dir
	output, err := cmd.Output()
	if err != nil {
		return []byte(err.Error())
	}
	return output
}

func DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func CmdStream(tc types.TerminalCommand) (lines []types.TerminalOutputLine) {
	fmt.Println("======== executing", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, tc.App, tc.Cmd...)
	if DirectoryExists(tc.Dir) {
		cmd.Dir = tc.Dir
	} else {
		dir, err := os.Getwd()
		if err != nil {
			return append(lines, types.TerminalOutputLine{
				Timestamp: "0.1",
				Line:      err.Error(),
			})
		}
		cmd.Dir = path.Join(dir, tc.Dir)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return append(lines, types.TerminalOutputLine{
			Timestamp: "0.1",
			Line:      err.Error(),
		})
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return append(lines, types.TerminalOutputLine{
			Timestamp: "0.1",
			Line:      err.Error(),
		})
	}

	timeStarted := time.Now()
	if err := cmd.Start(); err != nil {
		return append(lines, types.TerminalOutputLine{
			Timestamp: "0.1",
			Line:      err.Error(),
		})
	}

	var mu sync.Mutex

	go func() {
		scannerOut := bufio.NewScanner(stdout)
		for scannerOut.Scan() {
			now := time.Now()
			diff := now.Sub(timeStarted)
			ts := diff.Microseconds()
			seconds := ts / 1000000
			micros := ts % 1000000
			line := types.TerminalOutputLine{
				Timestamp: fmt.Sprintf("%d.%d", seconds, micros),
				Line:      scannerOut.Text(),
			}
			mu.Lock()
			lines = append(lines, line)
			mu.Unlock()
			fmt.Println(scannerOut.Text())
		}
	}()
	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			now := time.Now()
			diff := now.Sub(timeStarted)
			ts := diff.Microseconds()
			seconds := ts / 1000000
			micros := ts % 1000000
			line := types.TerminalOutputLine{
				Timestamp: fmt.Sprintf("%d.%d", seconds, micros),
				Line:      scannerErr.Text(),
			}
			mu.Lock()
			lines = append(lines, line)
			mu.Unlock()
			fmt.Println(scannerErr.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		now := time.Now()
		diff := now.Sub(timeStarted)
		ts := diff.Microseconds()
		seconds := ts / 1000000
		micros := ts % 1000000
		fmt.Println(err.Error())
		fmt.Println("======== finished ", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))
		return append(lines, types.TerminalOutputLine{
			Timestamp: fmt.Sprintf("%d.%d", seconds, micros),
			Line:      err.Error(),
		})
	}
	fmt.Println("======== finished ", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))

	return lines
}
func CmdStreamWS(tc types.TerminalCommand, ch chan string) {
	fmt.Println("======== executing", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, tc.App, tc.Cmd...)
	if DirectoryExists(tc.Dir) {
		cmd.Dir = tc.Dir
	} else {
		dir, err := os.Getwd()
		if err != nil {
			ch <- err.Error()
			close(ch)
			return
		}
		cmd.Dir = path.Join(dir, tc.Dir)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		ch <- err.Error()
		close(ch)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		ch <- err.Error()
		close(ch)
		return
	}

	if err := cmd.Start(); err != nil {
		ch <- err.Error()
		close(ch)
		return
	}

	go func() {
		scannerOut := bufio.NewScanner(stdout)
		for scannerOut.Scan() {
			ch <- scannerOut.Text()
			//fmt.Println(scannerOut.Text())
			log.Println(scannerOut.Text())
		}
	}()
	go func() {
		scannerErr := bufio.NewScanner(stderr)
		for scannerErr.Scan() {
			ch <- scannerErr.Text()
			fmt.Println(scannerErr.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		ch <- err.Error()
		fmt.Println(err.Error())
		fmt.Println("======== finished ", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))
		close(ch)
		return
	}
	fmt.Println("======== finished ", tc.Dir, tc.App, strings.Join(tc.Cmd, " "))

	close(ch)
	return
}
