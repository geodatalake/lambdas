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
	Init   string
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
	params.Init = loadFromEnv("init", "")
	return params
}

func (rp *RequestParams) String() string {
	return fmt.Sprintf("{Init: %s, Stats: %s}", rp.Init, rp.Stats)
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
	if params.Init == "true" {
		r := client.Set(geoindex.JScan, "0", 0)
		r1 := client.Set(geoindex.JGroup, "0", 0)
		r2 := client.Set(geoindex.JProcess, "0", 0)
		r_contracts := client.Set(geoindex.JScan+"_contracts", "10", 0)
		r1_contracts := client.Set(geoindex.JGroup+"_contracts", "20", 0)
		r2_contracts := client.Set(geoindex.JProcess+"_contracts", "200", 0)
		log.Println("initialized values:", r, r1, r2, r_contracts, r1_contracts, r2_contracts)
	}
	if m, ok := evt.(map[string]interface{}); ok {
		if m["httpMethod"] == "GET" {
			sb := client.Get(geoindex.JScan)
			gf := client.Get(geoindex.JGroup)
			pr := client.Get(geoindex.JProcess)
			sb_contracts := client.Get(geoindex.JScan + "_contracts")
			gf_contracts := client.Get(geoindex.JGroup + "_contracts")
			pr_contracts := client.Get(geoindex.JProcess + "_contracts")

			body := doc().
				AddKV("ScanBucket", fmt.Sprintf("%s of %d", sb.Val(), 10)).
				AddKV("GroupFiles", fmt.Sprintf("%s of %d", gf.Val(), 20)).
				AddKV("ProcessGeo", fmt.Sprintf("%s of %d", pr.Val(), 200)).
				AddKV("ScanBucket_Contracts", fmt.Sprintf("%s", sb_contracts.Val())).
				AddKV("GroupFiles_Contracts", fmt.Sprintf("%s", gf_contracts.Val())).
				AddKV("ProcessGeo_Contracts", fmt.Sprintf("%s", pr_contracts.Val())).Build()
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
				return geoindex.NewIndexerResponse(true), nil
			case geoindex.Leave:
				contract.Leave(req.Name)
				return geoindex.NewIndexerResponse(true), nil
			case geoindex.Reserve:
				return geoindex.NewIndexerResponse(contract.Reserve(req.Name)), nil
			case geoindex.Release:
				contract.Release(req.Name)
				return geoindex.NewIndexerResponse(true), nil
			}
			return geoindex.NewIndexerResponse(false), nil
		}
	} else {
		log.Println("evt is not of type map[string]interface{}, it is", reflect.TypeOf(evt).String())
		return nil, fmt.Errorf("evt is not of type map[string]interface{}, it is %s", reflect.TypeOf(evt).String())
	}
}
