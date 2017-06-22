// delete-all-passes-for-templateId
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type requestResult struct {
	p []byte
	e error
}

// these structs were generated with https://github.com/mholt/json-to-go
type ListPasses struct {
	Pagination struct {
		Start     int    `json:"start"`
		PageSize  int    `json:"pageSize"`
		Page      int    `json:"page"`
		Direction string `json:"direction"`
		Order     string `json:"order"`
	} `json:"pagination"`
	Passes []struct {
		CreatedAt    time.Time     `json:"createdAt"`
		SerialNumber string        `json:"serialNumber"`
		UaEntityID   string        `json:"uaEntityId"`
		ExternalID   string        `json:"externalId"`
		ID           string        `json:"id"`
		TemplateID   string        `json:"templateId"`
		URL          string        `json:"url"`
		UpdatedAt    time.Time     `json:"updatedAt"`
		Tags         []interface{} `json:"tags"`
	} `json:"passes"`
	Count int `json:"count"`
}

type DeletePass struct {
	Status string `json:"Status"`
	PassID string `json:"PassID"`
}

type GetPass struct {
	SerialNumber string        `json:"serialNumber"`
	UaEntityID   string        `json:"uaEntityId"`
	ExternalID   string        `json:"externalId"`
	TemplateID   string        `json:"templateId"`
	URL          string        `json:"url"`
	Tags         []interface{} `json:"tags"`
	CreatedAt    time.Time     `json:"createdAt"`
	PublicURL    struct {
		Path  string `json:"path"`
		Image string `json:"image"`
		Type  string `json:"type"`
	} `json:"publicUrl"`
	Locations []struct {
		RelevantText string  `json:"relevantText"`
		Latitude     float64 `json:"latitude"`
		ID           int     `json:"id"`
		Longitude    float64 `json:"longitude"`
	} `json:"locations"`
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
	Status    string    `json:"status"`
}

func doRequest(done chan<- requestResult, client *http.Client, req *http.Request) {
	rs, err := client.Do(req)
	if err != nil {
		done <- requestResult{e: err}
		return
	}
	defer rs.Body.Close()

	p, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		done <- requestResult{e: err}
		return
	}
	done <- requestResult{p: p}
}

func PassProviderRequest(method string, servicePath string, authKey string) *http.Request {
	req, _ := http.NewRequest(method, "https://wallet-api.urbanairship.com/v1"+servicePath, nil)
	req.Header.Add("Api-Revision", "1.2")
	if authKey != "" {
		req.Header.Add("Authorization", "Basic "+authKey)
	}
	return req
}

func main() {
	var authKey, templateId string

	flag.StringVar(&authKey, "authKey", "", "[optional] The authKey is the email and api key combined into a string [YOUR_EMAIL]:[YOUR_KEY] and base64 encoded. If supplied as a command-line option, it is used to create the HTTP Authorization header for Basic Authentication (I personally use a local proxy to add the Auth header instead, refer to https://www.charlesproxy.com/documentation/tools/rewrite/)")
	flag.StringVar(&templateId, "templateId", "", "[required] The templateId which you want to delete all passes from")

	flag.Parse()

	req := PassProviderRequest("GET", "/template/"+templateId+"/passes?pageSize=1000", authKey) // The max pageSize allowed by UA is 1000
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	done := make(chan requestResult, 1)

	go doRequest(done, client, req)

	res := <-done
	if res.e != nil {
		panic(res.e)
	}

	listPasses := ListPasses{}
	json.Unmarshal(res.p, &listPasses)

	fmt.Println("Template has " + strconv.Itoa(listPasses.Count) + " pass(es) that will be deleted")

	deletions := make(chan requestResult)
	go func() {
		for _, pass := range listPasses.Passes {
			req := PassProviderRequest("DELETE", "/pass/"+pass.ID, authKey)
			go doRequest(deletions, client, req)
			time.Sleep(time.Millisecond * 100)
		}
	}()

	finish := make(chan bool, 1)
	go printResponses(finish, deletions, listPasses.Count)
	<-finish
}

func printResponses(finish chan<- bool, responses <-chan requestResult, count int) {
	for i := 0; i < count; i++ {
		res := <-responses
		if res.e != nil {
			fmt.Println(res.e)
		} else {
			deletePass := GetPass{}
			json.Unmarshal(res.p, &deletePass)
			fmt.Println("Deleted PassId: " + deletePass.ID)
		}
	}
	finish <- true
}
