package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

func setupHeaders(req *http.Request) {
	req.Header.Add("User-Agent", "RssScraper 0.1")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("Accept-Charset", "utf-8")
	// If we request gzip, we have to manually gunzip it.
	// req.Header.Add("Accept-Encoding", "gzip")
}

func Post(client *http.Client, route string, data []byte, headers map[string]string) ([]byte, error) {
	if client == nil {
		return []byte{}, errors.New("null transport client")
	}
	buff := bytes.NewBuffer(data)

	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "POST", route, buff)
	setupHeaders(req)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return body, errors.New(strconv.Itoa(res.StatusCode))
	}
	return body, err
}

func Put(client *http.Client, route string, objData interface{}) ([]byte, error) {
	if client == nil {
		return []byte{}, errors.New("null transport client")
	}
	data, err := json.Marshal(objData)
	if err != nil {
		return nil, err
	}
	buff := bytes.NewBuffer(data)

	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "PUT", route, buff)
	setupHeaders(req)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return body, errors.New(strconv.Itoa(res.StatusCode))
	}
	return body, err
}

func Patch(client *http.Client, route string, objData interface{}) ([]byte, error) {
	if client == nil {
		return []byte{}, errors.New("null transport client")
	}
	data, err := json.Marshal(objData)
	if err != nil {
		return nil, err
	}
	buff := bytes.NewBuffer(data)
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "PATCH", route, buff)
	setupHeaders(req)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return body, errors.New(strconv.Itoa(res.StatusCode))
	}
	return body, err
}

func Get(client *http.Client, route string, headers map[string]string) ([]byte, error) {
	if client == nil {
		return []byte{}, errors.New("null transport client")
	}
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "GET", route, nil)
	setupHeaders(req)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return body, errors.New(strconv.Itoa(res.StatusCode))
	}
	return body, err
}

func GetWithHeaders(client *http.Client, url string, headers map[string]string) ([]byte, error) {
	if client == nil {
		return []byte{}, errors.New("null transport client")
	}
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	// setupHeaders(req)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return body, errors.New(strconv.Itoa(res.StatusCode))
	}
	return body, err
}

func Delete(client *http.Client, route string) ([]byte, error) {
	if client == nil {
		return []byte{}, errors.New("null transport client")
	}
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "DELETE", route, nil)
	setupHeaders(req)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		return body, errors.New(strconv.Itoa(res.StatusCode))
	}
	return body, err
}
