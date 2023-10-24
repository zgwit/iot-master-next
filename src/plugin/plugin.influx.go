package zgwit_plugin

import (
	"context"
	"fmt"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type Influx struct {
	Client influxdb2.Client
	Config InfluxConfig
}

type InfluxConfig struct {
	Url   string `form:"url" bson:"url" json:"url"`
	Token string `form:"token" bson:"token" json:"token"`
	Org   string `form:"org" bson:"org" json:"org"`
}

func NewInflux(config InfluxConfig) (influx Influx) {

	influx.Config = config

	influx.Client = influxdb2.NewClient(influx.Config.Url, influx.Config.Token)

	return
}

func (influx *Influx) Ping() (result bool, err error) {

	result, err = influx.Client.Ping(context.Background())

	return
}

func (influx *Influx) DeleteByMeasurement(bucket string, measurement string) (err error) {

	domain_org, err := influx.Client.OrganizationsAPI().FindOrganizationByName(context.Background(), influx.Config.Org)
	if err != nil {
		panic(err)
	}

	domain_bucket, err := influx.Client.BucketsAPI().FindBucketByName(context.Background(), bucket)
	if err != nil {
		panic(err)
	}

	return influx.Client.DeleteAPI().Delete(
		context.Background(),
		domain_org,
		domain_bucket,
		time.Unix(0, 0),
		time.Now(),
		fmt.Sprintf(`_measurement="%s"`, measurement),
	)
}

func (influx *Influx) DeleteByTag(bucket string, measurement string, tag_key, tag_value string) (err error) {

	domain_org, err := influx.Client.OrganizationsAPI().FindOrganizationByName(context.Background(), influx.Config.Org)
	if err != nil {
		panic(err)
	}

	domain_bucket, err := influx.Client.BucketsAPI().FindBucketByName(context.Background(), bucket)
	if err != nil {
		panic(err)
	}

	return influx.Client.DeleteAPI().Delete(
		context.Background(),
		domain_org,
		domain_bucket,
		time.Unix(0, 0),
		time.Now(),
		fmt.Sprintf(`_measurement="%s" and "%s" = "%s"`, measurement, tag_key, tag_value),
	)
}

func (influx *Influx) Query(flux_code string) (query_results []map[string]interface{}, err error) {

	var (
		results *api.QueryTableResult
	)

	query_results = []map[string]interface{}{}

	if results, err = influx.Client.QueryAPI(influx.Config.Org).Query(context.Background(), flux_code); err != nil {
		goto FAIL
	}

	for results.Next() {
		record := results.Record()
		values := record.Values()
		values["time"] = record.Time().Local().Unix()
		delete(values, "result")
		delete(values, "table")
		delete(values, "_time")
		query_results = append(query_results, values)
	}

	err = results.Err()

FAIL:
	if err != nil && strings.Contains(err.Error(), "dial tcp") {
		err = nil
	}

	return
}

type InfluxBatch struct {
	WriteApi api.WriteAPIBlocking
	Points   []*write.Point
}

func NewInfluxBatch(influx *Influx, bucket string) (batch InfluxBatch) {

	batch.WriteApi = influx.Client.WriteAPIBlocking(influx.Config.Org, bucket)
	batch.Points = []*write.Point{}

	return
}

func (batch *InfluxBatch) AddPoint(measurement string, tags map[string]string, fields map[string]interface{}, tim int64) {

	if len(fields) == 0 {
		return
	}

	batch.Points = append(batch.Points, write.NewPoint(measurement, tags, fields, time.Unix(tim, 0)))
}

func (batch *InfluxBatch) Write() (err error) {

	if len(batch.Points) == 0 {
		return
	}

	err = batch.WriteApi.WritePoint(context.Background(), batch.Points...)

	if err != nil && strings.Contains(err.Error(), "dial tcp") {
		err = nil
	}

	return
}
