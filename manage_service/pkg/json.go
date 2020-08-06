package pkg

import (
	"encoding/json"
	"fmt"
	"log"
	"score_query_server/modules"
	"score_query_server/utils"
)

// 同步成绩到七牛
// 发生错误时自动跳过
func GeneratedScoreJson() error {

	id := new(modules.Id)
	// 查询未同步到七牛且是有效用户的成绩
	ids, err := id.GetFind(map[string]interface{}{"is_sync_ok": false, "is_valid": true}, true)
	if err != nil {
		return fmt.Errorf("查询未同步用户信息失败: %v", err)
	}

	taskNum := len(ids)

MainFor:
	for i, v := range ids {
		log.Println(i+1,"/",taskNum, "上传成绩:", v)
		scoreCourseName := make([]string, 0)
		// 查询成绩所属课程
		for _, s := range v.Scores {
			course := new(modules.Course)
			if err := course.Get(map[string]interface{}{"id":s.CourseId}); err != nil {
				log.Println("查询课程失败:", err)
				continue MainFor
			}
			scoreCourseName = append(scoreCourseName, course.Name)
		}
		result := map[string]interface{}{"id": v, "score_course_name": scoreCourseName}
		// 序列化数据
		jsonData, err := json.Marshal(result)
		if err != nil {
			log.Println("序列号数据失败:", err)
			continue MainFor
		}
		// 上传到七牛
		_, err = utils.UploadFile("xl-blog", fmt.Sprintf("score/%v.json", v.Group), jsonData)
		if err != nil {
			log.Println("上传文件失败:", err)
			continue MainFor
		}
		// 更新同步状态
		err = v.Update(map[string]interface{}{"is_sync_ok": true})
		if err != nil {
			log.Println("更新用户同步状态失败:", err)
			continue MainFor
		}
	}

	return nil
}

// 将课程前九的数据上传到七牛
func GeneratedScoreTopJson() error {

	// 查询出所有课程
	course := new(modules.Course)
	courses, err := course.GetFind(map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("同步前九数据出错, 查询课程失败: %v", err)
	}

	tmpCount := 0

	results := make(map[string]interface{})
	for _, c := range courses {

		if tmpCount > 20 {
			break
		}
		tmpCount++

		// 查询课程成绩前九
		score := new(modules.Score)
		scores, err := score.GetTop(9, map[string]interface{}{"course_id": c.ID})
		if err != nil {
			return fmt.Errorf("查询前九数据失败: %v", err)
		}

		// TODO: 临时添加
		if len(scores) < 9 {
			log.Println("课程成绩不足9条（或为空）:", c.Name)
			continue
		}

		idData := make([]*map[string]string, 0)
		for _, s :=range scores {
			id:=modules.NewId()
			if err := id.Get(map[string]interface{}{"id": s.IdId}); err != nil {
				log.Println("查询前九用户信息失败:", err)
				continue
			}
			class:=new(modules.Class)
			if err := class.Get(map[string]interface{}{"id": id.ClassId}); err != nil {
				log.Println("查询班级信息失败:",err)
				continue
			}
			idData = append(idData, &map[string]string{"name": id.Name, "class": class.Name})
		}

		results[c.Name] = map[string]interface{}{"ids": idData, "scores": scores}
	}
	// 序列化数据
	jsonData, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("序列号数据失败: %v", err)
	}
	// 上传到七牛
	_, err = utils.UploadFile("xl-blog", "score/course_top.json", jsonData)
	if err != nil {
		return fmt.Errorf("上传失败: %v", err)
	}

	return nil
}