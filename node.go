package main

import (
	"log"

	"github.com/appwilldev/Instafig/conf"
	"github.com/appwilldev/Instafig/models"
)

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
