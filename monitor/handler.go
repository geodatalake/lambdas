package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/geodatalake/lambdas/elastichelper"
	"github.com/geodatalake/lambdas/geoindex"
	"github.com/go-redis/redis"
)

var version = "0.24"

func main() {}

type RequestParams struct {
	Stats  string
	client *redis.Client
}

func loadFromEnv(name, deflt string) string {
	if retval, ok := os.LookupEnv(name); !ok {
		return deflt
	} else {
		return retval
	}
}

func NewParams() *RequestParams {
	params := new(RequestParams)
	params.Stats = loadFromEnv("stats", "")
	return params
}

func (rp *RequestParams) String() string {
	return fmt.Sprintf("{Stats: %s}", rp.Stats)
}

func doc() *elastichelper.Document {
	return elastichelper.NewDoc()
}

func array() *elastichelper.ArrayBuilder {
	return elastichelper.StartArray()
}

func Handle(evt interface{}, ctx *runtime.Context) (interface{}, error) {
	params := NewParams()
	log.Println("Version", version)
	log.Println("Params:", params)
	var client *redis.Client
	if params.Stats != "" {
		client = redis.NewClient(&redis.Options{
			Addr:     params.Stats,
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		defer client.Close()
	} else {
		log.Println("No stats environment variable found")
		return "", fmt.Errorf("No stats environment variable found")
	}
	if m, ok := evt.(map[string]interface{}); ok {
		if m["httpMethod"] == "GET" {
			log.Println("Answering GET request")
			sb := client.Get(geoindex.JScan)
			gf := client.Get(geoindex.JGroup)
			pr := client.Get(geoindex.JProcess)
			sb_totalrun := client.Get(geoindex.JScan + geoindex.Totalrun)
			gf_totalrun := client.Get(geoindex.JGroup + geoindex.Totalrun)
			pr_totalrun := client.Get(geoindex.JProcess + geoindex.Totalrun)
			body := doc().
				AddKV("Version", version).
				AddKV("ScanBucket", sb.Val()).
				AddKV("GroupFiles", gf.Val()).
				AddKV("ProcessGeo", pr.Val()).
				AddKV("ScanBucket_Totalrun", sb_totalrun.Val()).
				AddKV("GroupFiles_Totalrun", gf_totalrun.Val()).
				AddKV("ProcessGeo_Totalrun", pr_totalrun.Val())
			llen := client.LLen("CompletedJobs")
			if llen.Val() > 0 {
				jobs := client.LRange("CompletedJobs", 0, 5)
				arr := array()
				for _, v := range jobs.Val() {
					dur := client.Get(v)
					arr.Add(doc().AddKV("Name", v).AddKV("Duration", dur.Val()))
				}
				body.AppendArray("Last_5_Jobs", arr)
			}
			b, _ := json.Marshal(body.Build())
			resp := doc().
				AddKV("statusCode", "200").
				AddKV("body", string(b))
			return resp.Build(), nil
		} else {
			req := new(geoindex.IndexerRequest)
			if err := req.Unmarshal(m); err != nil {
				log.Println("Error unmarshalling request", err)
				return nil, err
			} else {
				log.Println("Processing", req)
			}
			switch req.RequestType {
			case geoindex.Enter:
				client.IncrBy(req.Name, int64(req.Num))
				return geoindex.NewIndexerResponse(true, 0), nil
			case geoindex.Leave:
				client.DecrBy(req.Name, int64(req.Num))
				client.IncrBy(req.Name+geoindex.Totalrun, int64(req.Num))
				return geoindex.NewIndexerResponse(true, 0), nil
			case geoindex.Reset:
				r := client.Set(geoindex.JScan, "0", 0)
				r1 := client.Set(geoindex.JGroup, "0", 0)
				r2 := client.Set(geoindex.JProcess, "0", 0)
				r_totals := client.Set(geoindex.JScan+geoindex.Totalrun, "0", 0)
				r1_totals := client.Set(geoindex.JGroup+geoindex.Totalrun, "0", 0)
				r2_totals := client.Set(geoindex.JProcess+geoindex.Totalrun, "0", 0)
				cj := client.Del("CompletedJobs")
				log.Println("initialized values:", r, r1, r2, r_totals, r1_totals, r2_totals, cj)
				return geoindex.NewIndexerResponse(true, 0), nil
			case geoindex.JobComplete:
				client.Set(req.Name, req.Duration.String(), 0)
				client.LPush("CompletedJobs", req.Name)
				client.LTrim("CompletedJobs", 0, 5)
				return geoindex.NewIndexerResponse(true, 0), nil
			}
			log.Println("Unhandled request")
			return geoindex.NewIndexerResponse(false, 0), nil
		}
	} else {
		log.Println("evt is not of type map[string]interface{}, it is", reflect.TypeOf(evt).String())
		return nil, fmt.Errorf("evt is not of type map[string]interface{}, it is %s", reflect.TypeOf(evt).String())
	}
}
