package main

import (
	"log"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/appwilldev/Instafig/conf"
	"github.com/appwilldev/Instafig/models"
	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin"
)

func init() {
	if conf.IsEasyDeployMode() && !conf.IsMasterNode() {
		go func() {
			for {
				slaveCheckMaster()
				time.Sleep(60 * time.Second)
			}
		}()
	}
}

func checkNodeValidity() {
	if !conf.IsMasterNode() {
		return
	}

	nodes, err := models.GetAllNode(nil)
	if err != nil {
		log.Panicf("failed to get node info when init data: %s" + err.Error())
	}

	for _, node := range nodes {
		if node.Type == models.NODE_TYPE_MASTER && node.URL != conf.ClientAddr {
			if !conf.ReplaceMaster {
				log.Panicf("master[ %s ] already exists, you can start service with --replace-master to replace old master if need", node.URL)
			}
		} else {
			if err := models.DeleteDBModel(nil, node); err != nil {
				log.Panicf("failed to remove old master node: ", err.Error())
			}
			break
		}
	}
}

func updateNodeDataVersion(s *models.Session, node *models.Node, ver int) (err error) {
	if !conf.IsEasyDeployMode() {
		return
	}

	var _s *models.Session

	if s == nil {
		_s = models.NewSession()
		defer s.Close()
		if err = s.Begin(); err != nil {
			goto ERROR
		}

		confWriteMux.Lock()
		defer confWriteMux.Unlock()
	} else {
		_s = s
	}

	node.DataVersion = ver
	if err = models.UpdateDBModel(_s, node); err != nil {
		goto ERROR
	}

	if node.URL == conf.ClientAddr {
		if err = models.UpdateDataVersion(_s, ver); err != nil {
			goto ERROR
		}
	}

	if s != nil {
		return
	}

	if err = _s.Commit(); err != nil {
		goto ERROR
	}

	return
ERROR:
	if s == nil {
		_s.Rollback()
	}

	return
}

func syncData2SlaveIfNeed(data interface{}) []map[string]interface{} {
	if !conf.IsEasyDeployMode() {
		return nil
	}

	memConfMux.RLock()
	ver := memConfDataVersion
	nodes := memConfNodes
	memConfMux.RUnlock()

	failedNodes := make([]map[string]interface{}, 0)
	for _, node := range nodes {
		if node.Type == models.NODE_TYPE_MASTER {
			continue
		}

		if ver != node.DataVersion+1 {
			// failedNodes[node] = fmt.Sprintf("data version of slave node [%s] is %d, master's data version is %d, can't sync", node.DataVersion, ver)
			continue
		}

		if err := syncData2Slave(node, data, ver); err != nil {
			failedNodes = append(failedNodes, map[string]interface{}{"node": node, "err": err.Error()})
		}
	}

	return failedNodes
}

func syncData2Slave(node *models.Node, data interface{}, dataVer int) error {
	kind := ""
	switch data.(type) {
	case *models.User:
		kind = NODE_REQUEST_DATA_SYNC_KIND_USER
	case *models.App:
		kind = NODE_REQUEST_DATA_SYNC_KIND_APP
	case *models.Config:
		kind = NODE_REQUEST_DATA_SYNC_KIND_CONFIG
	default:
		log.Panicln("unkown node data sync type: ", reflect.TypeOf(data))
	}

	reqData := map[string]interface{}{
		"data_ver": dataVer,
		"kind":     kind,
		"data":     data,
	}

	_, err := nodeRequest(node.NodeURL, NODE_REQUEST_TYPE_SYNC2SLAVE, reqData)

	return err
}

func slaveCheckMaster() (err error) {
	memConfMux.RLock()
	node := memConfNodes[conf.ClientAddr]
	memConfMux.RUnlock()

	_, err = nodeRequest(conf.MasterAddr, NODE_REQUEST_TYPE_SLAVECHECKMASTER, node)
	return
}

const (
	NODE_REQUEST_TYPE_SYNC2SLAVE       = "SYNC2SLAVE"
	NODE_REQUEST_TYPE_SLAVECHECKMASTER = "SLAVECHECKMASTER"

	NODE_REQUEST_DATA_SYNC_KIND_USER   = "USER"
	NODE_REQUEST_DATA_SYNC_KIND_APP    = "APP"
	NODE_REQUEST_DATA_SYNC_KIND_CONFIG = "CONFIG"
)

func nodeRequest(targetNodeUrl string, reqType string, data interface{}) (interface{}, error) {
	url := fmt.Sprintf("http://%s/node/req/%s", targetNodeUrl, reqType)
	var transData []byte
	var err error
	switch data.(type) {
	case string:
		transData = []byte(data.(string))
	case []byte:
		transData = data.([]byte)
	default:
		transData, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("bad data format: ", err.Error())
		}
	}

	b := bytes.NewReader(transData)
	res, err := http.Post(url, "plain/text", b)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to call [%s], status code: %d", url, res.StatusCode)
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read reponse body data: %s", err.Error())
	}

	j, err := simplejson.NewJson(resBody)
	if err != nil {
		return nil, fmt.Errorf("bad reponse body format: %s", err.Error())
	}

	if !j.Get("status").MustBool() {
		return nil, fmt.Errorf(j.Get("code").MustString())
	}

	return j.Get("data").Interface(), nil
}

func NodeRequestHandler(c *gin.Context) {
	switch c.Param("req_type") {
	case NODE_REQUEST_TYPE_SYNC2SLAVE:
		handleSlaveSyncUpdateData(c)
	case NODE_REQUEST_TYPE_SLAVECHECKMASTER:
		handleSlaveCheckMaster(c)
	default:
		Error(c, BAD_REQUEST, "unkown node request type")
	}
}

func handleSlaveSyncUpdateData(c *gin.Context) {
	if conf.IsMasterNode() {
		Error(c, BAD_REQUEST, "invalid req type for master node: "+NODE_REQUEST_TYPE_SYNC2SLAVE)
		return
	}

	reqBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Error(c, BAD_REQUEST, "can read req body")
		return
	}

	j, err := simplejson.NewJson(reqBody)
	if err != nil {
		Error(c, BAD_REQUEST, "bad req body format")
		return
	}

	kind := j.Get("kind").MustString()
	data := j.Get("data").MustMap()
	ver := j.Get("data_ver").MustInt()

	//TODO: get lock to
	confWriteMux.Lock()
	defer confWriteMux.Unlock()
	if memConfDataVersion+1 != ver {
		Error(c, BAD_REQUEST, "slave node data version [%d] error for master data version [%d]", memConfDataVersion, ver)
	}

	switch kind {
	case NODE_REQUEST_DATA_SYNC_KIND_USER:
		user := &models.User{
			Key:  data["key"].(string),
			Name: data["name"].(string),
		}
		if _, err = updateUser(user); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}
		Success(c, nil)

	case NODE_REQUEST_DATA_SYNC_KIND_APP:
		app := &models.App{
			Key:     data["key"].(string),
			UserKey: data["user_key"].(string),
			Name:    data["name"].(string),
			Type:    data["type"].(string),
		}
		if _, err = updateApp(app); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}
		Success(c, nil)

	case NODE_REQUEST_DATA_SYNC_KIND_CONFIG:
		config := &models.Config{
			Key:    data["key"].(string),
			AppKey: data["app_key"].(string),
			K:      data["k"].(string),
			V:      data["v"].(string),
			VType:  data["v_type"].(string),
		}
		if _, err = updateConfig(config); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}
		Success(c, nil)

	default:
		Error(c, BAD_REQUEST, "unkown node data sync type: "+kind)
		return
	}
}

func handleSlaveCheckMaster(c *gin.Context) {
	if !conf.IsMasterNode() {
		Error(c, BAD_REQUEST, "invalid req type for slave node: "+NODE_REQUEST_TYPE_SLAVECHECKMASTER)
		return
	}

	reqBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Error(c, BAD_REQUEST, "can read req body")
		return
	}

	node := &models.Node{}
	if err = json.Unmarshal(reqBody, node); err != nil {
		Error(c, BAD_REQUEST, "bad req body format")
		return
	}

	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	memConfMux.RLock()
	oldNode := memConfNodes[node.URL]
	memConfMux.RUnlock()

	if oldNode == nil {
		if err := models.InsertDBModel(nil, node); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}
	} else {
		if err := models.UpdateDBModel(nil, node); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}
	}

	memConfMux.Lock()
	memConfNodes[node.URL] = node
	memConfMux.Unlock()

	Success(c, nil)
}
