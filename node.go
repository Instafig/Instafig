package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/appwilldev/Instafig/conf"
	"github.com/appwilldev/Instafig/models"
	"github.com/appwilldev/Instafig/utils"
	"github.com/gin-gonic/gin"
)

const (
	NODE_REQUEST_TYPE_SYNCSLAVE   = "SYNCSLAVE"
	NODE_REQUEST_TYPE_CHECKMASTER = "CHECKMASTER"
	NODE_REQUEST_TYPE_SYNCMASTER  = "SYNCMASTER"

	NODE_REQUEST_SYNC_TYPE_USER   = "USER"
	NODE_REQUEST_SYNC_TYPE_APP    = "APP"
	NODE_REQUEST_SYNC_TYPE_CONFIG = "CONFIG"
)


type syncDataT struct {
	DataVersion int    `json:"data_version"`
	Kind        string `json:"kind"`
	Data        string `json:"data"` // json string
}

type syncAllDataT struct {
	Nodes       map[string]*models.Node   `json:"nodes"`
	Users       map[string]*models.User   `json:"users"`
	Apps        map[string]*models.App    `json:"apps"`
	Configs     map[string]*models.Config `json:"configs"`
	DataVersion int                       `json:"data_version"`
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

func initLocalNodeData() {
	if memConfNodes[conf.ClientAddr] == nil {
		node := &models.Node{
			URL:         conf.ClientAddr,
			NodeURL:     conf.NodeAddr,
			Type:        conf.NodeType,
			DataVersion: memConfDataVersion,
			CreatedUTC:  utils.GetNowSecond(),
		}
		if err := models.InsertDBModel(nil, node); err != nil {
			log.Panicf("Failed to init node data: %s", err.Error())
		}
		memConfNodes[conf.ClientAddr] = node
	}

	node := memConfNodes[conf.ClientAddr]
	if node.Type != conf.NodeType {
		node.Type = conf.NodeType
		if err := models.UpdateDBModel(nil, node); err != nil {
			log.Panicf("Failed to update node data: %s", err.Error())
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
			errStr := fmt.Sprintf("data version of slave node [%s] is %d, master's data version is %d, can't sync", node.URL, node.DataVersion, ver)
			failedNodes = append(failedNodes, map[string]interface{}{"node": node, "err": errStr})
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
		kind = NODE_REQUEST_SYNC_TYPE_USER
	case *models.App:
		kind = NODE_REQUEST_SYNC_TYPE_APP
	case *models.Config:
		kind = NODE_REQUEST_SYNC_TYPE_CONFIG
	default:
		log.Panicln("unkown node data sync type: ", reflect.TypeOf(data))
	}

	bs, _ := json.Marshal(data)
	reqData := syncDataT{
		DataVersion: dataVer,
		Kind:        kind,
		Data:        string(bs),
	}

	_, err := nodeRequest(node.NodeURL, NODE_REQUEST_TYPE_SYNCSLAVE, reqData)

	return err
}

func slaveCheckMaster() error {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	memConfMux.RLock()
	node := memConfNodes[conf.ClientAddr]
	ver := memConfDataVersion
	localNode := memConfNodes[conf.ClientAddr]
	memConfMux.RUnlock()

	data, err := nodeRequest(conf.MasterAddr, NODE_REQUEST_TYPE_CHECKMASTER, node)
	if err != nil {
		return err
	}

	masterVer := int(data.(float64))
	if masterVer == ver {
		localNode.LastCheckUTC = utils.GetNowSecond()
		return models.UpdateDBModel(nil, localNode)
	}

	// slave's data_version not equals master's data_version, slave sync all data from master
	data, err = nodeRequest(conf.MasterAddr, NODE_REQUEST_TYPE_SYNCMASTER, nil)
	if err != nil {
		return err
	}

	resData := &syncAllDataT{}
	if err = json.Unmarshal([]byte(data.(string)), resData); err != nil {
		return fmt.Errorf("bad response data format: %s < %s >", err.Error(), data.(string))
	}

	users := make([]*models.User, 0)
	apps := make([]*models.App, 0)
	configs := make([]*models.Config, 0)
	nodes := make([]*models.Node, 0)

	s := models.NewSession()
	defer s.Close()
	if err = s.Begin(); err != nil {
		s.Rollback()
		return err
	}

	if err = models.ClearModeData(s); err != nil {
		s.Rollback()
		return err
	}

	for _, user := range resData.Users {
		if err = models.InsertDBModel(s, user); err != nil {
			s.Rollback()
			return err
		}
		users = append(users, user)
	}
	for _, app := range resData.Apps {
		if err = models.InsertDBModel(s, app); err != nil {
			s.Rollback()
			return err
		}
		apps = append(apps, app)
	}
	for _, config := range resData.Configs {
		if err = models.InsertDBModel(s, config); err != nil {
			s.Rollback()
			return err
		}
		configs = append(configs, config)
	}

	for _, node := range resData.Nodes {
		if node.URL == conf.ClientAddr {
			node = localNode
			localNode.DataVersion = resData.DataVersion
			node.LastCheckUTC = utils.GetNowSecond()
		}
		if err := models.InsertDBModel(s, node); err != nil {
			s.Rollback()
			return err
		}
		nodes = append(nodes, node)
	}

	if err = models.UpdateDataVersion(s, resData.DataVersion); err != nil {
		s.Rollback()
		return err
	}

	if err = s.Commit(); err != nil {
		s.Rollback()
		return err
	}

	fillMemConfData(users, apps, configs, nodes, resData.DataVersion)

	return nil
}

func nodeRequest(targetNodeUrl string, reqType string, data interface{}) (interface{}, error) {
	url := fmt.Sprintf("http://%s/node/req/%s", targetNodeUrl, reqType)
	var transData []byte
	var err error

	if data != nil {
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

	var resData struct {
		Status bool        `json:"status"`
		Data   interface{} `json:"data"`
		Code   string      `json:"code"`
	}

	err = json.Unmarshal(resBody, &resData)
	if err != nil {
		return nil, fmt.Errorf("bad reponse body format: %s", err.Error())
	}

	if !resData.Status {
		return nil, fmt.Errorf(resData.Code)
	}

	return resData.Data, nil
}

func NodeRequestHandler(c *gin.Context) {
	switch c.Param("req_type") {
	case NODE_REQUEST_TYPE_SYNCSLAVE:
		handleSlaveSyncUpdateData(c)
	case NODE_REQUEST_TYPE_CHECKMASTER:
		handleSlaveCheckMaster(c)
	case NODE_REQUEST_TYPE_SYNCMASTER:
		handleSyncMaster(c)
	default:
		Error(c, BAD_REQUEST, "unkown node request type")
	}
}

func handleSlaveSyncUpdateData(c *gin.Context) {
	if conf.IsMasterNode() {
		Error(c, BAD_REQUEST, "invalid req type for master node: "+NODE_REQUEST_TYPE_SYNCSLAVE)
		return
	}

	reqBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Error(c, BAD_REQUEST, "can read req body")
		return
	}

	reqData := &syncDataT{}
	if err = json.Unmarshal(reqBody, reqData); err != nil {
		Error(c, BAD_REQUEST, "bad req body format")
		return
	}

	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	memConfMux.RLock()
	ver := memConfDataVersion
	memConfMux.RUnlock()

	if ver+1 != reqData.DataVersion {
		Error(c, DATA_VERSION_ERROR, "slave node data version [%d] error for master data version [%d]", ver, reqData.DataVersion)
		return
	}

	switch reqData.Kind {
	case NODE_REQUEST_SYNC_TYPE_USER:
		user := &models.User{}
		if err = json.Unmarshal([]byte(reqData.Data), user); err != nil {
			Error(c, BAD_REQUEST, "bad data format for user model")
			return
		}
		if _, err = updateUser(user); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}
		Success(c, nil)

	case NODE_REQUEST_SYNC_TYPE_APP:
		app := &models.App{}
		if err = json.Unmarshal([]byte(reqData.Data), app); err != nil {
			Error(c, BAD_REQUEST, "bad data format for app model")
			return
		}
		if _, err = updateApp(app); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}
		Success(c, nil)

	case NODE_REQUEST_SYNC_TYPE_CONFIG:
		config := &models.Config{}
		if err = json.Unmarshal([]byte(reqData.Data), config); err != nil {
			Error(c, BAD_REQUEST, "bad data format for user model")
			return
		}
		if _, err = updateConfig(config); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}
		Success(c, nil)

	default:
		Error(c, BAD_REQUEST, "unkown node data sync type: "+reqData.Kind)
		return
	}
}

func handleSlaveCheckMaster(c *gin.Context) {
	if !conf.IsMasterNode() {
		Error(c, BAD_REQUEST, "invalid req type for slave node: "+NODE_REQUEST_TYPE_CHECKMASTER)
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

	node.LastCheckUTC = utils.GetNowSecond()
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
	ver := memConfDataVersion
	memConfMux.Unlock()

	Success(c, ver)
}

func handleSyncMaster(c *gin.Context) {
	if !conf.IsMasterNode() {
		Error(c, BAD_REQUEST, "invalid req type for slave node: "+NODE_REQUEST_TYPE_SYNCMASTER)
		return
	}

	// no need to hold the locker(confWriteMux) to avoid dead-lock, slave will eventually be consistent with master,
	//	confWriteMux.Lock()
	//	defer confWriteMux.Unlock()

	memConfMux.RLock()
	nodes := memConfNodes
	users := memConfUsers
	apps := memConfApps
	configs := memConfRawConfigs
	ver := memConfDataVersion
	resData, _ := json.Marshal(syncAllDataT{
		Nodes:       nodes,
		Users:       users,
		Apps:        apps,
		Configs:     configs,
		DataVersion: ver,
	})
	memConfMux.RUnlock()

	Success(c, string(resData))
}
