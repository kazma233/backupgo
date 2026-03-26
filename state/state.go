package state

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

const (
	stateFile = "backupgo.state.json"
)

type TaskState struct {
	LastRun    time.Time `json:"last_run"`
	LastStatus string    `json:"last_status"`
}

type State struct {
	mu    sync.RWMutex
	tasks map[string]*TaskState
}

var (
	globalState *State
	once        sync.Once
)

func GetState() *State {
	once.Do(func() {
		globalState = &State{
			tasks: make(map[string]*TaskState),
		}
		globalState.load()
	})
	return globalState
}

func (s *State) load() {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return
	}

	var tasks map[string]*TaskState
	if err := json.Unmarshal(data, &tasks); err != nil {
		return
	}

	s.mu.Lock()
	s.tasks = tasks
	s.mu.Unlock()
}

func (s *State) save() {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.tasks, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return
	}

	os.WriteFile(stateFile, data, 0644)
}

func (s *State) GetTaskState(taskID string) *TaskState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if state, ok := s.tasks[taskID]; ok {
		return state
	}
	return nil
}

func (s *State) SetTaskRun(taskID string, status string) {
	s.mu.Lock()
	if s.tasks[taskID] == nil {
		s.tasks[taskID] = &TaskState{}
	}
	s.tasks[taskID].LastRun = time.Now()
	s.tasks[taskID].LastStatus = status
	s.mu.Unlock()

	s.save()
}
