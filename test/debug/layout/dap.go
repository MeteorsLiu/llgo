package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// dapClient uses LLDB's native debug adapter over stdio. Neither this runner
// nor the debuggee loads a Python script or an LLGo presentation plugin.
type dapClient struct {
	command   *exec.Cmd
	input     io.WriteCloser
	output    *bufio.Reader
	log       *os.File
	sequence  int
	responses map[int]dapMessage
	events    []dapMessage
}

type dapMessage struct {
	Seq        int             `json:"seq"`
	Type       string          `json:"type"`
	RequestSeq int             `json:"request_seq"`
	Success    bool            `json:"success"`
	Command    string          `json:"command"`
	Event      string          `json:"event"`
	Message    string          `json:"message"`
	Body       json.RawMessage `json:"body"`
}

func startDAP(path, logPath string) (*dapClient, error) {
	log, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}
	command := exec.Command(path)
	if runtime.GOOS == "windows" {
		// CI already selects a target-native LLDB and its matching runtime.
		// Apply those dependencies only to the adapter child, as llgo lldb does.
		command.Env = os.Environ()
		searchPath := filepath.Dir(os.Getenv("LLGO_LLDB")) + ";" + os.Getenv("PATH")
		if home := os.Getenv("LLGO_WINDOWS_LLDB_PYTHON_HOME"); home != "" {
			searchPath = home + ";" + searchPath
			command.Env = append(command.Env, "PYTHONHOME="+home, "PYTHONUTF8=1", "PYTHONIOENCODING=utf-8")
			if pythonPath := os.Getenv("LLGO_WINDOWS_LLDB_PYTHONPATH"); pythonPath != "" {
				command.Env = append(command.Env, "PYTHONPATH="+pythonPath)
			}
		}
		command.Env = append(command.Env, "PATH="+searchPath)
	}
	command.Stderr = log
	input, err := command.StdinPipe()
	if err != nil {
		log.Close()
		return nil, err
	}
	output, err := command.StdoutPipe()
	if err != nil {
		input.Close()
		log.Close()
		return nil, err
	}
	if err := command.Start(); err != nil {
		input.Close()
		log.Close()
		return nil, err
	}
	return &dapClient{command: command, input: input, output: bufio.NewReader(output), log: log, responses: make(map[int]dapMessage)}, nil
}

func (c *dapClient) send(command string, args any) (int, error) {
	c.sequence++
	message := struct {
		Seq       int    `json:"seq"`
		Type      string `json:"type"`
		Command   string `json:"command"`
		Arguments any    `json:"arguments"`
	}{c.sequence, "request", command, args}
	data, err := json.Marshal(message)
	if err != nil {
		return 0, err
	}
	fmt.Fprintf(c.log, "SEND %s\n", data)
	_, err = fmt.Fprintf(c.input, "Content-Length: %d\r\n\r\n%s", len(data), data)
	return c.sequence, err
}

func (c *dapClient) receive() error {
	length := -1
	for {
		line, err := c.output.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if value, ok := strings.CutPrefix(line, "Content-Length:"); ok {
			length, err = strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return err
			}
		}
	}
	if length < 0 {
		return fmt.Errorf("LLDB DAP message has no Content-Length")
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(c.output, data); err != nil {
		return err
	}
	fmt.Fprintf(c.log, "RECV %s\n", data)
	var message dapMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}
	switch message.Type {
	case "response":
		c.responses[message.RequestSeq] = message
	case "event":
		c.events = append(c.events, message)
	default:
		return fmt.Errorf("unsupported LLDB DAP message: %s", data)
	}
	return nil
}

func (c *dapClient) response(sequence int, result any) error {
	for {
		if message, ok := c.responses[sequence]; ok {
			delete(c.responses, sequence)
			if !message.Success {
				return fmt.Errorf("LLDB %s: %s (%s)", message.Command, message.Message, message.Body)
			}
			if result != nil {
				return json.Unmarshal(message.Body, result)
			}
			return nil
		}
		if err := c.receive(); err != nil {
			return err
		}
	}
}

func (c *dapClient) request(command string, args, result any) error {
	sequence, err := c.send(command, args)
	if err != nil {
		return err
	}
	return c.response(sequence, result)
}

func (c *dapClient) event(name string, result any) error {
	for {
		for i, event := range c.events {
			if event.Event == name {
				c.events = append(c.events[:i], c.events[i+1:]...)
				if result != nil {
					return json.Unmarshal(event.Body, result)
				}
				return nil
			}
			if event.Event == "terminated" || event.Event == "exited" {
				return fmt.Errorf("process %s before %s: %s", event.Event, name, event.Body)
			}
		}
		for _, message := range c.responses {
			if !message.Success {
				return fmt.Errorf("LLDB %s: %s", message.Command, message.Message)
			}
		}
		if err := c.receive(); err != nil {
			return err
		}
	}
}

func (c *dapClient) readMemory(address uint64, size int) ([]byte, error) {
	var result struct {
		Data            string `json:"data"`
		UnreadableBytes int    `json:"unreadableBytes"`
	}
	if err := c.request("readMemory", map[string]any{"memoryReference": fmt.Sprintf("0x%x", address), "count": size}, &result); err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(result.Data)
	if err != nil {
		return nil, err
	}
	if result.UnreadableBytes != 0 || len(data) != size {
		return nil, fmt.Errorf("read %#x: got %d bytes, want %d; unreadable=%d", address, len(data), size, result.UnreadableBytes)
	}
	return data, nil
}

func (c *dapClient) close() error {
	// Do not terminate a debuggee on a diagnostic failure. Disconnect detaches;
	// a normal run has already exited after its last checkpoint.
	err := c.request("disconnect", map[string]any{"terminateDebuggee": false}, nil)
	c.input.Close()
	waitErr := c.command.Wait()
	c.log.Close()
	if err != nil {
		return err
	}
	return waitErr
}
