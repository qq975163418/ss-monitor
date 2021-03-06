package httphandler

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"strconv"
	logging "github.com/yanzay/log"
	"github.com/cool2645/ss-monitor/model"
	"github.com/cool2645/ss-monitor/manager"
	"encoding/json"
)

func GetTasks(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	req.ParseForm()
	class := "%"
	state := "%"
	ipVer := "%"
	nodeID := "%"
	callbackID := "%"
	order := "asc"
	var page uint = 1
	var perPage uint = 10
	if len(req.Form["class"]) == 1 {
		class = req.Form["class"][0]
	}
	if len(req.Form["state"]) == 1 {
		state = req.Form["state"][0]
	}
	if len(req.Form["ip_ver"]) == 1 {
		ipVer = req.Form["ip_ver"][0]
	}
	if len(req.Form["node_id"]) == 1 {
		nodeID = req.Form["node_id"][0]
	}
	if len(req.Form["callback_id"]) == 1 {
		callbackID = req.Form["callback_id"][0]
	}
	if len(req.Form["order"]) == 1 {
		order = req.Form["order"][0]
	}
	if len(req.Form["page"]) == 1 {
		page64, err := strconv.ParseUint(req.Form["page"][0], 10, 32)
		if err != nil {
			logging.Error(err)
		}
		page = uint(page64)
	}
	if len(req.Form["perPage"]) == 1 {
		perPage64, err := strconv.ParseUint(req.Form["perPage"][0], 10, 32)
		if err != nil {
			logging.Error(err)
		}
		perPage = uint(perPage64)
	}
	tasks, total, err := model.GetTasks(model.Db, class, state, ipVer, nodeID, callbackID, order, page, perPage)
	if err != nil {
		logging.Error(err)
		if err.Error() == "GetTasks: sql: no rows in result set" {
			res := map[string]interface{}{
				"code":   http.StatusNotFound,
				"result": false,
				"msg":    "Error occurred querying tasks: " + err.Error(),
			}
			responseJson(w, res, http.StatusNotFound)
			return
		}
		res := map[string]interface{}{
			"code":   http.StatusInternalServerError,
			"result": false,
			"msg":    "Error occurred querying tasks: " + err.Error(),
		}
		responseJson(w, res, http.StatusInternalServerError)
		return
	}
	if !checkAccessKey(req) && !checkAdmin(w, req) {
		for i, _ := range tasks {
			tasks[i].Node.Ss4Json = ""
			tasks[i].Node.Ss6Json = ""
			tasks[i].SsJson = ""
		}
	}
	data := map[string]interface{}{
		"total": total,
		"data":  tasks,
	}
	res := map[string]interface{}{
		"code":   http.StatusOK,
		"result": true,
		"data":   data,
	}
	responseJson(w, res, http.StatusOK)
}

func GetTask(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	req.ParseForm()
	taskID64, err := strconv.ParseUint(ps.ByName("id"), 10, 32)
	if err != nil {
		logging.Error(err)
		res := map[string]interface{}{
			"code":   http.StatusBadRequest,
			"result": false,
			"msg":    "Error occurred parsing task id.",
		}
		responseJson(w, res, http.StatusBadRequest)
		return
	}
	taskID := uint(taskID64)
	task, err := model.GetTask(model.Db, taskID)
	if err != nil {
		logging.Error(err)
		if err.Error() == "GetTask: record not found" {
			res := map[string]interface{}{
				"code":   http.StatusNotFound,
				"result": false,
				"msg":    "Error occurred querying tasks: " + err.Error(),
			}
			responseJson(w, res, http.StatusNotFound)
			return
		}
		res := map[string]interface{}{
			"code":   http.StatusInternalServerError,
			"result": false,
			"msg":    "Error occurred querying tasks: " + err.Error(),
		}
		responseJson(w, res, http.StatusInternalServerError)
		return
	}
	if !checkAccessKey(req) && !checkAdmin(w, req) {
		task.Node.Ss4Json = ""
		task.Node.Ss6Json = ""
		task.SsJson = ""
	}
	res := map[string]interface{}{
		"code":   http.StatusOK,
		"result": true,
		"data":   task,
	}
	responseJson(w, res, http.StatusOK)
}

func GetTaskLog(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	req.ParseForm()
	taskID64, err := strconv.ParseUint(ps.ByName("id"), 10, 32)
	if err != nil {
		logging.Error(err)
		res := map[string]interface{}{
			"code":   http.StatusBadRequest,
			"result": false,
			"msg":    "Error occurred parsing task id.",
		}
		responseJson(w, res, http.StatusBadRequest)
		return
	}
	taskID := uint(taskID64)
	task, err := model.GetTask(model.Db, taskID)
	if err != nil {
		logging.Error(err)
		if err.Error() == "GetTask: record not found" {
			res := map[string]interface{}{
				"code":   http.StatusNotFound,
				"result": false,
				"msg":    "Error occurred querying tasks: " + err.Error(),
			}
			responseJson(w, res, http.StatusNotFound)
			return
		}
		res := map[string]interface{}{
			"code":   http.StatusInternalServerError,
			"result": false,
			"msg":    "Error occurred querying tasks: " + err.Error(),
		}
		responseJson(w, res, http.StatusInternalServerError)
		return
	}
	res := map[string]interface{}{
		"code":   http.StatusOK,
		"result": true,
		"data":   task.Log,
	}
	responseJson(w, res, http.StatusOK)
}

func NewTask(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	var task model.Task
	if req.Header.Get("Content-Type") == "application/json" {
		err := json.NewDecoder(req.Body).Decode(&task)
		if err != nil {
			logging.Error(err)
			res := map[string]interface{}{
				"code":   http.StatusBadRequest,
				"result": false,
				"msg":    "Error occurred parsing json request.",
			}
			responseJson(w, res, http.StatusBadRequest)
			return
		}
		if task.NodeID != 0 && !authAdmin(w, req) {
			return
		}
	} else {
		req.ParseForm()
		if len(req.Form["class"]) != 1 {
			res := map[string]interface{}{
				"code":   http.StatusBadRequest,
				"result": false,
				"msg":    "Invalid worker class.",
			}
			responseJson(w, res, http.StatusBadRequest)
			return
		}
		class := req.Form["class"][0]
		var callbackID, nodeID, ipVer uint = 0, 0, 4
		if len(req.Form["callback_id"]) == 1 {
			callbackID64, err := strconv.ParseUint(req.Form["callback_id"][0], 10, 32)
			if err != nil {
				logging.Error(err)
				res := map[string]interface{}{
					"code":   http.StatusBadRequest,
					"result": false,
					"msg":    "Error occurred parsing callback id.",
				}
				responseJson(w, res, http.StatusBadRequest)
				return
			}
			callbackID = uint(callbackID64)
		}
		if len(req.Form["node_id"]) == 1 {
			nodeID64, err := strconv.ParseUint(req.Form["node_id"][0], 10, 32)
			if err != nil {
				logging.Error(err)
				res := map[string]interface{}{
					"code":   http.StatusBadRequest,
					"result": false,
					"msg":    "Error occurred parsing node id.",
				}
				responseJson(w, res, http.StatusBadRequest)
				return
			}
			nodeID = uint(nodeID64)
		}
		var serverName, ssJson string
		if nodeID == 0 {
			if len(req.Form["server_name"]) != 1 {
				res := map[string]interface{}{
					"code":   http.StatusBadRequest,
					"result": false,
					"msg":    "Invalid server name.",
				}
				responseJson(w, res, http.StatusBadRequest)
				return
			}
			serverName = req.Form["server_name"][0]
			if len(req.Form["ss_json"]) == 1 {
				ssJson = req.Form["ss_json"][0]
			}
		} else if !authAdmin(w, req) {
			return
		}
		if len(req.Form["ip_ver"]) == 1 {
			ipVer64, err := strconv.ParseUint(req.Form["ip_ver"][0], 10, 32)
			if err != nil {
				logging.Error(err)
				res := map[string]interface{}{
					"code":   http.StatusBadRequest,
					"result": false,
					"msg":    "Error occurred parsing ip version.",
				}
				responseJson(w, res, http.StatusBadRequest)
				return
			}
			ipVer = uint(ipVer64)
		}
		task.Class = class
		task.CallbackID = callbackID
		task.NodeID = nodeID
		task.IPVer = ipVer
		task.ServerName = serverName
		task.SsJson = ssJson
	}
	task, err := model.CreateTask(model.Db, task)
	if err != nil {
		logging.Error(err)
		res := map[string]interface{}{
			"code":   http.StatusInternalServerError,
			"result": false,
			"msg":    "Error occurred creating task: " + err.Error(),
		}
		responseJson(w, res, http.StatusInternalServerError)
		return
	}
	res := map[string]interface{}{
		"code":   http.StatusOK,
		"result": true,
		"data":   task,
	}
	responseJson(w, res, http.StatusOK)
}

func AssignTask(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	if !authAccessKey(w, req) {
		return
	}
	req.ParseForm()
	if len(req.Form["worker"]) != 1 {
		res := map[string]interface{}{
			"code":   http.StatusBadRequest,
			"result": false,
			"msg":    "Invalid worker name.",
		}
		responseJson(w, res, http.StatusBadRequest)
		return
	}
	worker := req.Form["worker"][0]
	taskID64, err := strconv.ParseUint(ps.ByName("id"), 10, 32)
	if err != nil {
		logging.Error(err)
		res := map[string]interface{}{
			"code":   http.StatusBadRequest,
			"result": false,
			"msg":    "Error occurred parsing task id.",
		}
		responseJson(w, res, http.StatusBadRequest)
		return
	}
	taskID := uint(taskID64)
	err = model.AssignTask(model.Db, taskID, worker)
	if err != nil {
		logging.Error(err)
		if err.Error() == "AssignTask: Check task status: Not queuing task" {
			res := map[string]interface{}{
				"code":   http.StatusOK,
				"result": false,
				"msg":    "Error occurred assigning task: " + err.Error(),
			}
			responseJson(w, res, http.StatusOK)
			return
		}
		res := map[string]interface{}{
			"code":   http.StatusInternalServerError,
			"result": false,
			"msg":    "Error occurred assigning task: " + err.Error(),
		}
		responseJson(w, res, http.StatusInternalServerError)
		return
	}
	res := map[string]interface{}{
		"code":   http.StatusOK,
		"result": true,
		"msg":    "success",
	}
	responseJson(w, res, http.StatusOK)
}

func SyncTaskStatus(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	if !authAccessKey(w, req) {
		return
	}
	req.ParseForm()
	if len(req.Form["worker"]) != 1 {
		res := map[string]interface{}{
			"code":   http.StatusBadRequest,
			"result": false,
			"msg":    "Invalid worker name.",
		}
		responseJson(w, res, http.StatusBadRequest)
		return
	}
	worker := req.Form["worker"][0]
	taskID64, err := strconv.ParseUint(ps.ByName("id"), 10, 32)
	if err != nil {
		logging.Error(err)
		res := map[string]interface{}{
			"code":   http.StatusBadRequest,
			"result": false,
			"msg":    "Error occurred parsing task id.",
		}
		responseJson(w, res, http.StatusBadRequest)
		return
	}
	taskID := uint(taskID64)
	var state, result, log string
	if len(req.Form["state"]) == 1 {
		state = req.Form["state"][0]
	}
	if len(req.Form["result"]) == 1 {
		result = req.Form["result"][0]
	}
	if len(req.Form["log"]) == 1 {
		log = req.Form["log"][0]
	}
	err = model.UpdateTaskStatus(model.Db, taskID, worker, state, result, log)
	if err != nil {
		logging.Error(err)
		if err.Error() == "SyncTaskStatus: Find task: record not found" {
			res := map[string]interface{}{
				"code":   http.StatusNotFound,
				"result": false,
				"msg":    "Error occurred updating task: " + err.Error(),
			}
			responseJson(w, res, http.StatusNotFound)
			return
		}
		if err.Error() == "SyncTaskStatus: Check task status: Not assigned worker" {
			res := map[string]interface{}{
				"code":   http.StatusUnauthorized,
				"result": false,
				"msg":    "Error occurred updating task: " + err.Error(),
			}
			responseJson(w, res, http.StatusUnauthorized)
			return
		}
		res := map[string]interface{}{
			"code":   http.StatusInternalServerError,
			"result": false,
			"msg":    "Error occurred updating task: " + err.Error(),
		}
		responseJson(w, res, http.StatusInternalServerError)
		return
	}
	res := map[string]interface{}{
		"code":   http.StatusOK,
		"result": true,
		"msg":    "success",
	}
	responseJson(w, res, http.StatusOK)
}

func ResetTask(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	if !authAdmin(w, req) {
		return
	}
	taskID64, err := strconv.ParseUint(ps.ByName("id"), 10, 32)
	if err != nil {
		logging.Error(err)
		res := map[string]interface{}{
			"code":   http.StatusBadRequest,
			"result": false,
			"msg":    "Error occurred parsing task id.",
		}
		responseJson(w, res, http.StatusBadRequest)
		return
	}
	taskID := uint(taskID64)
	err = model.ResetTask(model.Db, taskID)
	if err != nil {
		logging.Error(err)
		if err.Error() == "ResetTask: Find task: record not found" {
			res := map[string]interface{}{
				"code":   http.StatusNotFound,
				"result": false,
				"msg":    "Error occurred resetting task: " + err.Error(),
			}
			responseJson(w, res, http.StatusNotFound)
			return
		}
		res := map[string]interface{}{
			"code":   http.StatusInternalServerError,
			"result": false,
			"msg":    "Error occurred resetting task: " + err.Error(),
		}
		responseJson(w, res, http.StatusInternalServerError)
		return
	}
	res := map[string]interface{}{
		"code":   http.StatusOK,
		"result": true,
		"msg":    "success",
	}
	responseJson(w, res, http.StatusOK)
}

func TaskCallback(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	if !authAccessKey(w, req) {
		return
	}
	req.ParseForm()
	if len(req.Form["worker"]) != 1 {
		res := map[string]interface{}{
			"code":   http.StatusBadRequest,
			"result": false,
			"msg":    "Invalid worker name.",
		}
		responseJson(w, res, http.StatusBadRequest)
		return
	}
	worker := req.Form["worker"][0]
	taskID64, err := strconv.ParseUint(ps.ByName("id"), 10, 32)
	if err != nil {
		logging.Error(err)
		res := map[string]interface{}{
			"code":   http.StatusBadRequest,
			"result": false,
			"msg":    "Error occurred parsing task id.",
		}
		responseJson(w, res, http.StatusBadRequest)
		return
	}
	taskID := uint(taskID64)
	err = manager.TaskCallback(taskID, worker)
	if err != nil {
		logging.Error(err)
		if err.Error() == "TaskCallback: record not found" {
			res := map[string]interface{}{
				"code":   http.StatusNotFound,
				"result": false,
				"msg":    "Error occurred on task callback: " + err.Error(),
			}
			responseJson(w, res, http.StatusNotFound)
			return
		}
		if err.Error() == "TaskCallback: Not assigned worker" {
			res := map[string]interface{}{
				"code":   http.StatusUnauthorized,
				"result": false,
				"msg":    "Error occurred on task callback: " + err.Error(),
			}
			responseJson(w, res, http.StatusUnauthorized)
			return
		}
		res := map[string]interface{}{
			"code":   http.StatusInternalServerError,
			"result": false,
			"msg":    "Error occurred on task callback: " + err.Error(),
		}
		responseJson(w, res, http.StatusInternalServerError)
		return
	}
	res := map[string]interface{}{
		"code":   http.StatusOK,
		"result": true,
		"msg":    "success",
	}
	responseJson(w, res, http.StatusOK)
}
