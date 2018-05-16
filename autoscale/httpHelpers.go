package autoscale

import (
	"bytes"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"log"
)

type Oath2 struct {
	User        string
	Password    string
	AccessToken string
}

func (o *Oath2) GetSetAccessToken() (string, error) {
	client := http.DefaultClient
	body := bytes.NewBufferString("client_id=" + o.User + "&" + "client_secret=" + o.Password + "&" + "grant_type=client_credentials")
	req, err := http.NewRequest("POST", "https://auth.metapipe.uit.no/oauth2/token", body)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	type authResp struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}
	var auth authResp
	b, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(b, &auth)
	if err != nil {
		return "", err
	}

	o.AccessToken = auth.AccessToken

	return auth.AccessToken, nil
}

type RetryClient struct {
	Auth        Oath2
	MaxAttempts int
	Client      http.Client
}

func (rc *RetryClient) retryGet(url string) (*http.Response, error) {
	resp, err := rc.Client.Get(url)
	if err != nil {
		for i := 0; i < rc.MaxAttempts; i++ {
			retryResp, err := rc.Client.Get(url)
			if err == nil {
				resp = retryResp
				break
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func (rc *RetryClient) GetMetapipeJobSize(jobId string) (int64, error) {
	resp, err := rc.retryGet("https://jobs.metapipe.uit.no/jobs/" + jobId)
	defer resp.Body.Close()
	if err != nil {
		return 0, err
	}
	var mJob MetapipeJob
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&mJob)
	if err != nil {
		return 0, err
	}
	s, err := rc.fetchSize(mJob, rc.Auth.AccessToken)

	return s, nil
}

func (rc *RetryClient) fetchSize(mJob MetapipeJob, authToken string) (int64, error) {
	if mJob.Inputs.InputFas.Url == "" {
		return 0, nil
	}
	baseSizeRequest, err := http.NewRequest("HEAD", mJob.Inputs.InputFas.Url, bytes.NewBufferString(""))
	if err != nil {
		return 0, err
	}
	baseSizeRequest.Header.Add("Authorization", "Bearer "+authToken)
	resp, err := rc.Client.Do(baseSizeRequest)
	defer resp.Body.Close()
	if err != nil {
		if err == http.ErrHandlerTimeout {
			return 0, nil
		}
		return 0, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return 0, nil
	}
	return resp.ContentLength, nil
}

func (rc *RetryClient) GetAllMetapipeJobs() ([]MetapipeJob, error) {
	log.Printf("Downloading metapipe dataset")
	resp, err := rc.retryGet("https://jobs.metapipe.uit.no/jobs")
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	log.Printf("Done downloading dataset")

	var all []MetapipeJob
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&all)
	if err != nil {
		return nil, err
	}
	return all, nil
}
