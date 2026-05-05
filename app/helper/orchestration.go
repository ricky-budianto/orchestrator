package helper

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	jsonExt "github.com/json-iterator/go"
	"github.com/kevinburke/nacl"
	goredis "github.com/redis/go-redis/v9"
	"github.com/soluixdeveloper/ces-orchestratorService/app/model"
	orchestrationConfig "github.com/soluixdeveloper/ces-orchestratorService/config/orchestration"
	"github.com/soluixdeveloper/ces-orchestratorService/config/redis"
	"github.com/soluixdeveloper/ces-orchestratorService/config/utils"
	ceslogger "github.com/soluixdeveloper/ces-utilities/v2/ceslogger"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	redisModuleReferenceId = "reference_id"
)

func CollectData(address interface{}, dataCenter map[string]interface{}, logging ceslogger.Logger) (res model.ResponseData) {
	res.Message = "Success"
	res.ResponseCode = "S1"
	res.Data = ""
	var dataRequest interface{}
	logging.LogDebug("Address ", address)
	// outBreak:
	switch reflect.TypeOf(address).Kind() {
	case reflect.String:
		if strings.Contains(address.(string), "string(") {
			dataRequest = strings.Trim(address.(string), "string()")
		} else if !strings.Contains(address.(string), "${{") {
			dataRequest = address.(string)
		} else if strings.Contains(address.(string), "${{ibridge.Helpers") {
			dataRequest = ibridgeHelpers(address.(string), dataCenter, logging)
		} else {
			trimed := strings.Trim(address.(string), "${}")
			re := regexp.MustCompile(`<([^>]*)>`)
			matches := re.FindAllStringSubmatchIndex(trimed, -1)

			if matches != nil {
				replacedStr := trimed
				// Loop through matches in reverse order to handle index changes
				for i := len(matches) - 1; i >= 0; i-- {
					match := matches[i]
					start := match[2] // Start index of the string between <>
					end := match[3]   // End index of the string between <>
					substr := "${{" + trimed[start:end] + "}}"

					newIndexData := CollectData(substr, dataCenter, logging)
					if newIndexData.Message != "Success" {
						continue
					}
					newReplacement, ok := newIndexData.Data.(string)
					if !ok {
						continue
					}
					replacedStr = replacedStr[:match[0]] + newReplacement + replacedStr[match[1]:]
				}
				fmt.Println("Replaced string:", replacedStr)
				trimed = replacedStr
			}
			if strings.Contains(trimed, ".") {
				split := strings.Split(trimed, ".")
				var vardata interface{}
				skipData := ""

				if len(split) > 1 {
					var skipElement = 0
					for m, n := range split {
						if m < skipElement {
							continue
						}
						if strings.Contains(n, "<") {
							for i := m; i <= len(split); i++ {
								if i == m {
									continue
								}
								n += "." + split[i]
								if strings.Contains(split[i], ">") {
									skipElement = i + 1
									break
								}
							}
						}
						if n == skipData {
							skipData = ""
							continue
						}
						if m == 0 {
							vardata = dataCenter[n]
						} else {
							dataByte, _ := jsonExt.Marshal(vardata)

							newVarData := map[string]interface{}{}
							jsonExt.Unmarshal(dataByte, &newVarData)
							if strings.ContainsAny(n, "[]") {
								indexSplit := strings.Split(n, "[")
								indexStr := strings.Trim(indexSplit[1], "]")
								if strings.ContainsAny(indexStr, "<>") {
									var ok = false
									addr := "${{" + strings.Trim(indexStr, "<>") + "}}"
									newIndexData := CollectData(addr, dataCenter, logging)
									if newIndexData.Message != "Success" {
										continue
									}
									indexStr, ok = newIndexData.Data.(string)
									if !ok {
										continue
									}
								}
								index, _ := strconv.Atoi(indexStr)
								var temp interface{}
								switch b := vardata.(type) {
								case map[string]interface{}:
									temp = b[indexSplit[0]]
								case map[interface{}]interface{}:
									temp = b[indexSplit[0]]
								case map[int]interface{}:
									mapIndex, _ := strconv.Atoi(indexSplit[0])
									temp = b[mapIndex]
								}
								if temp == nil {
									vardata = nil
									break
								}
								switch t := temp.(type) {
								case []interface{}:

									if indexStr == "all" {
										sliceAll := []interface{}{}
										for _, d := range t {
											if len(split)-1 > m {
												switch l := d.(type) {
												case map[string]interface{}:
													sliceAll = append(sliceAll, l[split[m+1]])
													skipData = split[m+1]
												case map[interface{}]interface{}:
													sliceAll = append(sliceAll, l[split[m+1]])
													skipData = split[m+1]
												case map[int]interface{}:
													mapIndex, _ := strconv.Atoi(split[m+1])
													sliceAll = append(sliceAll, l[mapIndex])
													skipData = split[m+1]
												case []interface{}:
													for _, d := range l {
														sliceAll = append(sliceAll, d)
													}
													skipData = split[m+1]
												}
											}
										}
										vardata = sliceAll
										continue
									}
									if index >= 0 && index < len(t) {
										vardata = t[index]
									} else {
										vardata = nil
									}
								case []string:
									if index >= 0 && index < len(t) {
										vardata = t[index]
									} else {
										vardata = nil
									}
								case []int:
									if index >= 0 && index < len(t) {
										vardata = t[index]
									} else {
										vardata = nil
									}
								}
							} else if newVarData[n] != nil {
								if n == "body" || n == "Body" {
									if reflect.TypeOf(newVarData[n]).Kind() == reflect.String {
										dataByte2, _ := base64.StdEncoding.DecodeString(newVarData[n].(string))
										jsonExt.Unmarshal(dataByte2, &vardata)
									} else if reflect.TypeOf(newVarData[n]).Kind() == reflect.Slice {
										if data, ok := newVarData[n].([]byte); ok {
											jsonExt.Unmarshal(data, &vardata)
										} else if dataSlice, okSlice := newVarData[n].([]interface{}); okSlice {
											vardata = dataSlice
										}
									} else {
										vardata = newVarData[n]
									}
								} else {
									vardata = newVarData[n]
								}
							} else {
								vardata = nil
							}
						}
					}
					dataRequest = vardata
				}
			} else {
				dataRequest = dataCenter[trimed]
			}
		}
	case reflect.Slice:
		var mergeData interface{}
		sliceAppend := []interface{}{}

		switch t := address.(type) {
		case []interface{}:
			for _, l := range t {
				if reflect.TypeOf(l).Kind() != reflect.String {
					dataSlice := CollectData(l, dataCenter, logging)
					if dataSlice.Message != "Success" {
						continue
					}
					sliceAppend = append(sliceAppend, dataSlice.Data)
				} else if strings.Contains(l.(string), "${{ibridge.Append}}") {
					addressAppend := strings.Replace(l.(string), "${{ibridge.Append}}", "", 1)
					address := strings.Trim(addressAppend, "()")

					dataSlice := CollectData(address, dataCenter, logging)
					if dataSlice.Message != "Success" {
						continue
					}
					if dataSlice.Data != nil {
						switch b := dataSlice.Data.(type) {
						case []interface{}:
							sliceAppend = append(sliceAppend, b...)
						default:
							sliceAppend = append(sliceAppend, b)
						}
					}
				} else if strings.Contains(l.(string), "${{ibridge.Helpers") {
					dataRequest = ibridgeHelpers(l.(string), dataCenter, logging)
				} else if !strings.Contains(l.(string), "${{") {
					sliceAppend = append(sliceAppend, l)
				} else {
					//Get Variable
					var variableCheck interface{}
					trimed := strings.Trim(l.(string), "${}")
					if strings.Contains(trimed, ".") {
						split := strings.Split(trimed, ".")
						var vardata interface{}
						if len(split) > 1 {
							var skipElement = 0
							for m, n := range split {
								if m < skipElement {
									continue
								}
								if strings.Contains(n, "<") {
									for i := m; i <= len(split); i++ {
										if i == m {
											continue
										}
										n += "." + split[i]
										if strings.Contains(split[i], ">") {
											skipElement = i + 1
											break
										}
									}
								}
								if m == 0 {
									vardata = dataCenter[n]
								} else {

									dataByte, _ := jsonExt.Marshal(vardata)

									newVarData := map[string]interface{}{}
									jsonExt.Unmarshal(dataByte, &newVarData)
									if strings.ContainsAny(n, "[]") {
										indexSplit := strings.Split(n, "[")
										indexStr := strings.Trim(indexSplit[1], "]")
										if strings.Contains(indexStr, "<>") {
											var ok = false
											addr := "${{" + strings.Trim(indexStr, "<>") + "}}"
											newIndexData := CollectData(addr, dataCenter, logging)
											if newIndexData.Message != "Success" {
												continue
											}
											indexStr, ok = newIndexData.Data.(string)
											if !ok {
												continue
											}
										}
										index, _ := strconv.Atoi(indexStr)
										var temp interface{}
										switch b := vardata.(type) {
										case map[string]interface{}:
											temp = b[indexSplit[0]]
										case map[interface{}]interface{}:
											temp = b[indexSplit[0]]
										case map[int]interface{}:
											mapIndex, _ := strconv.Atoi(indexSplit[0])
											temp = b[mapIndex]
										}
										if temp == nil {
											vardata = nil
											break
										}

										if indexStr == "all" {
											vardata = temp
											continue
										}

										switch t := temp.(type) {
										case []interface{}:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										case []string:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										case []int:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										}
									} else if newVarData[n] != nil {
										if n == "body" || n == "Body" {
											if reflect.TypeOf(newVarData[n]).Kind() == reflect.String {
												dataByte2, _ := base64.StdEncoding.DecodeString(newVarData[n].(string))
												jsonExt.Unmarshal(dataByte2, &vardata)
											} else if reflect.TypeOf(newVarData[n]).Kind() == reflect.Slice {
												if data, ok := newVarData[n].([]byte); ok {
													jsonExt.Unmarshal(data, &vardata)
												} else if dataSlice, okSlice := newVarData[n].([]interface{}); okSlice {
													vardata = dataSlice
												}
											} else {
												vardata = newVarData[n]
											}
										} else {
											vardata = newVarData[n]
										}
									} else {
										vardata = nil
									}
								}
							}
							variableCheck = vardata
						}
					} else {
						variableCheck = dataCenter[trimed]
					}

					if mergeData != nil {
						mergeData = MergeObject(mergeData, variableCheck)
					} else {
						mergeData = variableCheck
					}
				}
			}
		case []string:
			for _, l := range t {
				if reflect.TypeOf(l).Kind() != reflect.String {
					dataSlice := CollectData(l, dataCenter, logging)
					if dataSlice.Message != "Success" {
						continue
					}
					sliceAppend = append(sliceAppend, dataSlice.Data)
				} else if strings.Contains(l, "${{ibridge.Append}}") {
					addressAppend := strings.Replace(l, "${{ibridge.Append}}", "", 1)
					address := strings.Trim(addressAppend, "()")

					dataSlice := CollectData(address, dataCenter, logging)
					if dataSlice.Message != "Success" {
						continue
					}
					sliceAppend = append(sliceAppend, dataSlice.Data)
					mergeData = sliceAppend
				} else if strings.Contains(l, "${{ibridge.Helpers") {
					dataRequest = ibridgeHelpers(l, dataCenter, logging)
				} else if !strings.Contains(l, "${{") {
					sliceAppend = append(sliceAppend, l)
				} else {
					//Get Variable
					var variableCheck interface{}
					trimed := strings.Trim(l, "${}")
					if strings.Contains(trimed, ".") {
						split := strings.Split(trimed, ".")
						var vardata interface{}
						if len(split) > 1 {
							var skipElement = 0
							for m, n := range split {
								if m < skipElement {
									continue
								}
								if strings.Contains(n, "<") {
									for i := m; i <= len(split); i++ {
										if i == m {
											continue
										}
										n += "." + split[i]
										if strings.Contains(split[i], ">") {
											skipElement = i + 1
											break
										}
									}
								}
								if m == 0 {
									vardata = dataCenter[n]
								} else {

									dataByte, _ := jsonExt.Marshal(vardata)

									newVarData := map[string]interface{}{}
									jsonExt.Unmarshal(dataByte, &newVarData)
									if strings.ContainsAny(n, "[]") {
										indexSplit := strings.Split(n, "[")
										indexStr := strings.Trim(indexSplit[1], "]")
										if strings.Contains(indexStr, "<>") {
											var ok = false
											addr := "${{" + strings.Trim(indexStr, "<>") + "}}"
											newIndexData := CollectData(addr, dataCenter, logging)
											if newIndexData.Message != "Success" {
												continue
											}
											indexStr, ok = newIndexData.Data.(string)
											if !ok {
												continue
											}
										}
										index, _ := strconv.Atoi(indexStr)
										var temp interface{}
										switch b := vardata.(type) {
										case map[string]interface{}:
											temp = b[indexSplit[0]]
										case map[interface{}]interface{}:
											temp = b[indexSplit[0]]
										case map[int]interface{}:
											mapIndex, _ := strconv.Atoi(indexSplit[0])
											temp = b[mapIndex]
										}
										if temp == nil {
											vardata = nil
											break
										}
										if indexStr == "all" {
											vardata = temp
											continue
										}
										switch t := temp.(type) {
										case []interface{}:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										case []string:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										case []int:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										}
									} else if newVarData[n] != nil {
										if n == "body" || n == "Body" {
											if reflect.TypeOf(newVarData[n]).Kind() == reflect.String {
												dataByte2, _ := base64.StdEncoding.DecodeString(newVarData[n].(string))
												jsonExt.Unmarshal(dataByte2, &vardata)
											} else if reflect.TypeOf(newVarData[n]).Kind() == reflect.Slice {
												if data, ok := newVarData[n].([]byte); ok {
													jsonExt.Unmarshal(data, &vardata)
												} else if dataSlice, okSlice := newVarData[n].([]interface{}); okSlice {
													vardata = dataSlice
												}
											} else {
												vardata = newVarData[n]
											}
										} else {
											vardata = newVarData[n]
										}
									} else {
										vardata = nil
									}
								}
							}
							variableCheck = vardata
						}
					} else if strings.ContainsAny(trimed, "[]") {
						var vardata interface{}
						indexSplit := strings.Split(trimed, "[")
						indexStr := strings.Trim(indexSplit[1], "]")
						if strings.Contains(indexStr, "<>") {
							var ok = false
							addr := "${{" + strings.Trim(indexStr, "<>") + "}}"
							newIndexData := CollectData(addr, dataCenter, logging)
							if newIndexData.Message != "Success" {
								continue
							}
							indexStr, ok = newIndexData.Data.(string)
							if !ok {
								continue
							}
						}
						index, _ := strconv.Atoi(indexStr)
						temp := dataCenter[indexSplit[0]]
						if temp == nil {
							vardata = nil
							break
						}
						if indexStr == "all" {
							vardata = temp
							continue
						}
						switch t := temp.(type) {
						case []interface{}:
							if index >= 0 && index < len(t) {
								vardata = t[index]
							} else {
								vardata = nil
							}
						case []string:
							if index >= 0 && index < len(t) {
								vardata = t[index]
							} else {
								vardata = nil
							}
						case []int:
							if index >= 0 && index < len(t) {
								vardata = t[index]
							} else {
								vardata = nil
							}
						}
						variableCheck = vardata
					} else {
						variableCheck = dataCenter[trimed]
						if dataCenter[trimed] == nil {
							variableCheck = trimed
						}
					}

					if mergeData != nil {
						mergeData = MergeObject(mergeData, variableCheck)

					} else {
						mergeData = variableCheck
					}
				}
			}
		}

		if len(sliceAppend) > 0 {
			dataRequest = sliceAppend
		} else {
			dataRequest = mergeData
		}
	case reflect.Map:
		mapData := map[interface{}]interface{}{}
		sliceAppend := []interface{}{}
		switch addr := address.(type) {
		case map[string]interface{}:

			for o, p := range addr {
				// Skip nil values to prevent panic
				if p == nil {
					continue
				}

				if reflect.TypeOf(p).Kind() == reflect.String && strings.Contains(p.(string), "string(") {
					if strings.Contains(o, "${{ibridge.Append}}") {
						sliceAppend = append(sliceAppend, strings.Trim(p.(string), "string()"))
						continue
					} else if strings.Contains(o, "${{ibridge.Helpers") {
						value := ibridgeHelpers(o, dataCenter, logging)
						if value != nil {
							mapData[o] = value
						}

					} else {
						value := strings.Trim(p.(string), "string()")
						if value != "" {
							mapData[o] = value
						}
					}
				} else if reflect.TypeOf(p).Kind() == reflect.String && strings.Contains(p.(string), ".(toString)") {
					toString := CollectData(strings.Trim(p.(string), ".(toString)"), dataCenter, logging)
					if toString.Message != "Success" {
						continue
					}
					str := fmt.Sprintf("%v", toString.Data)
					if strings.Contains(o, "${{ibridge.Append}}") {
						sliceAppend = append(sliceAppend, str)
						continue
					} else if strings.Contains(p.(string), "${{ibridge.Helpers") {
						value := ibridgeHelpers(p.(string), dataCenter, logging)
						if value != nil {
							mapData[o] = value
						}

					} else {
						if str != "" {
							mapData[o] = str
						}
					}
				} else if reflect.TypeOf(p).Kind() == reflect.String && strings.Contains(p.(string), "${{") && !strings.Contains(p.(string), "${{ibridge.") {
					trimed := strings.Trim(p.(string), "${}")
					if strings.Contains(trimed, ".") {
						split := strings.Split(trimed, ".")
						var vardata interface{}
						skipData := ""
						if len(split) > 1 {
							var skipElement = 0
							for m, n := range split {
								if m < skipElement {
									continue
								}
								if strings.Contains(n, "<") {
									for i := m; i <= len(split); i++ {
										if i == m {
											continue
										}
										n += "." + split[i]
										if strings.Contains(split[i], ">") {
											skipElement = i + 1
											break
										}
									}
								}
								if n == skipData {
									skipData = ""

									continue
								}
								if m == 0 {
									vardata = dataCenter[n]
								} else {

									dataByte, _ := jsonExt.Marshal(vardata)

									newVarData := map[string]interface{}{}
									jsonExt.Unmarshal(dataByte, &newVarData)
									if strings.ContainsAny(n, "[]") {
										indexSplit := strings.Split(n, "[")
										indexStr := strings.Trim(indexSplit[1], "]")
										if strings.ContainsAny(indexStr, "<>") {
											var ok = false
											addr := "${{" + strings.Trim(indexStr, "<>") + "}}"
											newIndexData := CollectData(addr, dataCenter, logging)
											if newIndexData.Message != "Success" {
												continue
											}
											indexStr, ok = newIndexData.Data.(string)
											if !ok {
												continue
											}
										}
										index, _ := strconv.Atoi(indexStr)
										var temp interface{}
										switch b := vardata.(type) {
										case map[string]interface{}:
											temp = b[indexSplit[0]]
										case map[interface{}]interface{}:
											temp = b[indexSplit[0]]
										case map[int]interface{}:
											mapIndex, _ := strconv.Atoi(indexSplit[0])
											temp = b[mapIndex]
										}
										if temp == nil {
											vardata = nil
											break
										}
										switch t := temp.(type) {
										case []interface{}:

											if indexStr == "all" {
												sliceAll := []interface{}{}
												for _, d := range t {
													if len(split)-1 > m {
														switch l := d.(type) {
														case map[string]interface{}:
															sliceAll = append(sliceAll, l[split[m+1]])
															skipData = split[m+1]
														case map[interface{}]interface{}:
															sliceAll = append(sliceAll, l[split[m+1]])
															skipData = split[m+1]
														case map[int]interface{}:
															mapIndex, _ := strconv.Atoi(split[m+1])
															sliceAll = append(sliceAll, l[mapIndex])
															skipData = split[m+1]
														}
													}
												}
												vardata = sliceAll
												continue
											}
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										case []string:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										case []int:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										}
									} else if newVarData[n] != nil {
										if n == "body" || n == "Body" {
											if reflect.TypeOf(newVarData[n]).Kind() == reflect.String {
												dataByte2, _ := base64.StdEncoding.DecodeString(newVarData[n].(string))
												jsonExt.Unmarshal(dataByte2, &vardata)
											} else if reflect.TypeOf(newVarData[n]).Kind() == reflect.Slice {
												if data, ok := newVarData[n].([]byte); ok {
													jsonExt.Unmarshal(data, &vardata)
												} else if dataSlice, okSlice := newVarData[n].([]interface{}); okSlice {
													vardata = dataSlice
												}
											} else {
												vardata = newVarData[n]
											}
										} else {
											vardata = newVarData[n]
										}
									} else {
										vardata = nil
									}
								}
							}
							if strings.Contains(o, "${{ibridge.Append}}") {
								sliceAppend = append(sliceAppend, vardata)
								continue
							} else if strings.Contains(o, "${{ibridge.Helpers") {
								dataRequest = ibridgeHelpers(o, dataCenter, logging)
							} else {
								if vardata != nil {
									mapData[o] = vardata
								}
							}
						}
					} else {
						if strings.Contains(o, "${{ibridge.Append}}") {
							sliceAppend = append(sliceAppend, dataCenter[trimed])
							continue
						} else if strings.Contains(o, "${{ibridge.Helpers") {
							dataRequest = ibridgeHelpers(o, dataCenter, logging)
						} else {
							if dataCenter[trimed] != nil {
								mapData[o] = dataCenter[trimed]
							}
						}
					}
				} else if reflect.TypeOf(p).Kind() == reflect.Map || reflect.TypeOf(p).Kind() == reflect.Slice {
					temp := CollectData(p, dataCenter, logging)
					if temp.Message != "Success" {
						continue
					}

					if strings.Contains(o, "${{ibridge.Append}}") {
						if temp.Data != nil {
							switch reflect.TypeOf(temp.Data).Kind() {
							case reflect.Slice:
								dataRequest = temp.Data
							case reflect.Map:
								dataRequest = temp.Data
							case reflect.String:
								appendSlice := []string{}
								appendSlice = append(appendSlice, temp.Data.(string))
								dataRequest = appendSlice
							case reflect.Int:
								appendSlice := []int{}
								appendSlice = append(appendSlice, temp.Data.(int))
								dataRequest = appendSlice
							}
						}
						continue
					} else if strings.Contains(o, "${{ibridge.Helpers") {
						dataRequest = ibridgeHelpers(o, dataCenter, logging)
					} else {
						if temp.Data != nil {
							mapData[o] = temp.Data
						}
					}
				} else {
					if strings.Contains(o, "${{ibridge.Append}}") {
						sliceAppend = append(sliceAppend, p)
						continue
					} else if reflect.TypeOf(p).Kind() == reflect.String && strings.Contains(p.(string), "${{ibridge.") {
						value := ibridgeHelpers(p.(string), dataCenter, logging)
						if value != nil {
							mapData[o] = ibridgeHelpers(p.(string), dataCenter, logging)
						}
					} else {
						if p != nil {
							mapData[o] = p
						}
					}
				}
			}
		case map[interface{}]interface{}:
			for o, p := range addr {
				// Skip nil values to prevent panic
				if p == nil {
					continue
				}

				if reflect.TypeOf(p).Kind() == reflect.String && strings.Contains(p.(string), "string(") {
					if reflect.TypeOf(o).Kind() == reflect.String && strings.Contains(o.(string), "${{ibridge.Append}}") {
						sliceAppend = append(sliceAppend, strings.Trim(p.(string), "string()"))
						continue
					} else if strings.Contains(p.(string), "${{ibridge.Helpers") {
						if value := ibridgeHelpers(p.(string), dataCenter, logging); value != nil {
							mapData[o] = value
						}
					} else {
						if value := strings.Trim(p.(string), "string()"); value != "" {
							mapData[o] = value
						}
					}
				} else if reflect.TypeOf(p).Kind() == reflect.String && strings.Contains(p.(string), "${{") && !strings.Contains(p.(string), "${{ibridge.") {
					trimed := strings.Trim(p.(string), "${}")
					if strings.Contains(trimed, ".") {
						split := strings.Split(trimed, ".")
						var vardata interface{}
						if len(split) > 1 {
							var skipElement = 0
							for m, n := range split {
								if m < skipElement {
									continue
								}
								if strings.Contains(n, "<") {
									for i := m; i <= len(split); i++ {
										if i == m {
											continue
										}
										n += "." + split[i]
										if strings.Contains(split[i], ">") {
											skipElement = i + 1
											break
										}
									}
								}
								if m == 0 {
									vardata = dataCenter[n]
								} else {

									dataByte, _ := jsonExt.Marshal(vardata)

									newVarData := map[string]interface{}{}
									jsonExt.Unmarshal(dataByte, &newVarData)
									if strings.ContainsAny(n, "[]") {
										indexSplit := strings.Split(n, "[")
										indexStr := strings.Trim(indexSplit[1], "]")
										if strings.Contains(indexStr, "<>") {
											var ok = false
											addr := "${{" + strings.Trim(indexStr, "<>") + "}}"
											newIndexData := CollectData(addr, dataCenter, logging)
											if newIndexData.Message != "Success" {
												continue
											}
											indexStr, ok = newIndexData.Data.(string)
											if !ok {
												continue
											}
										}
										index, _ := strconv.Atoi(indexStr)
										var temp interface{}
										switch b := vardata.(type) {
										case map[string]interface{}:
											temp = b[indexSplit[0]]
										case map[interface{}]interface{}:
											temp = b[indexSplit[0]]
										case map[int]interface{}:
											mapIndex, _ := strconv.Atoi(indexSplit[0])
											temp = b[mapIndex]
										}
										if temp == nil {
											vardata = nil
											break
										}
										if indexStr == "all" {
											vardata = temp
											continue
										}
										switch t := temp.(type) {
										case []interface{}:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										case []string:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										case []int:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										}
									} else if newVarData[n] != nil {
										if n == "body" || n == "Body" {
											if reflect.TypeOf(newVarData[n]).Kind() == reflect.String {
												dataByte2, _ := base64.StdEncoding.DecodeString(newVarData[n].(string))
												jsonExt.Unmarshal(dataByte2, &vardata)
											} else if reflect.TypeOf(newVarData[n]).Kind() == reflect.Slice {
												if data, ok := newVarData[n].([]byte); ok {
													jsonExt.Unmarshal(data, &vardata)
												} else if dataSlice, okSlice := newVarData[n].([]interface{}); okSlice {
													vardata = dataSlice
												}
											} else {
												vardata = newVarData[n]
											}
											vardata = newVarData[n]
										}
									} else {
										vardata = nil
									}
								}
							}
							if reflect.TypeOf(o).Kind() == reflect.String && strings.Contains(o.(string), "${{ibridge.Append}}") {
								sliceAppend = append(sliceAppend, vardata)
								continue
							} else if strings.Contains(o.(string), "${{ibridge.Helpers") {
								dataRequest = ibridgeHelpers(o.(string), dataCenter, logging)
							} else {
								if vardata != nil {
									mapData[o] = vardata
								}
							}
						}
					} else {
						if reflect.TypeOf(o).Kind() == reflect.String && strings.Contains(o.(string), "${{ibridge.Append}}") {
							sliceAppend = append(sliceAppend, dataCenter[trimed])
							continue
						} else if strings.Contains(o.(string), "${{ibridge.Helpers") {
							dataRequest = ibridgeHelpers(o.(string), dataCenter, logging)
						} else {
							if dataCenter[trimed] != nil {
								mapData[o] = dataCenter[trimed]
							}
						}
					}
				} else if reflect.TypeOf(p).Kind() == reflect.Map {
					temp := CollectData(p, dataCenter, logging)
					if temp.Message != "Success" {
						continue
					}
					if reflect.TypeOf(o).Kind() == reflect.String && strings.Contains(o.(string), "${{ibridge.Append}}") {
						sliceAppend = append(sliceAppend, temp.Data)
						continue
					} else if strings.Contains(p.(string), "${{ibridge.Helpers") {
						if value := ibridgeHelpers(p.(string), dataCenter, logging); value != nil {
							mapData[o] = value
						}
					} else {
						if temp.Data != nil {
							mapData[o] = temp.Data
						}
					}
				} else {
					if reflect.TypeOf(o).Kind() == reflect.String && strings.Contains(o.(string), "${{ibridge.Append}}") {
						sliceAppend = append(sliceAppend, p)
						continue
					} else if strings.Contains(p.(string), "${{ibridge.Helpers") {
						if value := ibridgeHelpers(p.(string), dataCenter, logging); value != nil {
							mapData[o] = value
						}
					} else {
						if p != nil {
							mapData[o] = p
						}
					}
				}
			}
		case map[int]interface{}:
			for o, p := range addr {
				// Skip nil values to prevent panic
				if p == nil {
					continue
				}

				if reflect.TypeOf(p).Kind() == reflect.String && strings.Contains(p.(string), "${{") && !strings.Contains(p.(string), "${{ibridge.") {
					trimed := strings.Trim(p.(string), "${}")
					if strings.Contains(trimed, ".") {
						split := strings.Split(trimed, ".")
						var vardata interface{}
						if len(split) > 1 {
							var skipElement = 0
							for m, n := range split {
								if m < skipElement {
									continue
								}
								if strings.Contains(n, "<") {
									for i := m; i <= len(split); i++ {
										if i == m {
											continue
										}
										n += "." + split[i]
										if strings.Contains(split[i], ">") {
											skipElement = i + 1
											break
										}
									}
								}
								if m == 0 {
									vardata = dataCenter[n]
								} else {

									dataByte, e := jsonExt.Marshal(vardata)
									if e != nil {
										res.Message = "Failed : Json Marshal " + e.Error()
										res.ResponseCode = "E1"
										return res
									}
									newVarData := map[string]interface{}{}
									jsonExt.Unmarshal(dataByte, &newVarData)
									if strings.ContainsAny(n, "[]") {
										indexSplit := strings.Split(n, "[")
										indexStr := strings.Trim(indexSplit[1], "]")
										if strings.Contains(indexStr, "<>") {
											var ok = false
											addr := "${{" + strings.Trim(indexStr, "<>") + "}}"
											newIndexData := CollectData(addr, dataCenter, logging)
											if newIndexData.Message != "Success" {
												continue
											}
											indexStr, ok = newIndexData.Data.(string)
											if !ok {
												continue
											}
										}
										index, _ := strconv.Atoi(indexStr)
										var temp interface{}
										switch b := vardata.(type) {
										case map[string]interface{}:
											temp = b[indexSplit[0]]
										case map[interface{}]interface{}:
											temp = b[indexSplit[0]]
										case map[int]interface{}:
											mapIndex, _ := strconv.Atoi(indexSplit[0])
											temp = b[mapIndex]
										}
										if temp == nil {
											vardata = nil
											break
										}
										if indexStr == "all" {
											vardata = temp
											continue
										}
										switch t := temp.(type) {
										case []interface{}:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										case []string:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										case []int:
											if index >= 0 && index < len(t) {
												vardata = t[index]
											} else {
												vardata = nil
											}
										}
									} else if newVarData[n] != nil {
										if n == "body" || n == "Body" {
											if reflect.TypeOf(newVarData[n]).Kind() == reflect.String {
												dataByte2, _ := base64.StdEncoding.DecodeString(newVarData[n].(string))
												jsonExt.Unmarshal(dataByte2, &vardata)
											} else if reflect.TypeOf(newVarData[n]).Kind() == reflect.Slice {
												if data, ok := newVarData[n].([]byte); ok {
													jsonExt.Unmarshal(data, &vardata)
												} else if dataSlice, okSlice := newVarData[n].([]interface{}); okSlice {
													vardata = dataSlice
												}
											} else {
												vardata = newVarData[n]
											}
										} else {
											vardata = newVarData[n]
										}
									} else {
										res.Message = "Failed : Nil data address " + strings.Join(split[:m], ".")
										res.ResponseCode = "E2"
										return res
									}
								}
							}
							if vardata != nil {
								mapData[o] = vardata
							}
						}
					} else {
						if dataCenter[trimed] != nil {
							mapData[o] = dataCenter[trimed]
						}
					}
				} else if reflect.TypeOf(p).Kind() == reflect.Map {
					temp := CollectData(p, dataCenter, logging)
					if temp.Data != nil {
						mapData[o] = temp.Data
					}
				} else {
					if p != nil {
						mapData[o] = p
					}
				}
			}
		}
		if len(sliceAppend) > 0 {
			dataRequest = sliceAppend
		} else if len(mapData) > 0 {
			dataRequest = mapData
		}
	}

	res.Data = dataRequest
	// if res.Data == nil {
	// 	res.Data = ""
	// }
	res.Success = true
	return res
}

func parameterSeparator(parameter, separator string) []string {
	internalVariableMark := "$"
	var separatedParameter []string
	startIndex := 0
	lastIndex := 0
	for i := 0; i < len(parameter); i++ {
		currentChar := fmt.Sprintf("%c", parameter[i])
		if i != len(parameter)-1 && currentChar == internalVariableMark {
			var curlyParenthesis []string
			for i += 1; i < len(parameter); i++ { // var i start from i + 1 because the cursor needs to move from `$` to `{`
				if fmt.Sprintf("%c", parameter[i]) == "{" {
					curlyParenthesis = append(curlyParenthesis, "{")
				} else if fmt.Sprintf("%c", parameter[i]) == "}" {
					popped := curlyParenthesis[len(curlyParenthesis)-1]
					curlyParenthesis = curlyParenthesis[:len(curlyParenthesis)-1]
					if popped != "{" {
						// if curly parenthesis not balance return separatedParameter immediately
						return separatedParameter
					}
				}
				if len(curlyParenthesis) == 0 {
					break
				}
			}
			var parenthesis []string
			if i != len(parameter)-1 && fmt.Sprintf("%c", parameter[i+1]) == "(" {
				for i += 1; i < len(parameter); i++ { // var i start from i + 1 because the cursor needs to move from `}` to `(`
					if fmt.Sprintf("%c", parameter[i]) == "(" {
						parenthesis = append(parenthesis, "(")
					} else if fmt.Sprintf("%c", parameter[i]) == ")" {
						popped := parenthesis[len(parenthesis)-1]
						parenthesis = parenthesis[:len(parenthesis)-1]
						if popped != "(" {
							// if parenthesis not balance return separatedParameter immediately
							return separatedParameter
						}
					}
					if len(parenthesis) == 0 {
						break
					}
				}
			}
		} else if currentChar == separator {
			lastIndex = i
			separatedParameter = append(separatedParameter, parameter[startIndex:lastIndex])
			startIndex = i + 1
		}
	}
	separatedParameter = append(separatedParameter, parameter[startIndex:])
	return separatedParameter
}

func ibridgeHelpers(address string, dataCenter map[string]interface{}, logging ceslogger.Logger) (resp interface{}) {
	var helpers, parameterArr []string
	var res, function, parameter string
	if strings.Contains(address, "(") {
		helpers = strings.Split(address, "(")
		function = helpers[0]
		parameter = strings.TrimSuffix(strings.Join(helpers[1:], "("), ")")
		parameterArr = parameterSeparator(parameter, ",")
		if len(parameterArr) > 1 {
			// If result from split by `,${{` is more than 1, it means the parameter have parameter more than 1
			for ip, p := range parameterArr {
				if strings.Contains(p, "${{") {
					dataSave := CollectData(p, dataCenter, logging)
					if dataSave.Message != "Success" {
						logging.LogError("ibridgeHelpers p", p, LogJSON(dataCenter))
						parameterArr[ip] = ""
					}
					collectData := dataSave.Data
					if collectData != nil {
						if reflect.TypeOf(collectData).Kind() == reflect.String {
							parameterArr[ip] = collectData.(string)
						} else if reflect.TypeOf(collectData).Kind() == reflect.Slice {
							parameterByte, _ := jsonExt.Marshal(collectData)
							parameterArr[ip] = string(parameterByte)
						} else if reflect.TypeOf(collectData).Kind() == reflect.Map {
							parameterByte, _ := jsonExt.Marshal(collectData)
							parameterArr[ip] = string(parameterByte)
						} else if reflect.TypeOf(collectData).Kind() == reflect.Float32 || reflect.TypeOf(collectData).Kind() == reflect.Float64 {
							parameterArr[ip] = strings.ReplaceAll(fmt.Sprintf("%.2f", collectData), ".00", "")
						} else if collectData == nil {
							parameterArr[ip] = ""
						} else {
							parameterArr[ip] = fmt.Sprintf("%v", collectData)
						}
					} else {
						parameterArr[ip] = ""
					}
				}
			}
		} else {
			if strings.Contains(parameter, "${{") {
				dataSave := CollectData(parameter, dataCenter, logging)
				if dataSave.Message != "Success" {
					logging.LogError("ibridgeHelpers parameter", parameter, LogJSON(dataCenter))
					parameter = ""
				}
				collectData := dataSave.Data
				if collectData != nil {
					if reflect.TypeOf(collectData).Kind() == reflect.String {
						parameter = collectData.(string)
					} else if reflect.TypeOf(collectData).Kind() == reflect.Slice {
						parameterByte, _ := jsonExt.Marshal(collectData)
						parameter = string(parameterByte)
					} else if reflect.TypeOf(collectData).Kind() == reflect.Float32 || reflect.TypeOf(collectData).Kind() == reflect.Float64 {
						parameter = strings.ReplaceAll(fmt.Sprintf("%.2f", collectData), ".00", "")
					} else {
						parameter = fmt.Sprintf("%v", collectData)
					}
				} else {
					parameter = ""
				}
			}
		}
	} else {
		function = address
	}
	switch function {
	case orchestrationConfig.Helpers.GenerateUUIDV4:
		res = UUIDgenerator()
	case orchestrationConfig.Helpers.GenerateUUIDV6:
		res = UUIDgenerator()
	case orchestrationConfig.Helpers.GenerateNumber:
		length, _ := strconv.Atoi(parameter)
		res = GenerateNumber(length)
	case orchestrationConfig.Helpers.GenerateHexadecimal:
		length, _ := strconv.Atoi(parameter)
		res = GenerateHexadecimal(length)
	case orchestrationConfig.Helpers.GenerateString:
		length, _ := strconv.Atoi(parameter)
		res = GenerateString(length)
	case orchestrationConfig.Helpers.GenerateReferenceId:
		// index 0 : printf-like format <string>
		// index 1 : length <integer>
		// index 2 : semicolon-separated: number;lower;upper
		format := parameterArr[0]
		length, _ := strconv.Atoi(parameterArr[1])
		typeParam := strings.ToLower(parameterArr[2])

		if !strings.Contains(format, "%s") {
			res = "<FAILED FORMAT> no string format"
			break
		}
		if strings.Index(format, "%s") != strings.LastIndex(format, "%s") {
			res = "<FAILED FORMAT> multiple string format"
			break
		}

		refId := "<FAILED GEN> timeout"
		for counter := 0; counter < 3; counter++ {
			genRefId := GenerateReferenceId(
				length,
				strings.Contains(typeParam, "number"),
				strings.Contains(typeParam, "lower"),
				strings.Contains(typeParam, "upper"),
			)
			if genRefId == "" {
				refId = "<FAILED GEN> empty"
				break
			}

			tempRefId := fmt.Sprintf(format, genRefId)

			// pastikan reference ID belum digunakan dalam waktu dekat
			_, err := redis.ReadCache(tempRefId, redisModuleReferenceId)
			if err != nil && !errors.Is(err, goredis.Nil) {
				refId = "<FAILED READ CACHE> " + err.Error()
				break
			}

			if errors.Is(err, goredis.Nil) {
				// refId tidak ditemukan di cache, berarti value aman digunakan
				err = redis.WriteCacheWith(tempRefId, redisModuleReferenceId, "x", redis.WithTimeToLive(24*time.Hour))
				if err != nil {
					refId = "<FAILED WRITE CACHE> " + err.Error()
					break
				}
				refId = tempRefId
				break
			}
		}
		res = refId

	case orchestrationConfig.Helpers.LocalTimestamp:
		res = time.Now().Local().String()
	case orchestrationConfig.Helpers.LocalTimestampFormat:
		res = time.Now().Local().Format(parameter)
	case orchestrationConfig.Helpers.UTCTimestamp:
		res = time.Now().UTC().String()
	case orchestrationConfig.Helpers.UTCTimestampFormat:
		res = time.Now().UTC().Format(parameter)
	case orchestrationConfig.Helpers.SubtractDateTime:
		// index 0 : what data to be returned. `years`, `months`, `weeks`, `days`, `hours`, `minutes`, `seconds`. Default: `days`
		// index 1 : dateTime format. Default: 2006-01-02
		// index 2 : dateTime 1
		// index 3 : dateTime 2
		// For every success subtraction or parsed format, this functions always return absolute int value.
		// If error occurs, return -1
		res = "-1"
		dataReturn := "days"
		dateTimeFormat := "2006-01-02"
		var dateTime1, dateTime2 time.Time
		var err error
		if len(parameterArr) >= 1 {
			dataReturn = parameterArr[0]
		}
		if len(parameterArr) >= 2 {
			dateTimeFormat = parameterArr[1]
		}
		if len(parameterArr) >= 3 {
			dateTime1Str := parameterArr[2]
			dateTime1, err = time.Parse(dateTimeFormat, dateTime1Str)
			if err != nil {
				return res
			}
		}
		if len(parameterArr) >= 4 {
			dateTime2Str := parameterArr[3]
			dateTime2, err = time.Parse(dateTimeFormat, dateTime2Str)
			if err != nil {
				return res
			}
		}
		difference := dateTime1.Sub(dateTime2)

		switch strings.ToLower(dataReturn) {
		case "years", "year":
			res = fmt.Sprintf("%d", int64(math.Abs(difference.Hours()/24/365)))
		case "months", "month":
			res = fmt.Sprintf("%d", int64(math.Abs(difference.Hours()/24/30)))
		case "weeks", "week":
			res = fmt.Sprintf("%d", int64(math.Abs(difference.Hours()/24/7)))
		case "hours", "hour":
			res = fmt.Sprintf("%.2f", math.Abs(difference.Hours()))
		case "minutes", "minute":
			res = fmt.Sprintf("%.2f", math.Abs(difference.Minutes()))
		case "seconds", "second":
			res = fmt.Sprintf("%.2f", math.Abs(difference.Seconds()))
		default:
			res = fmt.Sprintf("%d", int64(math.Abs(difference.Hours()/24)))
		}
	case orchestrationConfig.Helpers.AnyToString:
		res = parameter
	case orchestrationConfig.Helpers.URLToBase64:
		res = urlToImageBase64(parameter)
	case orchestrationConfig.Helpers.UnixTimestamp:
		res = fmt.Sprintf("%d", time.Now().Local().Unix())
	case orchestrationConfig.Helpers.UnixTimestampMilli:
		res = fmt.Sprintf("%d", time.Now().Local().UnixMilli())
	case orchestrationConfig.Helpers.UnixTimestampMicro:
		res = fmt.Sprintf("%d", time.Now().Local().UnixMicro())
	case orchestrationConfig.Helpers.UnixTimestampNano:
		res = fmt.Sprintf("%d", time.Now().Local().UnixNano())
	case orchestrationConfig.Helpers.StringConcat:
		for _, p := range parameterArr {
			res += p
		}
	case orchestrationConfig.Helpers.ConvertGenderTm:
		if parameter == "LAKI-LAKI" {
			res = "CUSTOMER_GENDER_MALE"
		} else if parameter == "PEREMPUAN" {
			res = "CUSTOMER_GENDER_FEMALE"
		} else {
			res = "CUSTOMER_GENDER_UNKNOWN"
		}
	case orchestrationConfig.Helpers.StringTrimLength:
		// index 0 : the string to be trimmed, default: string
		// index 1 : max length, default: 6
		// index 2 : from (left/right), default: left
		res = "string"
		maxLength := 6
		leftRight := "left"
		if len(parameterArr) >= 1 {
			res = parameterArr[0]
		}
		if len(parameterArr) >= 2 {
			maxLength, _ = strconv.Atoi(parameterArr[1])
		}
		if len(parameterArr) >= 3 {
			leftRight = parameterArr[2]
		}
		if leftRight == "right" {
			res = stringTrimmer(res, maxLength, false)
		} else {
			res = stringTrimmer(res, maxLength, true)
		}
	case orchestrationConfig.Helpers.StringPaddingZero:
		// index 0 : the string to be padded, default: empty
		// index 1 : max length, default: 10
		// index 2 : left/leading, right/trailing, default: left
		res = ""
		maxLength := 10
		leftRight := "left"
		if len(parameterArr) >= 1 {
			res = parameterArr[0]
		}
		if len(parameterArr) >= 2 {
			maxLength, _ = strconv.Atoi(parameterArr[1])
		}
		if len(parameterArr) >= 3 {
			leftRight = parameterArr[2]
		}
		switch leftRight {
		case "right", "trailing":
			for len(res) < maxLength {
				res += "0"
			}
		default:
			res = fmt.Sprintf("%0*s", maxLength, res)
		}
	case orchestrationConfig.Helpers.StringPaddingSpace:
		// index 0 : the string to be padded, default: empty
		// index 1 : max length, default: 10
		// index 2 : left/leading, right/trailing, default: left
		res = ""
		maxLength := 10
		leftRight := "left"
		if len(parameterArr) >= 1 {
			res = parameterArr[0]
		}
		if len(parameterArr) >= 2 {
			maxLength, _ = strconv.Atoi(parameterArr[1])
		}
		if len(parameterArr) >= 3 {
			leftRight = parameterArr[2]
		}
		switch leftRight {
		case "right", "trailing":
			res = fmt.Sprintf("%-*s", maxLength, res)
		default:
			res = fmt.Sprintf("%*s", maxLength, res)
		}
	case orchestrationConfig.Helpers.SetDefaultString:
		// index 0 : the string to be checked
		// index 1 : default value
		res = ""
		defaultValue := "default"
		if len(parameterArr) >= 1 {
			res = parameterArr[0]
		}
		if len(parameterArr) >= 2 {
			defaultValue = parameterArr[1]
		}
		if len(res) <= 0 {
			res = defaultValue
		}
	case orchestrationConfig.Helpers.DecodeStringTag:
		// index 0 : the string to be checked
		// index 1 : default value
		tag := map[string]interface{}{}
		for len(parameter) > 0 {
			tagName := parameter[:2]
			length, e := strconv.Atoi(parameter[2:4])
			if e != nil {
				break
			}
			value := ""
			init := length + 4

			if length < 4 {
				value = parameter[4:]
			} else {
				value = parameter[4:init]
			}

			tag[tagName] = value
			parameter = parameter[init:]
		}
		resp = tag
	case orchestrationConfig.Helpers.Sum:
		var sum float64
		for _, p := range parameterArr {
			if s, err := strconv.ParseFloat(p, 64); err == nil {
				sum += s
			}
		}
		resp = sum
	case orchestrationConfig.Helpers.TrimSpace:
		resp = strings.TrimSpace(parameter)
	case orchestrationConfig.Helpers.TrimPrefix:
		// index 1: text to be trimmed
		// index 2: prefix

		resp = strings.TrimPrefix(parameterArr[0], parameterArr[1])
	case orchestrationConfig.Helpers.JoinAndTrimLines:
		// Used to join lines and trim them
		// Example:
		// input -> []string{
		// 	"STRUK INI ADALAH BUKTI",
		// 	"PEMBAYARAN YANG SAH",
		// 	"",
		// 	"TERIMA KASIH TELAH MELAKUKAN   ",
		// 	"TOPUP                          ",
		// 	"EMAIL: CS@OVO.ID               ",
		// 	"CALL CENTER: 1500 696          ",
		// 	"                               ",
		// 	"                               ",
		// 	"                               ",
		// 	"                               ",
		// 	"                               ",
		// }
		// output -> "STRUK INI ADALAH BUKTI\nPEMBAYARAN YANG SAH\n\nTERIMA KASIH TELAH MELAKUKAN   \nTOPUP                          \nEMAIL: CS@OVO.ID          \nCALL CENTER: 1500 696"

		// Start from the end and trim trailing lines that are empty or all spaces
		end := len(parameterArr)
		for end > 0 && strings.TrimSpace(parameterArr[end-1]) == "" {
			end--
		}

		// Only use the non-empty tail
		cleaned := parameterArr[:end]

		// Join just once
		resp = strings.Join(cleaned, "\n")
	case orchestrationConfig.Helpers.AppendArray:
		// index 1: old array, or new array. If new array, pass `new` argument. Default: `new`
		// index 2: element to be added
		oldArrayStr := "new"
		var oldArr []interface{}
		var newElements []string
		if len(parameterArr) >= 1 {
			oldArrayStr = parameterArr[0]
		}
		if len(parameterArr) >= 2 {
			for _, e := range parameterArr[1:] {
				newElements = append(newElements, e)
			}
		}
		// check old array or initialization new array
		if oldArrayStr == "new" || len(oldArrayStr) < 1 {
			oldArr = []interface{}{}
		} else {
			if len(oldArrayStr) >= 2 && oldArrayStr[:1] == "[" && oldArrayStr[len(oldArrayStr)-1:] == "]" {
				jsonExt.Unmarshal([]byte(oldArrayStr), &oldArr)
			} else {
				return ""
			}
		}
		for _, elementStr := range newElements {
			if len(elementStr) >= 2 && elementStr[:1] == "[" && elementStr[len(elementStr)-1:] == "]" {
				// check if this string of array
				var temp []interface{}
				jsonExt.Unmarshal([]byte(elementStr), &temp)
				oldArr = append(oldArr, temp...)
			} else if len(elementStr) >= 2 && elementStr[:1] == "{" && elementStr[len(elementStr)-1:] == "}" {
				// check if this a map
				var temp map[string]interface{}
				jsonExt.Unmarshal([]byte(elementStr), &temp)
				oldArr = append(oldArr, temp)
			} else {
				oldArr = append(oldArr, elementStr)
			}
		}
		resp = oldArr
	case orchestrationConfig.Helpers.SortArray:
		// index 1: array
		// index 2: element name to be sorted
		// index 3: asc or desc, default: asc
		order := "asc"
		if len(parameterArr) >= 3 {
			order = parameterArr[2]
		}

		var arr []interface{}
		json.Unmarshal([]byte(parameterArr[0]), &arr)
		elementToOrder := parameterArr[1]

		// Sort the data based on the "variableName"
		sort.Slice(arr, func(i, j int) bool {
			// Convert to map[string]interface{} if applicable
			item1, ok1 := arr[i].(map[string]interface{})
			item2, ok2 := arr[j].(map[string]interface{})
			if !ok1 || !ok2 {
				return false
			}

			val1, ok1 := item1[elementToOrder]
			val2, ok2 := item2[elementToOrder]
			if !ok1 || !ok2 {
				return false
			}

			// Compare values based on the "order"
			if order == "asc" {
				return fmt.Sprint(val1) < fmt.Sprint(val2)
			}
			return fmt.Sprint(val1) > fmt.Sprint(val2)
		})

		resp = arr
	case orchestrationConfig.Helpers.AddLuhnAlgorithm:
		resp = strings.TrimSpace(parameter)
		stack := []int{}

		for i := 0; i < len(parameter); i++ {
			parsed, _ := strconv.Atoi(parameter[i : i+1])
			if i%2 == 1 {
				stack = append(stack, parsed*2)
			} else {
				stack = append(stack, parsed)
			}
		}

		totalDigit := 0
		for _, v := range stack {
			if v > 9 {
				totalDigit += (v - 9)
			} else {
				totalDigit += v
			}
		}

		multiply := totalDigit * 9
		resp = parameter + fmt.Sprintf("%d", multiply%10)
	case orchestrationConfig.Helpers.RemoveStringElement:
		// index 0 : the string to be removed it's element
		// index 1 : length of element to be remioved, default: 0
		// index 2 : left, right, default: left
		res = ""
		maxLength := 0
		leftRight := "left"
		if len(parameterArr) >= 1 {
			res = parameterArr[0]
		}
		if len(parameterArr) >= 2 {
			maxLength, _ = strconv.Atoi(parameterArr[1])
		}
		if len(parameterArr) >= 3 {
			leftRight = parameterArr[2]
		}
		switch leftRight {
		case "right", "trailing":
			if !(maxLength >= len(res)) {
				res = res[:len(res)-maxLength]
			}
		default:
			if !(maxLength >= len(res)) {
				res = res[maxLength:]
			}
		}
	case orchestrationConfig.Helpers.StringTrimLeft:
		// index 0 : the string to be trimmed, default: string
		// index 1 : element to trim
		res = "string"
		element := " "
		if len(parameterArr) >= 1 {
			res = parameterArr[0]
		}
		if len(parameterArr) >= 2 {
			element = parameterArr[1]
		}
		res = strings.TrimLeft(res, element)
	case orchestrationConfig.Helpers.StringTrimRight:
		// index 0 : the string to be trimmed, default: string
		// index 1 : element to trim
		res = "string"
		element := " "
		if len(parameterArr) >= 1 {
			res = parameterArr[0]
		}
		if len(parameterArr) >= 2 {
			element = parameterArr[1]
		}
		res = strings.TrimRight(res, element)
	case orchestrationConfig.Helpers.GetImageBase64FromURL:
		fileByte, err := DownloadFile(parameter)
		if err != nil {
			logging.LogError("download file ", err)
			return ""
		}
		res = base64.RawStdEncoding.EncodeToString(fileByte)
	case orchestrationConfig.Helpers.StringReplaceAll:
		// index 0 : the string to be edited
		// index 1 : old string, default : "."
		// index 2 : new string, default : ""
		res = parameter
		if len(parameterArr) >= 3 {
			res = strings.ReplaceAll(parameterArr[0], parameterArr[1], parameterArr[2])
		} else if len(parameterArr) >= 2 {
			res = strings.ReplaceAll(parameterArr[0], parameterArr[1], "")
		} else if len(parameterArr) >= 1 {
			res = strings.ReplaceAll(res, ".", "")
		}
	case orchestrationConfig.Helpers.GetLength:
		// index 0 : data to be counted
		// index 1 : default length
		length := 0
		defaultValue := 0
		if len(parameterArr) >= 1 {
			paramArray := []interface{}{}
			json.Unmarshal([]byte(parameterArr[0]), &paramArray)
			length = len(paramArray)
		}
		if len(parameterArr) >= 2 {
			defaultValue, _ = strconv.Atoi(parameterArr[1])
		}
		if len(parameter) >= 1 {
			paramArray := []interface{}{}
			json.Unmarshal([]byte(parameter), &paramArray)
			length = len(paramArray)
		}
		resp = length
		if length <= 0 {
			resp = defaultValue
		}
	case orchestrationConfig.Helpers.Multiply:
		initialVal := 1.0
		factor := 1.0
		if len(parameterArr) >= 1 {
			initialVal, _ = strconv.ParseFloat(parameterArr[0], 64)
		}
		if len(parameterArr) >= 2 {
			factor, _ = strconv.ParseFloat(parameterArr[1], 64)
		}
		resp = initialVal * factor
		if len(parameterArr) >= 3 {
			resp = fmt.Sprintf("%.12f", resp)
		}
	case orchestrationConfig.Helpers.GetEnv:
		res = ""
		if len(parameterArr) >= 1 {
			res = os.Getenv(parameterArr[0])
		}
		if len(parameterArr) >= 2 {
			res = os.Getenv(parameterArr[0])
			if res == "" {
				res = parameterArr[1]
			}
		}
	case orchestrationConfig.Helpers.Encrypt:
		// index 0 : data to be encrypted
		// index 1 : public key
		// return 0 if error
		var dataToBeEncrypted interface{}
		var publicKey string
		if len(parameterArr) >= 1 {
			dataToBeEncrypted = parameterArr[0]
		}
		if len(parameterArr) >= 2 {
			publicKey = parameterArr[1]
			if publicKey == "" {
				return "0"
			}
		}

		encryptedData, err := encrypt(dataToBeEncrypted, publicKey)
		if err != nil {
			return "0"
		}
		resp = encryptedData
	case orchestrationConfig.Helpers.ConvertTimestampFormat:
		// index 0 : date time string or timestamp string, default: current time
		// index 1 : source format, default: time.RFC3339Nano
		// index 2 : target format, default: time.RFC3339Nano
		// index 3 : convert timezone to `local` or `utc`, if not set, this index will be ignored
		// index 4 : convert language to `id` or `en` (default)
		// return source timestamp if error
		defaultFormat := time.RFC3339Nano
		sourceTimestamp := time.Now().Local().Format(defaultFormat)
		sourceFormat := defaultFormat
		targetFormat := defaultFormat
		targetLanguage := "en"
		if len(parameterArr) >= 1 {
			if !(parameterArr[0] == "" || parameterArr[0] == " " || parameterArr[0] == "default") {
				sourceTimestamp = parameterArr[0]
			}
		}
		if len(parameterArr) >= 2 {
			if !(parameterArr[1] == "" || parameterArr[1] == " " || parameterArr[1] == "default") {
				sourceFormat = parameterArr[1]
			}
		}
		if len(parameterArr) >= 3 {
			if !(parameterArr[2] == "" || parameterArr[2] == " " || parameterArr[2] == "default") {
				targetFormat = parameterArr[2]
			}
		}
		if len(parameterArr) >= 5 {
			if !(parameterArr[4] == "" || parameterArr[4] == " " || parameterArr[4] == "default") {
				targetLanguage = parameterArr[4]
			}
		}
		datetimeObject, err := time.Parse(sourceFormat, sourceTimestamp)
		if err != nil {
			res = sourceTimestamp
			break
		}
		if len(parameterArr) >= 4 {
			if parameterArr[3] == "local" {
				datetimeObject = datetimeObject.Local()
			} else if parameterArr[3] == "utc" {
				datetimeObject = datetimeObject.UTC()
			}
		}

		formattedDate := datetimeObject.Format(targetFormat)
		if targetLanguage == "id" {
			// Regex for textual months
			re := regexp.MustCompile(`\b(Jan|January|Feb|February|Mar|March|Apr|April|May|Jun|June|Jul|July|Aug|August|Sep|September|Oct|October|Nov|November|Dec|December)\b`)
			match := re.FindStringSubmatch(formattedDate)
			if len(match) > 1 {
				if (len(match[1])) > 3 {
					for en, id := range utils.MonthENtoID {
						formattedDate = strings.ReplaceAll(formattedDate, en, id)
					}
				} else {
					for en, id := range utils.MoENtoID {
						formattedDate = strings.ReplaceAll(formattedDate, en, id)
					}
				}
			}
		}
		resp = formattedDate
	case orchestrationConfig.Helpers.ConvertToUnixTimestamp:
		// index 0 : date time string or timestamp string, default: current time
		// index 1 : source format, default: time.RFC3339Nano
		defaultFormat := time.RFC3339Nano
		sourceTimestamp := time.Now().Local().Format(defaultFormat)
		sourceFormat := defaultFormat
		if len(parameterArr) >= 1 {
			if !(parameterArr[0] == "" || parameterArr[0] == " " || parameterArr[0] == "default") {
				sourceTimestamp = parameterArr[0]
			}
		}
		if len(parameterArr) >= 2 {
			if !(parameterArr[1] == "" || parameterArr[1] == " " || parameterArr[1] == "default") {
				sourceFormat = parameterArr[1]
			}
		}

		datetimeObject, err := time.Parse(sourceFormat, sourceTimestamp)
		if err != nil {
			res = sourceTimestamp
			break
		}

		unixTimestamp := datetimeObject.Unix()
		resp = unixTimestamp
	case orchestrationConfig.Helpers.MathDivision:
		// index 0 : number 1, default: 1
		// index 1 : number 2, default: 1
		// index 2 : rounding and precision, default: 2, if 0 will be ignored
		// index 3 : convert result to str
		// return 1 if error
		num1 := 1.0
		num2 := 1.0
		precision := 2.0
		if len(parameterArr) >= 1 {
			num1, _ = strconv.ParseFloat(parameterArr[0], 64)
		}
		if len(parameterArr) >= 2 {
			num2, _ = strconv.ParseFloat(parameterArr[1], 64)
		}
		if len(parameterArr) >= 3 {
			precision, _ = strconv.ParseFloat(parameterArr[2], 64)
		}
		result := num1 / num2
		if precision > 0 {
			ratio := math.Pow(10, precision)
			result = math.Round(result*ratio) / ratio
		}
		resp = result
		if len(parameterArr) >= 4 {
			resp = fmt.Sprintf("%.12f", resp)
		}
	case orchestrationConfig.Helpers.MathIntDiv:
		// Melakukan pembagian, namun hanya mengeluarkan angka bulatnya (MathIntDiv(13,3) = 4)
		// index 0 : angka yang dibagi, default: 1
		// index 1 : angka pembagi, default: 1
		// return -1 if error
		num1 := 1.0
		num2 := 1.0
		if len(parameterArr) >= 1 {
			num1, _ = strconv.ParseFloat(parameterArr[0], 64)
		}
		if len(parameterArr) >= 2 {
			num2, _ = strconv.ParseFloat(parameterArr[1], 64)
		}
		if num2 == 0 {
			resp = -1
			break
		}

		result := int(math.Floor(num1 / num2))
		resp = result
	case orchestrationConfig.Helpers.MathIntMod:
		// Melakukan pembagian, dan mengeluarkan sisa pembagiannya (MathIntMod(13,3) = 1)
		// index 0 : angka yang dibagi, default: 1
		// index 1 : angka pembagi, default: 1
		// return -1 if error
		num1 := 1.0
		num2 := 1.0
		if len(parameterArr) >= 1 {
			num1, _ = strconv.ParseFloat(parameterArr[0], 64)
		}
		if len(parameterArr) >= 2 {
			num2, _ = strconv.ParseFloat(parameterArr[1], 64)
		}
		if num2 == 0 {
			resp = -1
			break
		}

		result := int(math.Floor(math.Mod(num1, num2)))
		resp = result
	case orchestrationConfig.Helpers.FloatRounding:
		// index 0 : number 1, default: 1.00
		// index 2 : precision number, default: 2
		// return 1 if error
		num1 := 1.0
		precision := 2.0
		if len(parameterArr) >= 1 {
			num1, _ = strconv.ParseFloat(parameterArr[0], 64)
		}
		if len(parameterArr) >= 2 {
			precision, _ = strconv.ParseFloat(parameterArr[1], 64)
		}
		if precision > 0 {
			ratio := math.Pow(10, precision)
			num1 = math.Round(num1*ratio) / ratio
		}
		resp = num1
	case orchestrationConfig.Helpers.AddTimestamp:
		// index 0 : what data to be add. `years`, `months`, `days`, `hours`, `minutes`, `seconds`. Default: `days`
		// index 1 : dateTime format. Default: 2006-01-02
		// index 2 : dateTime 1
		// index 3 : addition variable
		res = "-1"
		dataReturn := "days"
		dateTimeFormat := "2006-01-02"
		addition := 0
		var dateTime1 time.Time
		var err error
		if len(parameterArr) >= 1 {
			dataReturn = parameterArr[0]
		}
		if len(parameterArr) >= 2 {
			dateTimeFormat = parameterArr[1]
		}
		if len(parameterArr) >= 3 {
			dateTime1Str := parameterArr[2]
			dateTime1, err = time.Parse(dateTimeFormat, dateTime1Str)
			if err != nil {
				return res
			}
		}
		if len(parameterArr) >= 4 {
			addition, err = strconv.Atoi(parameterArr[3])
			if err != nil {
				return res
			}
		}

		switch strings.ToLower(dataReturn) {
		case "years", "year":
			res = dateTime1.AddDate(addition, 0, 0).Format(dateTimeFormat)
		case "months", "month":
			res = dateTime1.AddDate(0, addition, 0).Format(dateTimeFormat)
		case "hours", "hour":
			res = dateTime1.Add(time.Duration(addition) * time.Hour).Format(dateTimeFormat)
		case "minutes", "minute":
			res = dateTime1.Add(time.Duration(addition) * time.Minute).Format(dateTimeFormat)
		case "seconds", "second":
			res = dateTime1.Add(time.Duration(addition) * time.Second).Format(dateTimeFormat)
		default:
			res = dateTime1.AddDate(0, 0, addition).Format(dateTimeFormat)
		}
	case orchestrationConfig.Helpers.CurrencyFormatID:
		// Use to format String amount to have thousand separator with 2 float decimal
		// example 1 -> 1,00
		// example 1234 -> 1.234,00
		// example 1234.56 -> 1.234,56
		// example 1234.567 -> 1.234,57
		// example 1234.111 -> 1.234,11
		// parameter : string to format

		normalizedStr := strings.Replace(parameter, ",", ".", -1)
		amount, _ := strconv.ParseFloat(normalizedStr, 64)

		// Format to two decimal places
		str := fmt.Sprintf("%.2f", amount)

		// Split the string into whole and decimal parts
		parts := strings.Split(str, ".")
		whole := parts[0]
		decimal := parts[1]

		// Insert thousand separators
		var result strings.Builder
		n := len(whole)
		for i, c := range whole {
			if i > 0 && (n-i)%3 == 0 {
				result.WriteRune('.')
			}
			result.WriteRune(c)
		}

		resp = result.String() + "," + decimal
	case orchestrationConfig.Helpers.CensorAccountNumber:
		// Use to censor account number
		// example 8888071035 -> 8888****35
		if parameter == "" {
			return ""
		}

		visibleStart := 4                                            // First 4 digits
		visibleEnd := 2                                              // Last 2 digits
		middleLength := len(parameter) - (visibleStart + visibleEnd) // Length of the censored part

		// Create the censor the middle part
		censoredMiddle := ""
		for i := 0; i < middleLength; i++ {
			censoredMiddle += "*"
		}

		// Construct the censored account number
		resp = parameter[:visibleStart] + censoredMiddle + parameter[len(parameter)-visibleEnd:]
	case orchestrationConfig.Helpers.ToLower:
		// Use to lower text
		// example Kurban -> kurban
		if parameter == "" {
			return ""
		}
		resp = strings.ToLower(parameter)
	case orchestrationConfig.Helpers.ToTitle:
		// Use to convert a string into title case
		// example BATARA KRESNA -> Batara Kresna
		if parameter == "" {
			return ""
		}
		caser := cases.Title(language.English)
		resp = caser.String(parameter)
	case orchestrationConfig.Helpers.ComputeCRC32:
		// Use to encypt using CRC32
		if parameter == "" {
			return ""
		}

		table := crc32.MakeTable(crc32.IEEE)
		resp = crc32.Checksum([]byte(parameter), table)
	case orchestrationConfig.Helpers.CalculateAge:
		// index 0 : birth date in string "YYYY-MM-DD"

		resp = -1
		dateTimeFormat := "2006-01-02"
		var birthdate time.Time
		var err error

		if parameter == "" {
			return resp
		}

		birthdate, err = time.Parse(dateTimeFormat, parameter)
		if err != nil {
			return resp
		}

		now := time.Now()

		age := now.Year() - birthdate.Year()

		// If birthday hasn't happened yet this year, subtract one
		if now.Month() < birthdate.Month() ||
			(now.Month() == birthdate.Month() && now.Day() < birthdate.Day()) {
			age--
		}

		resp = age
	case orchestrationConfig.Helpers.CensorString:
		// Use to censor string with flexible parameters
		// index 0 : string to censor
		// index 1 : symbol to use, default: '*'
		// index 2 : start from which position, default: 4
		// index 3 : how many last digits not to censor, default: 2
		// example: ${{ibridge.Helpers.CensorString}}(8888071035) -> 8888****35
		// example: ${{ibridge.Helpers.CensorString}}(8888071035,#,2,3) -> 88#####035

		stringToCensor := ""
		censorSymbol := "*"
		visibleStart := 4
		visibleEnd := 2

		if len(parameterArr) >= 1 {
			stringToCensor = parameterArr[0]
		}
		if len(parameterArr) >= 2 && parameterArr[1] != "" {
			censorSymbol = parameterArr[1]
		}
		if len(parameterArr) >= 3 && parameterArr[2] != "" {
			visibleStart, _ = strconv.Atoi(parameterArr[2])
		}
		if len(parameterArr) >= 4 && parameterArr[3] != "" {
			visibleEnd, _ = strconv.Atoi(parameterArr[3])
		}

		// Handle single parameter case
		if len(parameterArr) == 0 && parameter != "" {
			stringToCensor = parameter
		}

		if stringToCensor == "" {
			resp = ""
			break
		}

		// Validate that the string is long enough
		if len(stringToCensor) <= (visibleStart + visibleEnd) {
			resp = stringToCensor
			break
		}

		middleLength := len(stringToCensor) - (visibleStart + visibleEnd)

		// Create the censored middle part
		censoredMiddle := ""
		for i := 0; i < middleLength; i++ {
			censoredMiddle += censorSymbol
		}

		// Construct the censored string
		resp = stringToCensor[:visibleStart] + censoredMiddle + stringToCensor[len(stringToCensor)-visibleEnd:]

	case orchestrationConfig.Helpers.SumArray:
		var sum float64
		listElement := []float64{}
		e := json.Unmarshal([]byte(parameter), &listElement)
		if e != nil {
			listElementStr := []string{}
			e = json.Unmarshal([]byte(parameter), &listElementStr)
			if e != nil {
				sum = 0.0
				break
			}
			for _, el := range listElementStr {
				if s, err := strconv.ParseFloat(el, 64); err == nil {
					sum += s
				}
			}
		}
		for _, el := range listElement {
			sum += el
		}
		resp = sum
	}

	if resp == nil {
		resp = res
	}
	return resp
}

func encrypt(data interface{}, key string) (interface{}, error) {
	logging := ceslogger.Logger{}

	var err error
	var dataByte []byte
	switch v := data.(type) {
	case string:
		dataByte = []byte(v)
	case map[string]interface{}:
		dataByte, err = json.Marshal(v)
		if err != nil {
			logging.LogWarn(" json.Marshal(metadata)", err)
			return nil, err
		}
	}

	senderPublic, senderPrivate, errGenKey := box.GenerateKey(rand.Reader)
	if errGenKey != nil {
		logging.LogWarn("box.GenerateKey", errGenKey)
		return nil, errGenKey
	}

	nonce := nacl.NewNonce()
	recPub, _ := base64.StdEncoding.DecodeString(key)
	var pubkey [32]byte

	copy(pubkey[:], recPub)
	var adrnonce [24]byte
	var adrkeyOut [32]byte
	adrnonce = *nonce
	adrkeyOut = *senderPublic
	encData := box.Seal(nil, dataByte, nonce, &pubkey, senderPrivate)

	encodedData := base64.StdEncoding.EncodeToString(encData)

	hashedData := map[string]interface{}{
		"key":   base64.StdEncoding.EncodeToString(adrkeyOut[:]),
		"data":  encodedData,
		"nonce": base64.StdEncoding.EncodeToString(adrnonce[:]),
	}
	return hashedData, nil
}

func stringTrimmer(str string, maxLength int, fromLeft bool) string {
	if len(str) > maxLength {
		if fromLeft {
			str = str[:maxLength]
		} else {
			str = str[len(str)-maxLength:]
		}
	}
	return str
}

func urlToImageBase64(url string) (imageBase64 string) {
	logging := ceslogger.NewLogger("")
	fileName := GenerateKey()
	err := downloadFile(url, fileName)
	if err != nil {
		os.Remove("./" + fileName)
		logging.LogError("error download file", url, err)
		return ""
	}
	bytesFileKtp, err := os.ReadFile("./" + fileName)
	if err != nil {
		os.Remove("./" + fileName)
		logging.LogError("error read file", fileName, err)
		return ""
	}
	os.Remove("./" + fileName)

	return base64.StdEncoding.EncodeToString(bytesFileKtp)
}

func downloadFile(URL, fileName string) error {
	logging := ceslogger.NewLogger("")
	//Get the response bytes from the url
	logging.LogDebug(URL)

	var response *http.Response
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	r, err := http.NewRequest(http.MethodGet, "https://xip-bucket-124495977855.s3.ap-southeast-3.amazonaws.com/ocr/liveness/8ffef5a2-c57b-4e01-989f-7d16dff0ccf0.jpg", nil)
	if err != nil {
		logging.LogError(err.Error())
		return nil
	}

	response, err = client.Do(r)

	//Handle Error
	if err != nil {
		logging.LogError(err.Error())
		return nil
	} else if response.StatusCode != 200 {
		logging.LogError(LogJSON(response))
		return errors.New("received non 200 response code")
	}

	defer response.Body.Close()

	//Create a empty file
	file, err := os.Create(fileName)
	if err != nil {
		logging.LogError("os.Create", err)
		return err
	}

	//Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		logging.LogError("io.Copy", err)
		return err
	}
	response.Body.Close()
	file.Close()
	return nil
}
