/*
客户端功能：
	创建两个线程，每个线程先随机睡眠1-3S，然后生成一个0-2的随机数，
	并且分别以TCP协议或者UDP协议将数据传输给服务器
	重复上述操作10次
*/
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

//同步数据
var wgc = sync.WaitGroup{}

//睡眠函数
func sleepRandTime() uint16{
	rand.Seed(time.Now().UnixNano())
	sleepTime := rand.Intn(2) + 1
	time.Sleep((time.Second) * time.Duration(sleepTime))
	return uint16(sleepTime)
}

//玩家1通过tcp协议向服务器发送信息
func threadTcp() {
	//建立连接
	conn, err := net.Dial("tcp", "0.0.0.0:10086")
	if err != nil {
		fmt.Println("client dial err = ", err)
		return
	}

	//开始循环生成数据并且发送数据
	for i:=0;i<10;i++{
		//睡眠随机时间 1-3秒
		byteArrOne := bytes.NewBuffer([]byte{})
		sleepTime := sleepRandTime()
		binary.Write(byteArrOne,binary.LittleEndian,sleepTime)
		n1,err := conn.Write(byteArrOne.Bytes())
		if err != nil {
			fmt.Printf("tcp round %d 1 error",i)
		}

		//生成选择
		byteArrTwo := bytes.NewBuffer([]byte{})
		choice := rand.Int()%3
		binary.Write(byteArrTwo,binary.LittleEndian,uint16(choice))
		n2,err := conn.Write(byteArrTwo.Bytes())
		if err != nil {
			fmt.Printf("tcp round %d 1 error",i)
		}

		//输出提示信息
		fmt.Printf("tcp：round：%d  sleepTime：%d  choice：%d %d %d\n",i,sleepTime,choice,n1,n2)
	}
	conn.Close()
	wgc.Done()
}

//玩家2通过udp协议向服务器发送信息
func threadUdp() {
	//建立连接
	conn, err := net.Dial("udp", "127.0.0.1:10087")
	if err != nil{
		fmt.Printf("DialTcp error, os exit……")
		os.Exit(-1)
	}

	//开始循环生成数据并且发送数据
	for i:=0;i<10;i++{
		//睡眠随机时间 1-3秒
		byteArr := bytes.NewBuffer([]byte{})
		sleepTime := sleepRandTime()
		binary.Write(byteArr,binary.LittleEndian,sleepTime)
		n1,err := conn.Write(byteArr.Bytes())
		if err != nil{
			fmt.Printf("udp write data error……")
		}

		//生成选择
		byteArr = bytes.NewBuffer([]byte{})
		choice := rand.Int() % 3
		binary.Write(byteArr,binary.LittleEndian,uint16(choice))
		n2,err := conn.Write(byteArr.Bytes())
		if err != nil{
			fmt.Printf("udp write message error……")
		}
		fmt.Printf("udp：round：%d  sleepTime：%d  choice：%d %d %d \n",i,sleepTime,choice,n1,n2)
	}
	conn.Close()
	wgc.Done()
}

func main() {
	//创建两个玩家线程
	go threadTcp()
	go threadUdp()

	//等待玩家线程结束
	wgc.Add(2)
	wgc.Wait()
}
