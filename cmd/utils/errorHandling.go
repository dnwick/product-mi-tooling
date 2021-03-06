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
	"fmt"
	"os"
)

var IsVerbose bool

func HandleErrorAndExit(msg string, err error) {
	if err == nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", ProjectName, msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s: %v Reason: %v\n", ProjectName, msg, err.Error())
	}
	defer printAndExit()
}

func printAndExit() {

	if !IsVerbose {
		fmt.Println("Execute with --verbose to see detailed info.")
	}
	os.Exit(1)
}
