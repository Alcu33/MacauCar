package app

import (
	"log"
	"macaoapply-auto/internal/client"
	"macaoapply-auto/pkg/config"
	"math/rand"
	"time"
)

var quit = make(chan bool)
var running = false

func Quit() {
	if !running {
		log.Println("未启动, 无需退出")
		return
	}
	quit <- true
}

func Wait() {
	log.Println("等待随机6-10秒...")
	sec := rand.Intn(5) + 6
	time.Sleep(time.Duration(sec) * time.Second)
}

// 配置检查
func CheckConfig() bool {
	// 超级鹰
	if config.Config.CJYOption.Username == "" || config.Config.CJYOption.Password == "" || config.Config.CJYOption.SoftId == "" {
		log.Println("请先配置超级鹰")
		return false
	}
	// 用户
	if config.Config.UserOption.Username == "" || config.Config.UserOption.Password == "" {
		log.Println("请先配置账户信息")
		return false
	}
	// 预约
	if config.Config.AppointmentOption.PlateNumber == "" || config.Config.AppointmentOption.AppointmentDate == 0 {
		log.Println("请先配置预约信息")
		return false
	}
	return true
}

func BootStrap() {
	running = true
	defer func() {
		running = false
	}()
	// 配置检查
	if !CheckConfig() {
		return
	}
	applyInfo := config.Config.AppointmentOption
	log.Println("启动...")
	for {
		// 检查是否登录
		if client.IsLogin() {
			break
		}
		log.Println("未登录，正在登录...")
		client.Login()
		if client.IsLogin() {
			break
		}
		log.Println("登录失败")
		Wait()
	}
	log.Println("当前已登录")
	formInstance, err := getPassQualification(applyInfo.PlateNumber)
	if err != nil {
		log.Println("获取预约资格失败：" + err.Error())
		return
	}
	log.Println("获取预约资格成功" + formInstance.FormInstanceID)

	for {
		// 退出检测
		select {
		case <-quit:
			log.Println("退出")
			return
		default:
		}
		list, err := GetAppointmentDateList()
		if err != nil {
			log.Println("获取预约日期失败：" + err.Error())
			Wait()
			continue
		}
		actDate := time.Unix(applyInfo.AppointmentDate, 0).Format("2006-01-02")
		if !CheckAppointmentListHasAvailable(list, actDate) {
			log.Println("无可用预约")
			Wait()
			continue
		}
		log.Println("有可用预约，正在预约...")
		// 预约
		for {
			err = DoAppointment(applyInfo)
			if err != nil {
				log.Println("预约失败：" + err.Error())
				log.Println("等待30s...")
				time.Sleep(30 * time.Second)
				continue
			}
			log.Println("预约成功！预约进程即将退出...")
		}
	}
}