package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/appwilldev/Instafig/conf"
	"github.com/appwilldev/Instafig/models"
	"github.com/appwilldev/Instafig/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

const (
	NODE_REQUEST_TYPE_SYNCSLAVE   = "SYNCSLAVE"
	NODE_REQUEST_TYPE_CHECKMASTER = "CHECKMASTER"
	NODE_REQUEST_TYPE_SYNCMASTER  = "SYNCMASTER"

	NODE_REQUEST_SYNC_TYPE_USER    = "USER"
	NODE_REQUEST_SYNC_TYPE_APP     = "APP"
	NODE_REQUEST_SYNC_TYPE_WEBHOOK = "WEBHOOK"
	NODE_REQUEST_SYNC_TYPE_CONFIG  = "CONFIG"
	NODE_REQUEST_SYNC_TYPE_NODE    = "NODE"
)

var (
	nodeAuthString string
)

type syncDataT struct {
	DataVersion *models.DataVersion `json:"data_version"`
	Kind        string              `json:"kind"`
	Data        string              `json:"data"` // json string to bind go struct
	OpUserKey   string              `json:"op_user_key"`
}

type syncAllDataT struct {
	Nodes       map[string]*models.Node       `json:"nodes"`
	Users       map[string]*models.User       `json:"users"`
	Apps        map[string]*models.App        `json:"apps"`
	WebHooks    []*models.WebHook             `json:"web_hooks"`
	Configs     map[string]*models.Config     `json:"configs"`
	ConfHistory []*models.ConfigUpdateHistory `json:"conf_history"`
	DataVersion *models.DataVersion           `json:"data_version"`
}

type nodeRequestDataT struct {
	Auth string `json:"auth"`
	Data string `json:"data"` // json string to bind go struct
}

func init() {
	var err error
	if conf.IsEasyDeployMode() {
		nodeAuthToken := jwt.New(jwt.SigningMethodHS256)
		if nodeAuthString, err = nodeAuthToken.SignedString([]byte(conf.NodeAuth)); err != nil {
			log.Panicf("Failed to init node auth token: %s", err.Error())
		}

		checkNodeValidity()
		loadAllData()
		initNodeData()

		if !conf.IsMasterNode() {
			if err = slaveCheckMaster(); err != nil {
				log.Printf("slave node failed to check master: %s", err.Error())
				os.Exit(1)
			}

			go func() {
				for {
					time.Sleep(time.Duration(conf.CheckMasterInerval) * time.Second)
					go slaveCheckMaster()
				}
			}()
		}
	}
}

func checkNodeValidity() {
	nodes, err := models.GetAllNode(nil)
	if err != nil {
		log.Panicf("failed to check node validity: %s" + err.Error())
	}

	for _, node := range nodes {
		if conf.IsMasterNode() {
			// only one master in cluster
			if node.Type == models.NODE_TYPE_MASTER && node.URL != conf.ClientAddr {
				if err := models.DeleteDBModel(nil, node); err != nil {
					log.Panicf("failed to check node validity: %s" + err.Error())
				}
				break
			}
		} else {
			if node.Type == models.NODE_TYPE_MASTER && node.NodeURL != conf.MasterAddr {
				// this node is attached to a new master, sync full data from new master
				// just clear old-master data here, slave will sync new-master's data before serve for client
				if err = models.ClearModeData(nil); err != nil {
					log.Panicf("failed to check node validity: %s" + err.Error())
				}
				break
			}
		}
	}
}

func getMasterNode() models.Node {
	if !conf.IsEasyDeployMode() {
		return models.Node{}
	}

	memConfMux.RLock()
	defer memConfMux.RUnlock()

	for _, node := range memConfNodes {
		if node.Type == models.NODE_TYPE_MASTER {
			return *node
		}
	}

	log.Panicf("No master node found")

	return models.Node{}
}

func initNodeData() {
	if memConfNodes[conf.ClientAddr] == nil {
		bs, _ := json.Marshal(memConfDataVersion)
		node := &models.Node{
			URL:            conf.ClientAddr,
			NodeURL:        conf.NodeAddr,
			Type:           conf.NodeType,
			DataVersion:    memConfDataVersion,
			DataVersionStr: string(bs),
			CreatedUTC:     utils.GetNowSecond(),
		}

		if err := models.InsertRow(nil, node); err != nil {
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

	if !conf.IsMasterNode() && memConfNodes[conf.MasterAddr] == nil {
		// fake master node
		masterNode := &models.Node{
			URL:  conf.MasterAddr,
			Type: models.NODE_TYPE_MASTER,
		}
		if err := models.InsertRow(nil, masterNode); err != nil {
			log.Panicf("Failed to init master node data: %s", err.Error())
		}
		memConfNodes[conf.MasterAddr] = masterNode
	}
}

func updateNodeDataVersion(s *models.Session, node *models.Node, ver *models.DataVersion) (err error) {
	if !conf.IsEasyDeployMode() {
		return
	}

	var _s *models.Session
	var bs []byte

	if s == nil {
		_s = models.NewSession()
		defer _s.Close()
		if err = _s.Begin(); err != nil {
			goto ERROR
		}
	} else {
		_s = s
	}

	bs, _ = json.Marshal(ver)
	node.DataVersion = ver
	node.DataVersionStr = string(bs)
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

func syncData2SlaveIfNeed(data interface{}, opUserKey string) []map[string]interface{} {
	if !conf.IsEasyDeployMode() {
		return nil
	}

	var failedNodes []map[string]interface{}
	for _, node := range memConfNodes {
		if node.Type == models.NODE_TYPE_MASTER {
			continue
		}

		if memConfDataVersion.Version != node.DataVersion.Version+1 {
			errStr := fmt.Sprintf("data_version error: slave node's data_version [%s] is %d, master's data_version is %d", node.URL, node.DataVersion.Version, memConfDataVersion.Version)
			failedNodes = append(failedNodes, map[string]interface{}{"node": node, "err": errStr})
			continue
		}

		if err := syncData2Slave(node, data, memConfDataVersion, opUserKey); err != nil {
			failedNodes = append(failedNodes, map[string]interface{}{"node": node, "err": err.Error()})
		}
	}

	return failedNodes
}

func syncData2Slave(node *models.Node, data interface{}, dataVer *models.DataVersion, opUserKey string) error {
	kind := ""
	switch data.(type) {
	case *models.User:
		kind = NODE_REQUEST_SYNC_TYPE_USER
	case *models.App:
		kind = NODE_REQUEST_SYNC_TYPE_APP
	case *models.WebHook:
		kind = NODE_REQUEST_SYNC_TYPE_WEBHOOK
	case *models.Config:
		kind = NODE_REQUEST_SYNC_TYPE_CONFIG
	case *models.Node:
		kind = NODE_REQUEST_SYNC_TYPE_NODE
	default:
		log.Panicln("unknown node data sync type: ", reflect.TypeOf(data))
	}

	bs, _ := json.Marshal(data)
	syncDataString, _ := json.Marshal(&syncDataT{
		DataVersion: dataVer,
		Kind:        kind,
		Data:        string(bs),
		OpUserKey:   opUserKey,
	})

	reqData := nodeRequestDataT{
		Auth: nodeAuthString,
		Data: string(syncDataString),
	}
	_, err := nodeRequest(node.NodeURL, NODE_REQUEST_TYPE_SYNCSLAVE, reqData)

	if err == nil && kind != NODE_REQUEST_SYNC_TYPE_NODE {
		// update slave data version here
		dataVersionStr, _ := json.Marshal(dataVer)
		memConfMux.Lock()
		node.DataVersion = dataVer
		node.DataVersionStr = string(dataVersionStr)
		node.LastCheckUTC = utils.GetNowSecond()
		memConfMux.Unlock()
		updateNodeDataVersion(nil, node, dataVer)
	}

	return err
}

func slaveCheckMaster() error {
	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	localNode := *memConfNodes[conf.ClientAddr]
	nodeString, _ := json.Marshal(localNode)
	reqData := nodeRequestDataT{
		Auth: nodeAuthString,
		Data: string(nodeString),
	}
	data, err := nodeRequest(conf.MasterAddr, NODE_REQUEST_TYPE_CHECKMASTER, reqData)
	if err != nil {
		return err
	}

	masterVersion := &models.DataVersion{}
	if err = json.Unmarshal([]byte(data.(string)), masterVersion); err != nil {
		return fmt.Errorf("bad response data format: %s < %s >", err.Error(), data.(string))
	}

	if masterVersion.Version == memConfDataVersion.Version && masterVersion.Sign == memConfDataVersion.Sign {
		localNode.LastCheckUTC = utils.GetNowSecond()
		if err = models.UpdateDBModel(nil, &localNode); err != nil {
			return err
		}

		memConfMux.Lock()
		memConfNodes[conf.ClientAddr] = &localNode
		memConfMux.Unlock()

		return nil
	}

	reqData = nodeRequestDataT{
		Auth: nodeAuthString,
		Data: "",
	}
	// slave's data_version not equals master's data_version, slave sync all data from master
	data, err = nodeRequest(conf.MasterAddr, NODE_REQUEST_TYPE_SYNCMASTER, reqData)
	if err != nil {
		return err
	}

	resData := &syncAllDataT{}
	if err = json.Unmarshal([]byte(data.(string)), resData); err != nil {
		return fmt.Errorf("bad response data format: %s < %s >", err.Error(), data.(string))
	}

	var users []*models.User
	var apps []*models.App
	var configs []*models.Config
	var nodes []*models.Node

	bs, _ := json.Marshal(resData.DataVersion)
	localNode.DataVersion = resData.DataVersion
	localNode.DataVersionStr = string(bs)
	localNode.LastCheckUTC = utils.GetNowSecond()

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

	toInsertModels := make([]interface{}, 0)
	for _, node := range resData.Nodes {
		if node.URL == conf.ClientAddr {
			node.DataVersion = localNode.DataVersion
			node.DataVersionStr = localNode.DataVersionStr
			node.LastCheckUTC = localNode.LastCheckUTC
		}
		toInsertModels = append(toInsertModels, node)
		nodes = append(nodes, node)
	}
	if err := models.InsertMultiRows(s, toInsertModels); err != nil {
		s.Rollback()
		return err
	}

	toInsertModels = make([]interface{}, 0)
	for _, user := range resData.Users {
		users = append(users, user)
		toInsertModels = append(toInsertModels, user)
	}
	if err = models.InsertMultiRows(s, toInsertModels); err != nil {
		s.Rollback()
		return err
	}

	toInsertModels = make([]interface{}, 0)
	for _, app := range resData.Apps {
		toInsertModels = append(toInsertModels, app)
		apps = append(apps, app)
	}
	if err = models.InsertMultiRows(s, toInsertModels); err != nil {
		s.Rollback()
		return err
	}

	toInsertModels = make([]interface{}, len(resData.WebHooks))
	for ix, hook := range resData.WebHooks {
		toInsertModels[ix] = hook
	}
	if err = models.InsertMultiRows(s, toInsertModels); err != nil {
		s.Rollback()
		return err
	}

	toInsertModels = make([]interface{}, 0)
	for _, config := range resData.Configs {
		toInsertModels = append(toInsertModels, config)
		configs = append(configs, config)
	}
	if err = models.InsertMultiRows(s, toInsertModels); err != nil {
		s.Rollback()
		return err
	}

	toInsertModels = make([]interface{}, len(resData.ConfHistory))
	for ix, history := range resData.ConfHistory {
		toInsertModels[ix] = history
	}
	if err = models.InsertMultiRows(s, toInsertModels); err != nil {
		s.Rollback()
		return err
	}

	if err = models.UpdateDataVersion(s, resData.DataVersion); err != nil {
		s.Rollback()
		return err
	}

	if err = s.Commit(); err != nil {
		s.Rollback()
		return err
	}

	fillMemConfData(users, apps, resData.WebHooks, configs, nodes, resData.DataVersion)

	nodeString, _ = json.Marshal(&localNode)
	reqData = nodeRequestDataT{
		Auth: nodeAuthString,
		Data: string(nodeString),
	}

	nodeRequest(conf.MasterAddr, NODE_REQUEST_TYPE_CHECKMASTER, reqData)

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
				return nil, fmt.Errorf("bad data format: %s ", err.Error())
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
	reqBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		Error(c, BAD_REQUEST, "can read req body")
		return
	}

	reqData := &nodeRequestDataT{}
	if err = json.Unmarshal(reqBody, reqData); err != nil {
		Error(c, BAD_REQUEST, "bad req body format")
		return
	}

	if err = nodeAuth(reqData.Auth); err != nil {
		Error(c, NOT_PERMITTED, err.Error())
		return
	}

	switch c.Param("req_type") {
	case NODE_REQUEST_TYPE_SYNCSLAVE:
		handleSlaveSyncUpdateData(c, reqData.Data)
	case NODE_REQUEST_TYPE_CHECKMASTER:
		handleSlaveCheckMaster(c, reqData.Data)
	case NODE_REQUEST_TYPE_SYNCMASTER:
		handleSyncMaster(c, reqData.Data)
	default:
		Error(c, BAD_REQUEST, "unknown node request type")
	}
}

func handleSlaveSyncUpdateData(c *gin.Context, data string) {
	if conf.IsMasterNode() {
		Error(c, BAD_REQUEST, "invalid req type for master node: "+NODE_REQUEST_TYPE_SYNCSLAVE)
		return
	}

	syncData := &syncDataT{}
	err := json.Unmarshal([]byte(data), syncData)
	if err != nil {
		Error(c, BAD_REQUEST, "bad req body format")
		return
	}

	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	if syncData.Kind != NODE_REQUEST_SYNC_TYPE_NODE {
		if memConfDataVersion.Version+1 != syncData.DataVersion.Version {
			Error(c, DATA_VERSION_ERROR, "slave node data version [%d] error for master data version [%d]", memConfDataVersion.Version, syncData.DataVersion.Version)
			return
		}
		if memConfDataVersion.Sign != syncData.DataVersion.OldSign {
			Error(c, DATA_VERSION_ERROR, "slave node's data sign [%s] not equal master node's old data sign [%s]", memConfDataVersion.Sign, syncData.DataVersion.OldSign)
			return
		}
	}

	switch syncData.Kind {
	case NODE_REQUEST_SYNC_TYPE_USER:
		user := &models.User{}
		if err = json.Unmarshal([]byte(syncData.Data), user); err != nil {
			Error(c, BAD_REQUEST, "bad data format for user model")
			return
		}
		if _, err = updateUser(user, syncData.DataVersion); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}

	case NODE_REQUEST_SYNC_TYPE_APP:
		app := &models.App{}
		if err = json.Unmarshal([]byte(syncData.Data), app); err != nil {
			Error(c, BAD_REQUEST, "bad data format for app model")
			return
		}
		if _, err = updateApp(app, syncData.DataVersion); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}

	case NODE_REQUEST_SYNC_TYPE_WEBHOOK:
		hook := &models.WebHook{}
		if err = json.Unmarshal([]byte(syncData.Data), hook); err != nil {
			Error(c, BAD_REQUEST, "bad data format for webHook model")
			return
		}
		if _, err = updateWebHook(hook, syncData.DataVersion); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}

	case NODE_REQUEST_SYNC_TYPE_CONFIG:
		config := &models.Config{}
		if err = json.Unmarshal([]byte(syncData.Data), config); err != nil {
			Error(c, BAD_REQUEST, "bad data format for user model")
			return
		}
		if _, err = updateConfig(config, syncData.OpUserKey, syncData.DataVersion); err != nil {
			Error(c, SERVER_ERROR, err.Error())
			return
		}

	case NODE_REQUEST_SYNC_TYPE_NODE:
		node := &models.Node{}
		if err = json.Unmarshal([]byte(syncData.Data), node); err != nil {
			Error(c, BAD_REQUEST, "bad data format for node model")
			return
		}

		if memConfNodes[node.URL] == nil {
			if err := models.InsertRow(nil, node); err != nil {
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
		return

	default:
		Error(c, BAD_REQUEST, "unknown node data sync type: "+syncData.Kind)
		return
	}

	masterNode := getMasterNode()
	masterNode.DataVersion = syncData.DataVersion
	bs, _ := json.Marshal(syncData.DataVersion)
	masterNode.DataVersionStr = string(bs)
	if err = models.UpdateDBModel(nil, &masterNode); err != nil {
		memConfMux.Lock()
		memConfNodes[masterNode.URL] = &masterNode
		memConfMux.Unlock()
	}

	Success(c, nil)
}

func handleSlaveCheckMaster(c *gin.Context, data string) {
	if !conf.IsMasterNode() {
		Error(c, BAD_REQUEST, "invalid req type for slave node: "+NODE_REQUEST_TYPE_CHECKMASTER)
		return
	}

	node := &models.Node{}
	if err := json.Unmarshal([]byte(data), node); err != nil {
		Error(c, BAD_REQUEST, "bad req body format")
		return
	}

	confWriteMux.Lock()
	defer confWriteMux.Unlock()

	oldNode := memConfNodes[node.URL]
	node.LastCheckUTC = utils.GetNowSecond()
	if oldNode == nil {
		if err := models.InsertRow(nil, node); err != nil {
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
	bs, _ := json.Marshal(memConfDataVersion)
	memConfMux.Unlock()

	go masterSyncNodeToSlave(node)

	Success(c, string(bs))
}

func handleSyncMaster(c *gin.Context, data string) {
	if !conf.IsMasterNode() {
		Error(c, BAD_REQUEST, "invalid req type for slave node: "+NODE_REQUEST_TYPE_SYNCMASTER)
		return
	}

	// no need to hold the locker(confWriteMux) to avoid dead-lock, slave will eventually be consistent with master,
	//	confWriteMux.Lock()
	//	defer confWriteMux.Unlock()

	history, err := models.GetAllConfigUpdateHistory(nil)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}

	memConfMux.RLock()
	webHooks := memConfGlobalWebHooks
	for _, hooks := range memConfAppWebHooks {
		webHooks = append(webHooks, hooks...)
	}
	resData, _ := json.Marshal(syncAllDataT{
		Nodes:       memConfNodes,
		Users:       memConfUsers,
		Apps:        memConfApps,
		WebHooks:    webHooks,
		Configs:     memConfRawConfigs,
		DataVersion: memConfDataVersion,
		ConfHistory: history,
	})
	memConfMux.RUnlock()

	Success(c, string(resData))
}

func masterSyncNodeToSlave(node *models.Node) {
	nodes := make([]*models.Node, 0)
	memConfMux.RLock()
	for _, _node := range memConfNodes {
		nodes = append(nodes, _node)
	}
	memConfMux.RUnlock()

	for _, _node := range nodes {
		if _node.URL == node.URL || _node.Type == models.NODE_TYPE_MASTER {
			continue
		}
		if err := syncData2Slave(_node, node, nil, ""); err != nil {
			log.Println("Failed to sync slave node to other slaves: " + err.Error())
		}
	}
}

func nodeAuth(authString string) error {
	token, err := jwt.Parse(authString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(conf.NodeAuth), nil
	})
	if err != nil {
		return err
	}

	if token.Valid {
		return nil
	}

	return fmt.Errorf("invalid node auth")
}
