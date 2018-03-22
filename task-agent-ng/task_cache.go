package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/raintank/raintank-apps/task-server/model"
	"github.com/raintank/worldping-api/pkg/log"
)

type ScheduledTask struct {
	ID                 int64
	Name               string
	CreationTimestamp  int64
	LastRunTimestamp   int64
	FailedCount        int64
	LastFailureMessage string
	State              string
}

type TaskCache struct {
	sync.RWMutex
	tsdbURL     *string
	Tasks       map[int64]*model.TaskDTO
	ActiveTasks map[string]*ScheduledTask
	initialized bool
}

func (t *TaskCache) AddTask(task *model.TaskDTO) error {
	t.Lock()
	defer t.Unlock()
	return t.addTask(task)
}

func (t *TaskCache) addTask(task *model.TaskDTO) error {
	t.Tasks[task.Id] = task
	if !t.initialized {
		return nil
	}
	taskName := fmt.Sprintf("raintank-apps:%d", task.Id)
	snapTask, ok := t.ActiveTasks[taskName]
	if !ok {
		log.Debug("New task recieved %s", taskName)
		// just append it
		aTask := ScheduledTask{0, taskName, 0, 0, 0, "", ""}
		t.ActiveTasks[taskName] = &aTask
	} else {
		log.Debug("task %s already in the cache.", taskName)
		if task.Updated.After(time.Unix(snapTask.CreationTimestamp, 0)) {
			log.Debug("%s needs to be updated", taskName)
			// need to update task
			_, ok := t.ActiveTasks[taskName]
			if ok {
				delete(t.ActiveTasks, taskName)
			}
			aTask := ScheduledTask{0, taskName, 0, 0, 0, "", ""}
			t.ActiveTasks[taskName] = &aTask
		}
	}

	return nil
}

func (t *TaskCache) UpdateTasks(tasks []*model.TaskDTO) {
	seenTaskIds := make(map[int64]struct{})
	t.Lock()
	for _, task := range tasks {
		seenTaskIds[task.Id] = struct{}{}
		err := t.addTask(task)
		if err != nil {
			log.Error(3, err.Error())
		}
	}
	tasksToDel := make([]*model.TaskDTO, 0)
	for id, task := range t.Tasks {
		if _, ok := seenTaskIds[id]; !ok {
			tasksToDel = append(tasksToDel, task)
		}
	}
	t.Unlock()
	if len(tasksToDel) > 0 {
		for _, task := range tasksToDel {
			if err := t.RemoveTask(task); err != nil {
				log.Error(3, "Failed to remove task %d", task.Id)
			}
		}
	}
}

func (t *TaskCache) Sync() {
	tasksByName := make(map[string]*model.TaskDTO)
	t.Lock()
	for _, task := range t.Tasks {
		name := fmt.Sprintf("raintank-apps:%d", task.Id)
		tasksByName[name] = task
		log.Debug("seen %s", name)
		err := t.addTask(task)
		if err != nil {
			log.Error(3, err.Error())
		}
	}

	for name := range t.ActiveTasks {
		// dont remove tasks that were not added by us.
		if !strings.HasPrefix(name, "raintank-apps") {
			continue
		}
		if _, ok := tasksByName[name]; !ok {
			log.Info("%s not in taskList. removing from snap.", name)
			if err := t.removeActiveTask(name); err != nil {
				log.Error(3, "failed to remove snapTask %s. %s", name, err)
			}
		}
	}
	t.Unlock()

}

func (t *TaskCache) RemoveTask(task *model.TaskDTO) error {
	t.Lock()
	defer t.Unlock()
	snapTaskName := fmt.Sprintf("raintank-apps:%d", task.Id)
	log.Debug("removing snap task %s", snapTaskName)
	if err := t.removeActiveTask(snapTaskName); err != nil {
		return err
	}

	delete(t.Tasks, task.Id)
	return nil
}

func (t *TaskCache) removeActiveTask(taskName string) error {
	_, ok := t.ActiveTasks[taskName]
	if !ok {
		log.Debug("task to remove not in cache. %s", taskName)
	} else {
		delete(t.ActiveTasks, taskName)
	}
	return nil
}

var GlobalTaskCache *TaskCache

func InitTaskCache(tsdbURL *string) {
	GlobalTaskCache = &TaskCache{
		tsdbURL:     tsdbURL,
		Tasks:       make(map[int64]*model.TaskDTO),
		ActiveTasks: make(map[string]*ScheduledTask),
	}
}

func HandleTaskList() interface{} {
	return func(data []byte) {
		tasks := make([]*model.TaskDTO, 0)
		err := json.Unmarshal(data, &tasks)
		if err != nil {
			log.Error(3, "failed to decode taskUpdate payload. %s", err)
			return
		}
		log.Debug("TaskList. %s", data)
		GlobalTaskCache.UpdateTasks(tasks)
	}
}

func HandleTaskUpdate() interface{} {
	return func(data []byte) {
		task := model.TaskDTO{}
		err := json.Unmarshal(data, &task)
		if err != nil {
			log.Error(3, "failed to decode taskUpdate payload. %s", err)
			return
		}
		log.Debug("TaskUpdate. %s", data)
		if err := GlobalTaskCache.AddTask(&task); err != nil {
			log.Error(3, "failed to add task to cache. %s", err)
		}
	}
}

func HandleTaskAdd() interface{} {
	return func(data []byte) {
		task := model.TaskDTO{}
		err := json.Unmarshal(data, &task)
		if err != nil {
			log.Error(3, "failed to decode taskAdd payload. %s", err)
			return
		}
		log.Debug("Adding Task. %s", data)
		if err := GlobalTaskCache.AddTask(&task); err != nil {
			log.Error(3, "failed to add task to cache. %s", err)
		}
	}
}

func HandleTaskRemove() interface{} {
	return func(data []byte) {
		task := model.TaskDTO{}
		err := json.Unmarshal(data, &task)
		if err != nil {
			log.Error(3, "failed to decode taskAdd payload. %s", err)
			return
		}
		log.Debug("Removing Task. %s", data)
		if err := GlobalTaskCache.RemoveTask(&task); err != nil {
			log.Error(3, "failed to remove task from cache. %s", err)
		}
	}
}
