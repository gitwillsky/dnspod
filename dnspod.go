package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gitwillsky/slimgo/config"
)

// DomainList域名列表
type DomainsList struct {
	RStatus  Status     `json:"status"`
	RInfo    DomainInfo `json:"info"`
	RDomains []Domains  `json:"domains"`
}

// Status 返回状态信息
type Status struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// DomainInfo 域名信息
type DomainInfo struct {
	TotalDomain int `json:"domain_total"`
	AllTotal    int `json:"all_total"`
}

// Domains 域名字段
type Domains struct {
	ID      int    `json:"id"`
	Status  string `json:"status"`
	GroupID string `json:"group_id"`
	TTL     string `json:"ttl"`
	Name    string `json:"name"`
	Records string `json:"records"`
}

// RecordList 记录列表
type RecordList struct {
	Records     []Record     `json:"records"`
	RStatus     Status       `json:"status"`
	RrecordInfo RecordInfo   `json:"info"`
	Domain      RecordDomain `json:"domain"`
}

// RecordInfo 记录信息
type RecordInfo struct {
	SubDomains  string `json:"sub_domains"`
	RecordTotal string `json:"record_total"`
}

// RecordDomain 记录域名信息
type RecordDomain struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Record 记录字段
type Record struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	TTL    string `json:"ttl"`
	Value  string `json:"value"`
	Status string `json:"status"`
	Line   string `json:"line"`
	Mx     string `json:"mx"`
}

// RecordUpdateList 更新的RecordList
type RecordUpdateList struct {
	DomainID   int
	RecordID   string
	RecordType string
	RecordLine string
	RecordMX   string
	SubDomain  string
	OldValue   string
}

// 用户相关信息
var (
	email    string // 邮箱
	password string // 密码
)

// UpdateRecords 更新域名记录
func UpdateRecords(records []RecordUpdateList) error {
	var ip string
	// GET ip Value
	resp, e := http.Get("http://ip.3322.org/")
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	result, er := ioutil.ReadAll(resp.Body)
	if er != nil {
		return er
	}
	reg := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)
	ip = reg.FindString(string(result))
	fmt.Println("New IP address: " + ip)
	dat := RecordList{}

	for _, record := range records {
		fmt.Printf("Update Record ID: %s  Domain ID : %d\n", record.RecordID, record.DomainID)
		resp, err := getClient().PostForm("https://dnsapi.cn/Record.Modify", url.Values{
			"login_email":    {email},
			"login_password": {password},
			"format":         {"json"},
			"domain_id":      {strconv.Itoa(record.DomainID)},
			"record_id":      {record.RecordID},
			"record_type":    {record.RecordType},
			"record_line":    {record.RecordLine},
			"value":          {ip},
			"sub_domain":     {record.SubDomain},
			"mx":             {record.RecordMX},
		})
		if err != nil {
			return err
		}

		// 读取返回字节流并解析
		tempData, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			return e
		}

		// 解析JSON
		if err = json.Unmarshal(tempData, &dat); err != nil {
			return err
		}

		fmt.Println(dat.RStatus.Message)
	}

	return nil
}

// GetRecordList 获取记录列表
func GetRecordList(clientIds []int) ([]RecordList, error) {
	var dat = make([]RecordList, len(clientIds))

	fmt.Println("Request domain record list...")
	for i := 0; i < len(clientIds); i++ {
		resp, err := getClient().PostForm("https://dnsapi.cn/Record.List", url.Values{
			"login_email":    {email},
			"login_password": {password},
			"format":         {"json"},
			"domain_id":      {strconv.Itoa(clientIds[i])},
		})
		if err != nil {
			return nil, err
		}

		// 读取返回字节流并解析
		tempData, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			return nil, e
		}

		// 解析JSON
		if err = json.Unmarshal(tempData, &dat[i]); err != nil {
			return nil, err
		}
	}

	return dat, nil
}

// GetDomainList 获取域名列表
func GetDomainList() (*DomainsList, error) {
	var dat = &DomainsList{}

	// 发送Post请求
	resp, err := getClient().PostForm("https://dnsapi.cn/Domain.List", url.Values{
		"login_email":    {email},
		"login_password": {password},
		"format":         {"json"},
	})

	// 请求是否成功
	if err != nil {
		return nil, err
	}

	// 读取返回字节流并解析
	tempData, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, e
	}
	// 解析JSON
	if err = json.Unmarshal(tempData, &dat); err != nil {
		return nil, err
	}

	// 判断用户是否登录成功
	if strings.Contains(dat.RStatus.Message, "fail") {
		return nil, errors.New("Login failed, Code: " + dat.RStatus.Code)
	}

	return dat, nil
}

// new https client
func getClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
		DisableCompression: true,
	}

	client := &http.Client{Transport: tr}

	return client
}

func main() {
	var (
		reader               = bufio.NewReader(os.Stdin)
		selectDomains        = make([]int, 0)
		domainList           *DomainsList
		recordList           []RecordList
		err                  error
		i                    = 0
		tempList             = make([]RecordUpdateList, 0)
		updateList           = make([]RecordUpdateList, 0)
		domainNum, recordNum string
		useConfig            bool
	)

	switch len(os.Args) {
	case 2:
		// use config
		if strings.TrimSpace(os.Args[1]) == "config" {
			// 打开配置文件
			Appconfig := config.New()
			dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			if err = Appconfig.SetConfig("SlimGoConfig", dir+"/ddns.conf"); err != nil {
				panic("Open config file error: " + err.Error())
			}
			//  从配置文件读取信息
			email = Appconfig.GetString("email")
			password = Appconfig.GetString("password")
			domainNum = Appconfig.GetString("domains")
			recordNum = Appconfig.GetString("records")
			useConfig = true
		} else {
			fmt.Println("Input params wrong")
			return
		}
	case 3:
		// use command
		email = strings.TrimSpace(os.Args[1])
		password = strings.TrimSpace(os.Args[2])
		useConfig = false
	default:
		fmt.Println("Input params wrong")
		return
	}

	// 获取域名列表
	if domainList, err = GetDomainList(); err != nil {
		fmt.Println(err.Error())
		return
	}

	if !useConfig {
		// 显示部分域名列表信息
		fmt.Println("Result: " + domainList.RStatus.Message)
		fmt.Println("Total Domains: " + strconv.Itoa(domainList.RInfo.TotalDomain))
		fmt.Println("Domain ID              Domain Name")
		for _, val := range domainList.RDomains {
			fmt.Println(strconv.Itoa(val.ID) + "              " + val.Name)
		}

		// 选择域名
		for {
			fmt.Printf("Select Domain:(1 / 1,2,3)?:")
			domainNum, _ = reader.ReadString('\n')
			if domainNum == "" {
				continue
			}

			s := strings.Split(domainNum, ",")
			if len(s) > len(domainList.RDomains) {
				fmt.Println("Too many domain number")
				continue
			}

			for _, v := range s {
				id, e := strconv.Atoi(strings.TrimSpace(v))
				if e != nil {
					fmt.Println("Domain number Format wrong, must be number")
					continue
				}
				selectDomains = append(selectDomains, domainList.RDomains[id-1].ID)
			}

			break
		}
	}

	if useConfig {
		if domainNum == "" {
			fmt.Println("Please check config domains value")
			return
		}
		s := strings.Split(domainNum, ",")
		if len(s) > len(domainList.RDomains) {
			fmt.Println("Too many domain number")
			return
		}

		for _, v := range s {
			id, e := strconv.Atoi(strings.TrimSpace(v))
			if e != nil {
				fmt.Println("Domain number Format wrong, must be number")
				return
			}
			selectDomains = append(selectDomains, domainList.RDomains[id-1].ID)
		}
	}

	// 获取选择域名的记录信息
	if recordList, err = GetRecordList(selectDomains); err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, val := range recordList {
		for _, v := range val.Records {
			record := RecordUpdateList{
				DomainID:   val.Domain.ID,
				RecordID:   v.ID,
				RecordType: v.Type,
				RecordLine: v.Line,
				RecordMX:   v.Mx,
				SubDomain:  v.Name,
				OldValue:   v.Value,
			}
			tempList = append(tempList, record)
			i++
		}
	}

	if !useConfig {
		// 显示域名记录信息
		fmt.Println("DomainID		RecordID		Record		Type		Line		Value")
		for _, v := range tempList {
			fmt.Printf("%d		%s		%s		%s		%s		%s\n",
				v.DomainID,
				v.RecordID,
				v.SubDomain,
				v.RecordType,
				v.RecordLine,
				v.OldValue,
			)
		}

		// 选择域名纪录
		for {
			fmt.Printf("Select Records(1/1,2,3):")
			recordNum, _ = reader.ReadString('\n')
			if recordNum == "" {
				continue
			}
			r := strings.Split(recordNum, ",")
			if len(r) > i {
				fmt.Println("Too many record numbers")
				continue
			}
			for _, v := range r {
				num, e := strconv.Atoi(strings.TrimSpace(v))
				if e != nil {
					fmt.Println("Records must be number format")
					continue
				}
				updateList = append(updateList, tempList[num-1])
			}
			break
		}
	}

	if useConfig {
		if recordNum == "" {
			fmt.Println("Please check your config records value")
			return
		}

		r := strings.Split(recordNum, ",")
		if len(r) > i {
			fmt.Println("Too many record numbers")
			return
		}

		for _, v := range r {
			num, e := strconv.Atoi(strings.TrimSpace(v))
			if e != nil {
				fmt.Println("Records must be number format")
				continue
			}
			updateList = append(updateList, tempList[num-1])
		}
	}

	// update
	if err = UpdateRecords(updateList); err != nil {
		fmt.Println(err.Error())
	}
}
