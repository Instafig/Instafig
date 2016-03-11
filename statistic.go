package main

import (
	"log"
	"os"
	"time"

	"fmt"

	"strconv"

	"github.com/appwilldev/Instafig/conf"
	"github.com/gin-gonic/gin"
	influx "github.com/influxdata/influxdb/client/v2"
)

var (
	statisticCh       = make(chan *influx.Point, 100000)
	influxClient      influx.Client
	influxBatchPoints influx.BatchPoints
	microSecondUnit   = time.Microsecond / time.Nanosecond
	secondUnit        = time.Second / time.Nanosecond
)

func init() {
	if conf.StatisticEnable {
		var err error
		influxClient, err = influx.NewHTTPClient(
			influx.HTTPConfig{
				Addr:     conf.InfluxURL,
				Username: conf.InfluxUser,
				Password: conf.InfluxPassword,
			})
		if err != nil {
			log.Println("Failed to init influx client: ", err.Error())
			os.Exit(1)
		}

		influxBatchPoints, err = influx.NewBatchPoints(influx.BatchPointsConfig{
			Database:  conf.InfluxDB,
			Precision: "s",
		})
		if err != nil {
			log.Println("Failed to init influx batch points: ", err.Error())
			os.Exit(1)
		}
		go logStatisticTask()
	}
}

func StatCheck(c *gin.Context) {
	if !conf.StatisticEnable {
		Error(c, NOT_PERMITTED, "statistic not enable on server side")
		c.Abort()
	}
}

func StatisticHandler(c *gin.Context) {
	now := time.Now()
	c.Next()
	clientData := getClientData(c)
	if clientData == nil {
		return
	}

	respTime := time.Now().Sub(now)
	tags := map[string]string{
		"node": conf.ClientAddr,
		"app":  clientData.AppKey,
	}

	fields := map[string]interface{}{
		"status":     getServiceStatus(c),
		"error_code": getServiceErrorCode(c),
		"ip":         clientData.Ip,
		"lang":       clientData.Lang,
		"os":         clientData.OSType,
		"osv":        clientData.OSVersion,
		"appv":       clientData.AppVersion,
		"deviceid":   clientData.DeviceId,
		"resp_time":  int(respTime / microSecondUnit),
	}

	p, _ := influx.NewPoint("client_request", tags, fields, now)
	select {
	case statisticCh <- p:
	default:
		log.Println("failed to send influx point to channel")
	}
}

func logStatisticTask() {
	exitChan := make(chan struct{})
	for {
		go func() {
			defer func() { exitChan <- struct{}{} }()
			for {
				logStatistic(<-statisticCh)
			}
		}()

		<-exitChan
	}
}

func logStatistic(p *influx.Point) {
	influxBatchPoints.AddPoint(p)
	if len(influxBatchPoints.Points()) < conf.InfluxBatchPointsCount {
		return
	}

	if err := influxClient.Write(influxBatchPoints); err != nil {
		log.Println("Failed to dump point to influxdb: ", err.Error())
	}
	influxBatchPoints, _ = influx.NewBatchPoints(
		influx.BatchPointsConfig{
			Database:  conf.InfluxDB,
			Precision: "s",
		})
}

func queryInflux(q string) (*influx.Response, error) {
	client, err := influx.NewHTTPClient(
		influx.HTTPConfig{
			Addr:     conf.InfluxURL,
			Username: conf.InfluxUser,
			Password: conf.InfluxPassword,
		})
	if err != nil {
		return nil, err
	}

	return client.Query(
		influx.Query{
			Command:   q,
			Database:  conf.InfluxDB,
			Precision: "s",
		})
}

func GetDeviceCountOfAppLatestConfig(c *gin.Context) {
	memConfMux.RLock()
	app := memConfApps[c.Param("app_key")]
	memConfMux.RUnlock()

	if app == nil {
		Error(c, BAD_REQUEST, "app not found for app key: "+c.Param("app_key"))
		return
	}

	q := fmt.Sprintf("SELECT COUNT(DISTINCT(deviceid)) FROM client_request where app = '%s' AND time >= %d", app.Key, app.LastUpdateUTC*int(secondUnit))
	resp, err := queryInflux(q)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}
	if resp.Error() != nil {
		Error(c, SERVER_ERROR, resp.Error().Error())
		return
	}

	if len(resp.Results[0].Series) == 0 {
		Success(c, 0)
		return
	}

	Success(c, resp.Results[0].Series[0].Values[0][1])
}

func GetAppConfigResponseData(c *gin.Context) {
	memConfMux.RLock()
	app := memConfApps[c.Param("app_key")]
	memConfMux.RUnlock()

	if app == nil {
		Error(c, BAD_REQUEST, "app not found for app key: "+c.Param("app_key"))
		return
	}

	startTime, err := strconv.Atoi(c.Query("start_time"))
	if err != nil {
		Error(c, BAD_REQUEST, "start_time not number")
		return
	}
	endTime, err := strconv.Atoi(c.Query("end_time"))
	if err != nil {
		Error(c, BAD_REQUEST, "end_time not number")
		return
	}
	unit := c.Query("unit")

	if (endTime-startTime)/(24*3600) > 30 {
		Error(c, BAD_REQUEST, "only max 30 days duration support")
		return
	}
	unitKind := unit[len(unit)-1]
	if unitKind != 'h' && unitKind != 'd' && unitKind != 'w' {
		Error(c, BAD_REQUEST, "only 'd' or 'h' or 'w' unit support")
		return
	}

	q := fmt.Sprintf(
		"SELECT COUNT(resp_time), MEAN(resp_time) FROM client_request where app = '%s' AND time >= %d AND time <= %d GROUP BY time(%s) fill(0)",
		app.Key, startTime*int(secondUnit), endTime*int(secondUnit), unit)
	resp, err := queryInflux(q)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}
	if resp.Error() != nil {
		Error(c, SERVER_ERROR, resp.Error().Error())
		return
	}

	res := make([][]interface{}, 0)
	if len(resp.Results[0].Series) > 0 {
		for _, val := range resp.Results[0].Series[0].Values {
			res = append(res, []interface{}{val[0], val[1], val[2]})
		}
	}

	Success(c, res)
}

func GetNodeConfigResponseData(c *gin.Context) {
	memConfMux.RLock()
	node := memConfNodes[c.Param("node_url")]
	memConfMux.RUnlock()

	if node == nil {
		Error(c, BAD_REQUEST, "node not found for node url "+c.Param("node_url"))
		return
	}

	startTime, err := strconv.Atoi(c.Query("start_time"))
	if err != nil {
		Error(c, BAD_REQUEST, "start_time not number")
		return
	}
	endTime, err := strconv.Atoi(c.Query("end_time"))
	if err != nil {
		Error(c, BAD_REQUEST, "end_time not number")
		return
	}
	unit := c.Query("unit")

	if (endTime-startTime)/(24*3600) > 30 {
		Error(c, BAD_REQUEST, "only max 30 days duration support")
		return
	}
	unitKind := unit[len(unit)-1]
	if unitKind != 'h' && unitKind != 'd' && unitKind != 'w' {
		Error(c, BAD_REQUEST, "only 'd' or 'h' or 'w' unit support")
		return
	}

	q := fmt.Sprintf(
		"SELECT COUNT(resp_time), MEAN(resp_time) FROM client_request where node = '%s' AND time >= %d AND time <= %d GROUP BY time(%s) fill(0)",
		node.URL, startTime*int(secondUnit), endTime*int(secondUnit), unit)
	resp, err := queryInflux(q)
	if err != nil {
		Error(c, SERVER_ERROR, err.Error())
		return
	}
	if resp.Error() != nil {
		Error(c, SERVER_ERROR, resp.Error().Error())
		return
	}

	res := make([][]interface{}, 0)
	if len(resp.Results[0].Series) > 0 {
		for _, val := range resp.Results[0].Series[0].Values {
			res = append(res, []interface{}{val[0], val[1], val[2]})
		}
	}

	Success(c, res)
}
