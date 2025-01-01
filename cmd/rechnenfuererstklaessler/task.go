package main

import (
	"math/rand/v2"
	"time"
)

type TaskType int

const (
	Add10 = iota
	Add20
	AddQuest10
	AddQuest20
	Sub10
	Sub20
	SubQuest10
	SubQuest20
	Mix
	Add5Elem
	Mix5Elem
)

type Task struct {
	parameter1  int
	parameter2  int
	task_result int
	task_type   TaskType
}

func genTaks(taskType TaskType) Task {
	var task Task
	task.task_type = taskType

	var maxParameter = 10
	var maxSum = taskType.getMaxSum()

	var par1 = 1 + rand.IntN(maxParameter-1)
	var par2 int
	if maxSum-par1 == 1 {
		par2 = 1
	} else {
		par2 = 1 + rand.IntN(min(maxSum-par1-1, maxParameter))
	}
	var result = par1 + par2
	var ctaskType = taskType
	if taskType == Mix {
		var rndType = []TaskType{Add20, AddQuest20, Sub20, SubQuest20}
		ctaskType = rndType[rand.IntN(len(rndType))]
	}

	switch ctaskType {
	case Sub10:
		fallthrough
	case Sub20:
		fallthrough
	case SubQuest10:
		fallthrough
	case SubQuest20:
		var sum = result
		var add1 = par1
		var add2 = par2
		par1 = sum
		par2 = add1
		result = add2
	}

	task.parameter1 = par1
	task.parameter2 = par2
	task.task_result = result
	task.task_type = ctaskType
	return task
}

func (task Task) userExpectedResult() int {
	if task.task_type.isParameterQuest() {
		return task.parameter2
	} else {
		return task.task_result
	}
}

func (taskType TaskType) taskOp() string {
	switch taskType {
	case Add10:
		fallthrough
	case Add20:
		fallthrough
	case AddQuest10:
		fallthrough
	case AddQuest20:
		return "+"
	default:
		return "-"
	}
}

func (taskType TaskType) getMaxSum() int {
	var result int
	switch taskType {
	case Add10:
		fallthrough
	case Sub10:
		fallthrough
	case AddQuest10:
		fallthrough
	case SubQuest10:
		result = 10
	default:
		result = 20
	}
	return result
}

func (taskType TaskType) isParameterQuest() bool {
	switch taskType {
	case AddQuest10:
		fallthrough
	case AddQuest20:
		fallthrough
	case SubQuest10:
		fallthrough
	case SubQuest20:
		return true
	default:
		return false
	}
}

func (taskType TaskType) getTimeDisplayAbacus() time.Duration {
	switch taskType {
	case AddQuest10:
		fallthrough
	case AddQuest20:
		fallthrough
	case SubQuest10:
		fallthrough
	case SubQuest20:
		fallthrough
	case Sub10:
		fallthrough
	case Sub20:
		return TIME_DISPLAY_ABACUS * 2
	default:
		return TIME_DISPLAY_ABACUS
	}
}

func (task Task) checkResult(user_result int) bool {
	return task.userExpectedResult() == user_result
}
