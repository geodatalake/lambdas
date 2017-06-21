package main

import (
	"encoding/json"
	"reflect"
	//	"encoding/json"
	"fmt"
	"log"
	"os"
	//	"strconv"
	//	"strings"

	"github.com/eawsy/aws-lambda-go-core/service/lambda/runtime"
	"github.com/geodatalake/lambdas/elastichelper"
	"github.com/geodatalake/lambdas/geoindex"
	"github.com/go-redis/redis"
)

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
		return "", fmt.Errorf("No stats environment variable found")
	}
	if m, ok := evt.(map[string]interface{}); ok {
		if m["httpMethod"] == "GET" {
			sb := client.Get(geoindex.JScan)
			gf := client.Get(geoindex.JGroup)
			pr := client.Get(geoindex.JProcess)
			sb_contracts := client.Get(geoindex.JScan + geoindex.Contracts)
			gf_contracts := client.Get(geoindex.JGroup + geoindex.Contracts)
			pr_contracts := client.Get(geoindex.JProcess + geoindex.Contracts)
			sb_totalrun := client.Get(geoindex.JScan + geoindex.Totalrun)
			gf_totalrun := client.Get(geoindex.JGroup + geoindex.Totalrun)
			pr_totalrun := client.Get(geoindex.JProcess + geoindex.Totalrun)
			body := doc().
				AddKV("pr_contracts", pr_contracts.String()).AddKV("sb_contracts", sb_contracts.String()).AddKV("gf_contracts", gf_contracts.String()).
				AddKV("ScanBucket", fmt.Sprintf("%s of %d", sb.Val(), 10)).
				AddKV("GroupFiles", fmt.Sprintf("%s of %d", gf.Val(), 50)).
				AddKV("ProcessGeo", fmt.Sprintf("%s of %d", pr.Val(), 100)).
				AddKV("ScanBucket_Contracts", fmt.Sprintf("%s", sb_contracts.Val())).
				AddKV("GroupFiles_Contracts", fmt.Sprintf("%s", gf_contracts.Val())).
				AddKV("ProcessGeo_Contracts", fmt.Sprintf("%s", pr_contracts.Val())).
				AddKV("ScanBucket_Totalrun", fmt.Sprintf("%s", sb_totalrun.Val())).
				AddKV("GroupFiles_Totalrun", fmt.Sprintf("%s", gf_totalrun.Val())).
				AddKV("ProcessGeo_Totalrun", fmt.Sprintf("%s", pr_totalrun.Val())).Build()
			b, _ := json.Marshal(body)
			resp := doc().
				AddKV("statusCode", "200").
				AddKV("body", string(b))
			return resp.Build(), nil
		} else {
			req := new(geoindex.IndexerRequest)
			if err := req.Unmarshal(m); err != nil {
				log.Println("Error unmarshalling request", err)
				return nil, err
			}
			contract := geoindex.NewContractTracker(client, nil, "")
			switch req.RequestType {
			case geoindex.Enter:
				contract.Enter(req.Name)
				return geoindex.NewIndexerResponse(true, 0), nil
			case geoindex.Leave:
				contract.Leave(req.Name)
				return geoindex.NewIndexerResponse(true, 0), nil
			case geoindex.Reserve:
				success := contract.ReserveMany(req.Name, req.Num)
				return geoindex.NewIndexerResponse(success > 0, success), nil
			case geoindex.Release:
				if req.Name == geoindex.JProcess {
					log.Println("Release(", req.Num, ")")
					mySqs := geoindex.NewSqsInstance().
						WithQueue("https://sqs.us-west-2.amazonaws.com/414519249282/process-geo-test-queue").
						WithRegion("us-west-2")
					if _, err := mySqs.Send(fmt.Sprintf("Release(%d)", req.Num)); err != nil {
						log.Println("Error sending to SQS", err)
					}
				}
				for i := req.Num; i > 0; i-- {
					contract.Release(req.Name)
					client.Incr(req.Name + geoindex.Totalrun)
				}
				return geoindex.NewIndexerResponse(true, req.Num), nil
			case geoindex.Reset:
				r := client.Set(geoindex.JScan, "-1", 0)
				r1 := client.Set(geoindex.JGroup, "-1", 0)
				r2 := client.Set(geoindex.JProcess, "-1", 0)
				r_contracts := client.Set(geoindex.JScan+geoindex.Contracts, "10", 0)
				r1_contracts := client.Set(geoindex.JGroup+geoindex.Contracts, "50", 0)
				r2_contracts := client.Set(geoindex.JProcess+geoindex.Contracts, "100", 0)
				r_totals := client.Set(geoindex.JScan+geoindex.Totalrun, "0", 0)
				r1_totals := client.Set(geoindex.JGroup+geoindex.Totalrun, "0", 0)
				r2_totals := client.Set(geoindex.JProcess+geoindex.Totalrun, "0", 0)
				log.Println("initialized values:", r, r1, r2, r_contracts, r1_contracts, r2_contracts, r_totals, r1_totals, r2_totals)
				return geoindex.NewIndexerResponse(true, 0), nil
			}
			return geoindex.NewIndexerResponse(false, 0), nil
		}
	} else {
		log.Println("evt is not of type map[string]interface{}, it is", reflect.TypeOf(evt).String())
		return nil, fmt.Errorf("evt is not of type map[string]interface{}, it is %s", reflect.TypeOf(evt).String())
	}
}
