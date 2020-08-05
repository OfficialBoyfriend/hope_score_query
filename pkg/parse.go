package pkg

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"score_query_server/modules"
	"score_query_server/utils"
	"strconv"
	"strings"
	"time"
)

func ParseAndWriteScore(studentId string, pageNumData ...interface{}) error {

	// 获取数据
	urlStr := "http://182.135.187.250:99/setReportParams"
	params := make(url.Values)
	params.Add("LS_XH", studentId)
	params.Add("resultPage", "http://182.135.187.250:99/reportFiles/cj/cj_zwcjd.jsp?")
	if len(pageNumData) == 2 && pageNumData[0].(string) != "" && pageNumData[1].(uint) > 1 {
		urlStr = "http://182.135.187.250:99/reportFiles/cj/cj_zwcjd.jsp"
		params = make(url.Values)
		params.Add("reportParamsId", pageNumData[0].(string))
		params.Add("report1_currPage", fmt.Sprint(pageNumData[1].(uint)))
	}
	req, err := GenerateRequest("POST", urlStr, params.Encode())
	if err != nil {
		return err
	}
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	bodyStr := utils.ConvertToString(string(body), "gbk", "utf-8")
	if !strings.Contains(bodyStr, "学生成绩表") {
		return errors.New("服务器异常")
	}

	// 本地保存副本
	/*if err := ioutil.WriteFile(fmt.Sprintf("data/html/%v.html", studentId), body, 0644); err != nil {
		log.Println("保存副本失敗:",err)
	}*/

	// 获取请求标识
	reportParamsId := ""
	if len(pageNumData) == 2 && pageNumData[0].(string) != "" {
		reportParamsId = pageNumData[0].(string)
	} else {
		reportParamsIdTmp, err := resp.Request.Response.Location()
		if err != nil {
			return fmt.Errorf("获取请求标识失败: %v", err)
		}
		reportParamsId = reportParamsIdTmp.Query().Get("reportParamsId")
	}

	// 提取页码
	tmp := regexp.MustCompile(`页号([\d])/([\d])`)
	pageNum := tmp.FindStringSubmatch(bodyStr)
	oldPageNum, _ := strconv.ParseUint(pageNum[1], 10,64)
	nextPageNum, _ := strconv.ParseUint(pageNum[2], 10,64)

	// 获取身份信息
	studentInfo, err := ParseStudentInfo(bodyStr)
	if err != nil {
		return err
	}
	log.Println(studentInfo)

	// 班级信息
	class := new(modules.Class)
	if err := class.GetAutoCreate(map[string]interface{}{"name":studentInfo["班级"]}); err != nil {
		return err
	}

	// 学生信息
	id := new(modules.Id)
	if err := id.GetAutoCreate(map[string]interface{}{
		"group": studentId,
		//"class_id": class.ID,
	}); err != nil {
		return err
	}

	// 更新学生信息
	if err := id.Update(map[string]interface{}{
		"class_id": class.ID,
		"name": studentInfo["姓名"],
		"gender": studentInfo["性别"],
		"id_number": studentInfo["身份证号"],
		"nation": studentInfo["民族"],
		"hometown": studentInfo["籍贯"],
		"political_status": studentInfo["政治面貌"],
		"date_of_birth": studentInfo["出生日期"],
		"enrollment_date": studentInfo["入学日期"],
		"graduation_date": studentInfo["毕业日期"],
		"profession": studentInfo["专业"],
		"professional_direction": studentInfo["专业方向"],
		"department": studentInfo["系所"],
	}); err != nil {
		return err
	}

	// 获取成绩信息
	scores := make([]*modules.Score, 0)
	scoresMap, totalCreditStr, err := ParseScore(bodyStr)
	if err != nil {
		return fmt.Errorf("解析成绩是发生错误: %v", err)
	}
	for _, s := range scoresMap {
		// 课程信息
		course := new(modules.Course)
		if err := course.GetAutoCreate(map[string]interface{}{"name": s["课程名"]}); err != nil {
			return fmt.Errorf("查询课程信息时时发生错误: %v, %v", err, studentId)
		}
		// 成绩
		resultScore, err := strconv.ParseFloat(s["成绩"], 64)
		if err != nil {
			log.Println(bodyStr, studentId)
			return fmt.Errorf("转换成绩数据类型时发生错误: %v, %v", err, studentId)
		}
		// 学分
		credit, err := strconv.ParseFloat(s["学分"], 64)
		if err != nil {
			return fmt.Errorf("转换学分数据类型时发生错误: %v, %v", err, studentId)
		}
		scores = append(scores, &modules.Score{
			IdId:     id.ID,
			CourseId: course.ID,
			Result:   resultScore,
			ExaminationTime: s["考试时间"],
			Credit: credit,
			EduType: s["修读方式"],
		})
	}

	// 将成绩记录插入数据库
	for _, v := range scores {
		if err := v.GetAutoCreateUseStruct(); err != nil {
			return fmt.Errorf("创建成绩记录时发生错误: %v", err)
		}
	}

	// 更新学生信息
	totalCredit, err := strconv.ParseFloat(totalCreditStr, 64)
	if err != nil {
		return fmt.Errorf("转换总学分数据类型时发生错误: %v", err)
	}
	if err := id.Update(map[string]interface{}{
		"total_credits": totalCredit,
		"is_valid": true,
	}); err != nil {
		return err
	}

	// 检测是否有下一页
	if oldPageNum < nextPageNum && (oldPageNum+1) > 1 {
		return ParseAndWriteScore(studentId, reportParamsId, uint(oldPageNum+1))
	}

	return new(ParseSuccess)
}

func ParseStudentInfo(data string) (map[string]string, error) {

	// 提取信息
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	results := make([]string, 0)
	dom.Find("table[id=\"report1\"] tr[height=\"20\"] td").Each(func(i int, selection *goquery.Selection) {
		results = append(results, selection.Text())
	})

	dataName := []string{"姓名", "学号", "性别", "身份证号","民族", "籍贯", "政治面貌", "出生日期","班级", "入学日期", "毕业日期","专业", "专业方向","系所", "培养方案"}
	oldIndex := 0
	studentInfo := make(map[string]string)

	for _, n := range dataName {
		for i := oldIndex; i < len(results); i++ {
			if results[i] != n {
				continue
			}
			studentInfo[n] = results[i+1]
			oldIndex = i
			break
		}
		if _, ok := studentInfo[n]; !ok {
			switch n {
			case "民族":
				studentInfo[n] = results[9]
			case "班级":
				studentInfo[n] = results[16]
			case "专业":
				studentInfo[n] = results[21]
			case "系所":
				studentInfo[n] = results[24]
			default:
				studentInfo[n] = ""
			}
		}
	}

	if studentInfo["姓名"] == "" {
		return nil, fmt.Errorf("学号无效")
	}

	// 验证数据是否缺失
	for k,v:=range studentInfo {
		if k == "培养方案" || k == "毕业日期" || k == "籍贯" {
			continue
		}
		if v == "" {
			log.Println("缺少学生信息:",studentInfo)
			return nil, fmt.Errorf("缺少学生信息")
		}
	}

	return studentInfo, nil
}

func ParseScore(data string, time ...string) ([]map[string]string, string, error) {

	// 提取信息
	dom, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return nil, "", err
	}
	results := make([]string, 0)
	dom.Find("table[id=\"report1\"] tr[height=\"20\"] td").Each(func(i int, selection *goquery.Selection) {
		results = append(results, selection.Text())
	})

	oldIndex := 0
	scoreData := make([]map[string]string, 0)

	for i := 0; i < len(results); i++ {
		if strings.Contains(results[i], "培养方案") {
			oldIndex = i
			break
		}
	}

	// 获取成绩信息
	for i := oldIndex; i < len(results); i++ {
		if strings.Contains(results[i], "平均学分绩点") {
			oldIndex = i
			break
		}
		// 查询成绩标识
		if len(time) > 0 && time[0] != "" {
			if results[i] != time[0] {
				continue
			}
		}else{
			tmp, err := regexp.MatchString(`([\d]{8})`,results[i])
			if err != nil {
				return nil, "", err
			}
			if !tmp {
				continue
			}
		}
		// 检测课程名称
		if results[i-5] == "" {
			results[i-5] = "未知课程"
		}
		scoreData = append(scoreData, map[string]string {
			"课程名": results[i-5],
			"学分": results[i-4],
			"成绩": results[i-3],
			"修读方式":results[i-2],
			"课程属性":results[i-1],
			"考试时间":results[i],
		})
	}

	if len(scoreData) < 1 {
		return nil, "", errors.New("缺少成绩信息")
	}

	for _, v := range scoreData {
		switch {
		//case v["课程名"] == "":
		//	fallthrough
		case v["学分"] == "":
			fallthrough
		case v["成绩"] == "":
		//	fallthrough
		//case v["考试时间"] == "":
			return nil, "", errors.New("成绩缺少学分或成绩")
		}
	}

	totalCredit := results[oldIndex-1]
	if totalCredit == "" {
		return nil, "", errors.New("缺少总学分信息")
	}

	return scoreData, totalCredit, nil
}

func GenerateRequest(reqType, url, args string, sessionID ...string) (*http.Request, error) {

	// 构建登录请求
	req, err := http.NewRequest(reqType, url,
		strings.NewReader(args))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	if reqType == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("User-Agent", "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:78.0) Gecko/20100101 Firefox/78.0")
	req.Header.Set("Referer", "http://182.135.187.250:99/")
	req.Header.Set("Origin", "http://182.135.187.250:99")
	req.Header.Set("Accept-Language", "zh-CN")
	if len(sessionID) > 0 && sessionID[0] != "" {
		req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: sessionID[0]})
	}
	return req, nil
}