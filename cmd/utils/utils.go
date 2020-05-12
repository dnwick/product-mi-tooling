/*
* Copyright (c) 2019, WSO2 Inc. (http://www.wso2.org) All Rights Reserved.
*
* WSO2 Inc. licenses this file to you under the Apache License,
* Version 2.0 (the "License"); you may not use this file except
* in compliance with the License.
* You may obtain a copy of the License at
*
*    http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing,
* software distributed under the License is distributed on an
* "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
* KIND, either express or implied. See the License for the
* specific language governing permissions and limitations
* under the License.
 */

package utils

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh/terminal"
	"github.com/magiconair/properties"
	"gopkg.in/resty.v1"
	"path"
)

// Invoke http-post request using go-resty
func InvokePOSTRequest(url string, headers map[string]string, body map[string]string) (*resty.Response, error) {

    if headers == nil {
        headers =  make(map[string]string)
    }

    if headers[HeaderAuthorization] == "" {
        headers[HeaderAuthorization] = HeaderValueAuthPrefixBearer + " " +
        RemoteConfigData.Remotes[RemoteConfigData.CurrentRemote].AccessToken
    }

	AllowInsecureSSLConnection()
	resp, err := resty.R().SetHeaders(headers).SetBody(body).Post(url)

	return resp, err
}

// Invoke http-get request using go-resty
func InvokeGETRequest(url string, headers map[string]string, params map[string]string) (*resty.Response, error) {

	AllowInsecureSSLConnection()
	Logln(LogPrefixInfo + "InvokeGETRequest(): URL: " + url)
	resp, err := resty.R().SetQueryParams(params).SetHeaders(headers).Get(url)

	return resp, err
}

// Invoke http-put request using go-resty
func InvokeUPDATERequest(url string, headers map[string]string, body map[string]string) (*resty.Response, error) {

	AllowInsecureSSLConnection()
	resp, err := resty.R().SetHeaders(headers).SetBody(body).Patch(url)

	return resp, err
}

// Invoke http-delete request using go-resty
func InvokeDELETERequest(url string, headers map[string]string) (*resty.Response, error) {

    if headers == nil {
	    headers =  make(map[string]string)
    }

    if headers[HeaderAuthorization] == "" {
	    headers[HeaderAuthorization] = HeaderValueAuthPrefixBearer + " " +
	    RemoteConfigData.Remotes[RemoteConfigData.CurrentRemote].AccessToken
    }

	AllowInsecureSSLConnection()
	resp, err := resty.R().SetHeaders(headers).Delete(url)

	return resp, err
}

func PromptForUsername() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')

	return strings.TrimSpace(username)
}

func PromptForPassword() string {
	fmt.Print("Enter Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)
	fmt.Println()
	return strings.TrimSpace(password)
}

// return a string containing the file name, function name
// and the line number of a specified entry on the call stack
func WhereAmI(depthList ...int) string {
	var depth int
	if depthList == nil {
		depth = 1
	} else {
		depth = depthList[0]
	}
	function, file, line, _ := runtime.Caller(depth)
	return fmt.Sprintf("File: %s Line: %d Function: %s ", chopPath(file), line, runtime.FuncForPC(function).Name())
}

// return the source filename after the last slash
func chopPath(original string) string {
	i := strings.LastIndex(original, "/")
	if i == -1 {
		return original
	} else {
		return original[i+1:]
	}
}

func PrintList(list []string) {
	for _, item := range list {
		fmt.Println(item)
	}
}

func AllowInsecureSSLConnection() {
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
}

// Unmarshal Data from the response to the respective struct
// @param url: url of rest api
// @param headers: HTTP headers
// @param model: struct object
// @param params: parameters for the HTTP call
// @return struct object
// @return error
func UnmarshalData(url string, headers map[string]string, params map[string]string,
	model interface{}) (interface{}, error) {

	if headers == nil {
		headers = make(map[string]string)
	}

	if headers[HeaderAuthorization] == "" {
		headers[HeaderAuthorization] = HeaderValueAuthPrefixBearer + " " +
			RemoteConfigData.Remotes[RemoteConfigData.CurrentRemote].AccessToken
	}

	resp, err := InvokeGETRequest(url, headers, params)

	if err != nil {
		HandleErrorAndExit("Unable to connect to "+url, nil)
	}

	Logln(LogPrefixInfo+"Response:", resp.Status())

	if resp.StatusCode() == http.StatusOK {
		response := model
		unmarshalError := json.Unmarshal(resp.Body(), &response)

		if unmarshalError != nil {
			HandleErrorAndExit(LogPrefixError+"invalid JSON response", unmarshalError)
		}
		return response, nil
	} else {
		if resp.StatusCode() == http.StatusUnauthorized {
			// not logged in to MI
			fmt.Println("User not logged in or session timed out. Please login to the current Micro Integrator instance")
			fmt.Println("Execute '" + ProjectName + " remote login --help' for more information")
		}
		if len(resp.Body()) == 0 {
			return nil, errors.New(resp.Status())
		} else {
			data := UnmarshalJsonToStringMap(resp.Body())
			return data["Error"], errors.New(resp.Status())
		}
	}
}

func UnmarshalLogFileData(url string, headers map[string]string, params map[string]string, filename string) {
    if headers == nil {
        headers = make(map[string]string)
    }

    if headers[HeaderAuthorization] == "" {
        headers[HeaderAuthorization] = HeaderValueAuthPrefixBearer + " " +
            RemoteConfigData.Remotes[RemoteConfigData.CurrentRemote].AccessToken
    }

    resp, err := InvokeGETRequest(url, headers, params)

    if err != nil {
        HandleErrorAndExit("Unable to connect to "+url, nil)
    }

    ioutil.WriteFile(filename, resp.Body(), 0644)
    Logln(LogPrefixInfo+"Response:", resp.Status())
}

func UpdateMILogger(loggerName, loggingLevel string) (interface{}, error) {

	url := GetRESTAPIBase() + PrefixLogging
	Logln(LogPrefixInfo+"URL:", url)
	headers := make(map[string]string)
	body := make(map[string]string)
	body["loggerName"] = loggerName
	body["loggingLevel"] = loggingLevel

	if headers[HeaderAuthorization] == "" {
		headers[HeaderAuthorization] = HeaderValueAuthPrefixBearer + " " +
			RemoteConfigData.Remotes[RemoteConfigData.CurrentRemote].AccessToken
	}

	resp, err := InvokeUPDATERequest(url, headers, body)

	if err != nil {
		HandleErrorAndExit("Unable to connect to " + url, err)
	}

	Logln(LogPrefixInfo+"Response:", resp.Status())

	if resp.StatusCode() == http.StatusUnauthorized {
		// not logged in to MI
		fmt.Println("User not logged in or session timed out. Please login to the current Micro Integrator instance")
		fmt.Println("Execute '" + ProjectName + " remote login --help' for more information")
	}
	if len(resp.Body()) == 0 {
		return nil, errors.New(resp.Status())
	} else {
		data := UnmarshalJsonToStringMap(resp.Body())
		if resp.StatusCode() == http.StatusOK {
			return data["message"], nil
		} else {
			return nil, errors.New(data["Error"])
		}
	}
}

func GetUrlAndParams(urlPrefix, key, value string) (string, map[string]string) {
	url := GetRESTAPIBase() + urlPrefix
	params := make(map[string]string)
	params[key] = value
	return url, params
}

func PutQueryParamsToMap(paramMap map[string]string, key string, value string) map[string]string {
	paramMap[key] = value
	return paramMap
}

func GetCmdFlags(cmd string) string {
	var showCmdFlags = "Flags:\n" +
		"  -h, --help\t\thelp for " + cmd + "\n" +
		"Global Flags:\n" +
		"  -v, --verbose\t\tEnable verbose mode\n"
	return showCmdFlags
}

func GetCmdUsage(program, cmd, subcmd, arg string) string {
	var showCmdUsage = "Usage:\n" +
		"  " + program + " " + cmd + " " + subcmd + "\n" +
		"  " + program + " " + cmd + " " + subcmd + " " + arg + "\n\n"
	return showCmdUsage
}

func GetCmdUsageMultipleArgs(program, cmd, subcmd string, args []string) string {
    var showCmdUsage = "Usage:\n" +
	    "  " + program + " " + cmd + " " + subcmd + "\n"
    for _, arg := range args {
	    showCmdUsage += "  " + program + " " + cmd + " " + subcmd + " " + arg + "\n"
    }
    return showCmdUsage
}

func GetCmdUsageForNonArguments(program, cmd, subcmd string) string {
	var showCmdUsage = "Usage:\n" +
		"  " + program + " " + cmd + " " + subcmd + "\n\n"
	return showCmdUsage
}

func InitRemoteConfigData() {

	filePath := GetRemoteConfigFilePath()
	if IsFileExist(filePath) {
		RemoteConfigData.Load(filePath)
	} else {
		Logln(LogPrefixWarning + "RemoteConfig: file not found at: " + filePath +
			" Adding the default config file.")
		RemoteConfigData.Reset()
		_ = RemoteConfigData.AddRemote(DefaultRemoteName, DefaultHost, DefaultPort)
		_ = RemoteConfigData.SelectRemote(DefaultRemoteName)
		RemoteConfigData.Persist(filePath)
	}
}

func GetRESTAPIBase() string {

	var restAPIBase string
	if RemoteConfigData.CurrentRemote != "" {
		restAPIBase = HTTPSProtocol + RemoteConfigData.Remotes[RemoteConfigData.CurrentRemote].Url + ":" +
			RemoteConfigData.Remotes[RemoteConfigData.CurrentRemote].Port + "/" + Context + "/"
	} else {
		// this cannot happen usually
		errMessage := `micro integrator is not specified. Please run "` + ProjectName + ` remote" command`
		HandleErrorAndExit(LogPrefixError, errors.New(errMessage))
	}

	return restAPIBase
}

func UnmarshalJsonToStringMap(body []byte) map[string]string {
	var data map[string]string
	unmarshalError := json.Unmarshal(body, &data)
	if unmarshalError != nil {
		HandleErrorAndExit(LogPrefixError+"invalid JSON response", unmarshalError)
	}
	return data
}

func GetTableWriter() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetColumnSeparator(" ")
	return table
}

func printTable(columnData []string, dataChannel <-chan []string) {
	table := GetTableWriter()

	table.Append(columnData)

	for v := range dataChannel {
		table.Append(v)
	}
	table.Render()
}

func PrintItemList(itemList IterableStringArray, columnData []string, emptyWarning string) {
	if itemList.GetCount() > 0 {
		printTable(columnData, itemList.GetDataIterator())
	} else {
		fmt.Println(emptyWarning)
	}
}

func CreateKeyValuePairs(mapData map[string]string) string {
	if len(mapData) > 0 {
		builder := new(bytes.Buffer)
		_, _ = fmt.Fprintf(builder, " {\n")
		for key, value := range mapData {
			_, _ = fmt.Fprintf(builder, "\t\t  %s = \"%s\"\n", key, value)
		}
		_, _ = fmt.Fprintf(builder, " \t\t}")
		return builder.String()
	} else {
		return "{}"
	}
}

func UpdateMIMessageProcessor(messageProcessorName, messageProcessorState string) (interface{}, error) {

	url := GetRESTAPIBase() + PrefixMessageProcessors
	Logln(LogPrefixInfo+"URL:", url)
	headers := make(map[string]string)
	body := make(map[string]string)
	body["name"] = messageProcessorName
	body["status"] = messageProcessorState

	if headers[HeaderAuthorization] == "" {
		headers[HeaderAuthorization] = HeaderValueAuthPrefixBearer + " " +
			RemoteConfigData.Remotes[RemoteConfigData.CurrentRemote].AccessToken
	}

	resp, err := InvokePOSTRequest(url, headers, body)

	if err != nil {
		HandleErrorAndExit("Unable to connect to " + url, err)
	}

	Logln(LogPrefixInfo + "Response:", resp.Status())

	if resp.StatusCode() == http.StatusUnauthorized {
		// not logged in to MI
		fmt.Println("User not logged in or session timed out. Please login to the current Micro Integrator instance")
		fmt.Println("Execute '" + ProjectName + " remote login --help' for more information")
	}

	if len(resp.Body()) == 0 {
		return nil, errors.New(resp.Status())
	} else {
		data := UnmarshalJsonToStringMap(resp.Body())
		if data["Message"] != "" {
			return data["Message"], nil
		} else {
			return nil, errors.New(data["Error"])
		}
	}
}

func IsValidConsoleInput(inputs map[string]string) (bool) {
	for key, input := range inputs {
		if len(strings.TrimSpace(input)) == 0 {
			fmt.Println("Invalid input for " + key)
			return false
		}
	}
	return true
}

func SetProperties(variables map[string]string, fileName string)  {
	props := properties.LoadMap(variables)
	writer, _ := os.Create(fileName)
	props.Write(writer, properties.UTF8)
	writer.Close()
}

func GetSecurityDirectoryPath() string {
	workingDirectory, _ := os.Getwd()
	return path.Join(workingDirectory, "security")
}

func GetkeyStoreInfoFileLocation() string {
	return path.Join(GetSecurityDirectoryPath(),  "keystore-info.properties")
}
