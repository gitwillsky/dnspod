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
	"regexp"
	"strconv"
	"strings"

	"github.com/gitwillsky/slimgo_com/config"
)

// 域名列表
type DomainsList struct {
	RStatus  Status     `json:"status"`
	RInfo    DomainInfo `json:"info"`
	RDomains []Domains  `json:"domains"`
}

// 返回状态信息
type Status struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Created_at string `json:"created_at"`
}

// 域名信息
type DomainInfo struct {
	TotalDomain int `json:"domain_total"`
	AllTotal    int `json:"all_total"`
}

// 域名字段
type Domains struct {
	ID      int    `json:"id"`
	Status  string `json:"status"`
	GroupID string `json:"group_id"`
	TTL     string `json:"ttl"`
	Name    string `json:"name"`
	Records string `json:"records"`
}

// 记录列表
type RecordList struct {
	Records     []Record     `json:"records"`
	RStatus     Status       `json:"status"`
	RrecordInfo RecordInfo   `json:"info"`
	Domain      RecordDomain `json:"domain"`
}

// 记录信息
type RecordInfo struct {
	SubDomains  string `json:"sub_domains"`
	RecordTotal string `json:"record_total"`
}

// 记录域名信息
type RecordDomain struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// 记录字段
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

// 更新的RecordList
type RecordUpdateList struct {
	DomainID   int
	RecordID   string
	RecordType string
	RecordLine string
	RecordMX   string
	SubDomain  string
}

// 用户相关信息
var (
	email    string // 邮箱
	password string // 密码
)

// 更新记录
func UpdateRecords(records []RecordUpdateList) error {
	var ip string
	// GET ip Value
	resp, e := http.Get("http://1111.ip138.com/ic.asp")
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
		if tempData, e := ioutil.ReadAll(resp.Body); e != nil {
			return e
		} else {
			// 解析JSON
			if err = json.Unmarshal(tempData, &dat); err != nil {
				return err
			}
		}

		fmt.Println(dat.RStatus.Message)
	}

	return nil
}

// 获取记录列表
func GetRecordList(client_ids []int) ([]RecordList, error) {
	var dat = make([]RecordList, len(client_ids))

	fmt.Println("Request domain record list...")
	for i := 0; i < len(client_ids); i++ {
		resp, err := getClient().PostForm("https://dnsapi.cn/Record.List", url.Values{
			"login_email":    {email},
			"login_password": {password},
			"format":         {"json"},
			"domain_id":      {strconv.Itoa(client_ids[i])},
		})
		if err != nil {
			return nil, err
		}

		// 读取返回字节流并解析
		if tempData, e := ioutil.ReadAll(resp.Body); e != nil {
			return nil, e
		} else {
			// 解析JSON
			if err = json.Unmarshal(tempData, &dat[i]); err != nil {
				return nil, err
			}
		}
	}

	return dat, nil
}

// 获取域名列表
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
	if tempData, e := ioutil.ReadAll(resp.Body); e != nil {
		return nil, e
	} else {
		// 解析JSON
		if err = json.Unmarshal(tempData, &dat); err != nil {
			return nil, err
		}
	}

	// 判断用户是否登录成功
	if strings.Contains(dat.RStatus.Message, "fail") {
		return nil, errors.New("Login failed, Code: " + dat.RStatus.Code)
	}

	return dat, nil
}

func main() {
	var (
		reader                 = bufio.NewReader(os.Stdin)
		selectDomains          = make([]int, 0)
		domainList             *DomainsList
		recordList             []RecordList
		err                    error
		i                      = 0
		tempList               = make([]RecordUpdateList, 0)
		updateList             = make([]RecordUpdateList, 0)
		domain_num, record_num string
	)
	if len(os.Args) == 2 && strings.TrimSpace(os.Args[1]) == "config" {
		// 打开配置文件
		Appconfig := config.New()
		if err = Appconfig.SetConfig("SlimGoConfig", "ddns.conf"); err != nil {
			panic("Open config file error: " + err.Error())
		}
		//  从配置文件读取信息
		email = Appconfig.GetString("email")
		password = Appconfig.GetString("password")
		domain_num = Appconfig.GetString("domains")
		record_num = Appconfig.GetString("records")
	} else {

		if len(os.Args) == 3 {
			email = strings.TrimSpace(os.Args[1])
			password = strings.TrimSpace(os.Args[2])
		} else {
			fmt.Println("Input Params wrong!")
			return
		}
	}

	// 获取域名列表
	if domainList, err = GetDomainList(); err != nil {
		fmt.Println(err.Error())
		return
	}

	// 显示部分域名列表信息
	fmt.Println("Result: " + domainList.RStatus.Message)
	fmt.Println("Total Domains: " + strconv.Itoa(domainList.RInfo.TotalDomain))
	fmt.Println("Domain ID              Domain Name")
	for _, val := range domainList.RDomains {
		fmt.Println(strconv.Itoa(val.ID) + "              " + val.Name)
	}

	// 选择要修改的域名
label1:
	if len(domain_num) == 0 {
		fmt.Printf("Select Domain:(1/1,2,3)?:")
		domain_num, _ = reader.ReadString('\n')
	}
	s := strings.Split(domain_num, ",")
	// 检查输入
	if len(s) > len(domainList.RDomains) {
		fmt.Println("input number too huge.")
		return
	}
	for _, v := range s {
		if i, err := strconv.Atoi(v); err != nil {
			fmt.Println("Input Format Error!")
			goto label1
		} else {
			selectDomains = append(selectDomains, domainList.RDomains[i-1].ID)
		}
	}

	// 获取选择域名的记录信息
	if recordList, err = GetRecordList(selectDomains); err != nil {
		fmt.Println(err.Error())
		return
	}

	// 显示域名记录信息
	fmt.Println("DomainID		RecordID		Record		Type		Value		TTL")
	for _, val := range recordList {
		for _, v := range val.Records {
			fmt.Printf("%d		%s		%s		%s		%s		%s\n",
				val.Domain.Id,
				v.ID,
				v.Name,
				v.Type,
				strings.TrimSpace(v.Value),
				v.TTL)

			record := RecordUpdateList{
				DomainID:   val.Domain.Id,
				RecordID:   v.ID,
				RecordType: v.Type,
				RecordLine: v.Line,
				RecordMX:   v.Mx,
				SubDomain:  v.Name,
			}
			tempList = append(tempList, record)
			i++
		}
	}

	// 选择要修改的记录
	if len(record_num) == 0 {
		fmt.Printf("Select Records(1/1,2,3):")
		record_num, _ = reader.ReadString('\n')
	}
	r := strings.Split(record_num, ",")
	// 检查输入
	if len(r) > i {
		fmt.Println("input number too huge.")
		return
	}

	for _, v := range r {
		num, _ := strconv.Atoi(v)
		updateList = append(updateList, tempList[num-1])
	}

	// update
	if err = UpdateRecords(updateList); err != nil {
		fmt.Println(err.Error())
	}
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
