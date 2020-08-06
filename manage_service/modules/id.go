package modules

import (
	"errors"
)

// Id 学生号信息
type Id struct {
	Model
	ClassId uint `json:"class_id"` // 班级ID
	Name string `json:"name"` // 名称
	Gender string `json:"gender"` // 性别
	IdNumber string  `json:"id_number"` // 身份证号
	Nation string `json:"nation"` // 民族
	Hometown string `json:"hometown"` // 籍贯
	PoliticalStatus string  `json:"political_status"` // 政治面貌
	DateOfBirth string `json:"date_of_birth"` // 出生日期
	EnrollmentDate string `json:"enrollment_date"` // 入学日期
	GraduationDate string `json:"graduation_date"` // 毕业日期
	Profession string `json:"profession"` // 专业
	ProfessionalDirection string `json:"professional_direction"` // 专业方向
	Department string `json:"department"` // 系所
	TotalCredits float64 `json:"total_credits"` // 总学分
	Year string `json:"year"` // 入学年份
	Other string `json:"other"` // 未知含义
	Number string `json:"number"` // 所在班级
	Num string `json:"num"` // 班级内编号
	Group string `json:"group"` // 学号
	IsValid bool `json:"is_valid"` //是否有效
	IsSyncOk bool `json:"is_sync_ok"` // 是否同步完成（同步到七牛对象存储）
	Scores []*Score `json:"scores"` // 成绩数据
}

func NewId() *Id {
	return new(Id)
}

func (s *Id) GetAutoCreate(where map[string]interface{}) error {
	return db.Where(where).FirstOrCreate(s).Error
}

func (s *Id) Get(where map[string]interface{}) error {
	return db.Where(where).First(s).Error
}

func (s *Id) GetUseStruct() error {
	return db.Where(s).First(s).Error
}

func (s *Id) GetFind(where map[string]interface{}, isPreload ...bool) ([]*Id, error) {
	var results []*Id
	if len(isPreload) > 0 && isPreload[0] {
		err := db.Where(where).Preload("Scores").Find(&results).Error
		return results, err
	}
	err := db.Where(where).Find(&results).Error
	return results, err
}

func (s *Id) Create() error {
	if !db.NewRecord(s) {
		return errors.New("检测到错误，主键已存在")
	}
	return db.Create(s).Error
}

func (s Id) Update(value map[string]interface{}) error {
	return db.Model(&s).Updates(value).Error
}

func (s Id) UpdateUseStruct() error {
	return db.Model(&s).Updates(s).Error
}

func (Id) Delete(where map[string]interface{}) error {
	return db.Where(where).Delete(new(Id)).Error
}