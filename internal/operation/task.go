package operation

import "github.com/vilasle/backilli/internal/model"

func RunTask(taskId string) error {
	//get task from database by id

	var task model.Task
	// task, err := ___GetTaskById(taskId)
	// if err != nil {
	// 	return err
	// }

	//check if task pass filter then continue task
	if !task.NeedToRun() {
		return nil
	} 

	report := NewReport()
	//loop by source
	for _, s := task.Source() {
		//copy-compress-split on temp directory and move it on volume
		// or move from volume to temp directory, merge-extract-restore to source
		result, err = task.HandleSource(s)
		report.Add(s, result, err)
	}
	return nil
}
